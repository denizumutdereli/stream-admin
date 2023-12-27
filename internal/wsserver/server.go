package wsserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

const (
	maxMessageSize = 1024 * 1024 // 1 MB
	pongWait       = 10 * time.Second
	writeWait      = 10 * time.Second
	clientTimeout  = 60 * time.Minute // production 1 hour
)

var (
	newline                = []byte{'\n'}
	space                  = []byte{' '}
	activeConnectionsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_active_connections",
			Help: "Current number of active websocket connections",
		})
	totalBridgesGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_active_bridges",
			Help: "Current number of active nats bridges",
		})
	totalWorkWeightGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_active_load",
			Help: "Current number of active users multiple with subscribed nats bridges",
		})
	upgrader = websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Client struct {
	conn       *websocket.Conn
	lastActive time.Time
	bridges    map[string]chan []byte
	writeCh    chan []byte
	closeCh    chan bool
	closeChans map[string]chan struct{}
	subs       map[string]*nats.Subscription
}

type Server struct {
	ctx                      context.Context
	cancel                   context.CancelFunc
	config                   *config.Config
	logger                   *zap.Logger
	channels                 []string
	assets                   []string
	redis                    *transport.RedisManager
	nats                     *transport.NatsManager
	allowedIPs               map[string]struct{}
	maxConnections           int
	maxExtremum              int
	maxMessageSize           int64
	activeConns              int32
	clients                  map[*Client]bool
	clientPool               *sync.Pool
	klients                  sync.Map
	requestChannel           chan SubsRequest
	prometheusConnectionChan chan bool
	prometheusBridgeChan     chan int
	removeClientChan         chan Client
	natsCleanupChan          chan UnsubscribeNatsTopics
	mu                       sync.Mutex
}

type Request struct {
	Action string `json:"action" validate:"required,oneof=subscribe unsubscribe"`
	Topics string `json:"topics" validate:"required"`
}

type SubsRequest struct {
	Client    *Client
	Action    string
	NatsTopic string
}

type UnsubscribeNatsTopics struct {
	subs map[string]*nats.Subscription
}

type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewServer(appContext *types.ExchangeConfig) *Server {

	ipMap := make(map[string]struct{})
	for _, ip := range appContext.Config.AllowedSpecificIps {
		ipMap[ip] = struct{}{}
	}

	server := &Server{
		config:         appContext.Config,
		logger:         appContext.Logger,
		redis:          appContext.Redis,
		nats:           appContext.Nats,
		channels:       appContext.Channels,
		assets:         appContext.StreamAssets,
		allowedIPs:     ipMap,
		maxConnections: appContext.Config.MaxConnections,
		maxMessageSize: 1024 * 1024, // 1 MB
		maxExtremum:    appContext.Config.MaxConnections * 2,
		activeConns:    0,
		clients:        make(map[*Client]bool),
		clientPool: &sync.Pool{
			New: func() interface{} {
				return &Client{}
			},
		},
		requestChannel:           make(chan SubsRequest, 1000),
		prometheusConnectionChan: make(chan bool, 100),
		prometheusBridgeChan:     make(chan int, 100),
		removeClientChan:         make(chan Client, 1000),
		natsCleanupChan:          make(chan UnsubscribeNatsTopics, 1000),
	}

	ctx, cancel := context.WithCancel(context.Background())
	server.ctx = ctx
	server.cancel = cancel

	go func() {
		server.handleSubRequest(ctx)
	}()

	go server.prometheusTotalConnections(ctx)
	go server.prometheusTotalBridges(ctx)

	go server.unsubscribeFromNats(ctx)

	go server.removeClientPermanently(ctx)

	return server
}

func (s *Server) Serve(port string) {
	http.HandleFunc("/", utils.EnableCORS(s.handleConnections, s.config.CorsWhitelist))
	http.HandleFunc("/heartbeat", s.handleHeartbeat)
	server := &http.Server{Addr: ":" + port}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			s.logger.Info("Websocket Server stopped", zap.Error(err))
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)
	s.logger.Info("Websocket Server gracefully stopped")
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer func() {
		// client, ok := s.klients.Load(ws)
		// connection := client.(*Client)
		// if ok {
		// 	select {
		// 	case s.removeClientChan <- *connection:
		// 	default:
		// 		s.logger.Debug("There is removeClientChan signal problem...")
		// 	}

		// }

		s.prometheusConnectionChan <- false
	}()

	ws.SetReadLimit(maxMessageSize)
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Prometheus
	if atomic.LoadInt32(&s.activeConns) >= int32(s.maxConnections) {
		http.Error(w, "Too many connections", http.StatusServiceUnavailable)
		return
	}

	clientA := s.clientPool.Get().(*Client)
	defer func() {
		clientA = &Client{}
	}()

	clientA.conn = ws
	clientA.lastActive = time.Now()
	clientA.bridges = make(map[string]chan []byte)
	clientA.writeCh = make(chan []byte)
	clientA.closeCh = make(chan bool)
	clientA.closeChans = make(map[string]chan struct{})
	clientA.subs = make(map[string]*nats.Subscription)

	s.klients.Store(clientA.conn, clientA)
	defer s.clientPool.Put(clientA)

	s.prometheusConnectionChan <- true

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.writeToClient(clientA)
	}()

	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Warn("Unexpected closing client side error")
				return //TODO: add here cleanup logic
			}
			break
		}
		p = bytes.TrimSpace(bytes.Replace(p, newline, space, -1))
		var req Request
		err = json.Unmarshal(p, &req)
		if err != nil {

			http.Error(w, "Invalid request format.", http.StatusBadRequest)
			s.logger.Error("Invalid format received", zap.Error(err))
			s.sendError(clientA.conn, http.StatusBadRequest, err.Error())
			clientA.conn.Close()
			return

		}
		wg.Add(1)
		go func(req Request) {
			defer wg.Done()
			s.handleRequest(clientA, req)
		}(req)
	}
	wg.Wait()

}

func (s *Server) writeToClient(client *Client) {
	for {
		select {
		case message, ok := <-client.writeCh:
			//c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				//s.logger.Error("Error writing message to client", zap.Error(err))
				s.removeClientChan <- *client
				return
			}

		case <-client.closeCh:
			return
		}
	}
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if s.IsConnected() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR"))
	}
}

func (s *Server) IsConnected() bool {
	return true // TODO: live logic
}

func (s *Server) validateRequest(req Request) bool {
	action := strings.ToUpper(req.Action)
	return action == "SUBSCRIBE" || action == "UNSUBSCRIBE"
}

func (s *Server) sendError(conn *websocket.Conn, errorCode int, errorMessage string) {

	errorResponse := ErrorMessage{
		Code:    errorCode,
		Message: errorMessage,
	}

	jsonResponse, err := json.Marshal(errorResponse)
	if err != nil {
		s.logger.Error("Error marshaling error response", zap.Error(err))
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, jsonResponse)
	if err != nil {
		s.logger.Error("Error sending error message", zap.Error(err))
	}

	// if !s.hasActiveSubscriptions(conn) {
	// 	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, errorMessage))
	// 	if err != nil {
	// 		s.logger.Error("Error sending error closure", zap.Error(err))
	// 	}
	// }
}

func (s *Server) handleSubRequest(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for req := range s.requestChannel {

		select {
		case <-ctx.Done():
			return

		default:

			natsTopic := req.NatsTopic
			client := req.Client
			if req.Action == "SUBSCRIBE" {
				if _, ok := client.bridges[natsTopic]; !ok {
					ch := make(chan []byte)
					closeCh := make(chan struct{})
					client.bridges[natsTopic] = ch
					client.closeChans[natsTopic] = closeCh

					var sub *nats.Subscription
					messageHandler := func(m *nats.Msg) {
						select {
						case <-client.closeCh:
							s.logger.Debug("connection closed, ignoring this time...")
						default:
							ch <- m.Data
						}

					}
					sub, _ = s.nats.Subscribe(natsTopic, messageHandler)

					client.subs[natsTopic] = sub
					s.prometheusBridgeChan <- 1

					go func() {
						for msg := range ch {
							select {
							case <-client.closeCh:
								s.logger.Debug("connection closed, ignoring this time...")
								return
							default:
								client.writeCh <- msg
							}
						}
					}()

				} else {
					s.logger.Debug("Subscribed earlier", zap.String("channel", natsTopic))
				}
			} else if req.Action == "UNSUBSCRIBE" {
				if closeChan, exists := client.closeChans[natsTopic]; exists {
					close(closeChan)
					delete(client.closeChans, natsTopic)
				}
				if ch, ok := client.bridges[natsTopic]; ok {
					if sub, exists := client.subs[natsTopic]; exists {
						sub.Unsubscribe()
						delete(client.subs, natsTopic)
					}
					close(ch)
					delete(client.bridges, natsTopic)
				} else {
					s.logger.Debug("Not subscribed earlier", zap.String("channel", natsTopic))
				}

				s.prometheusBridgeChan <- -1

			}
		}

	}

}

func (s *Server) unsubscribeFromNats(ctx context.Context) {

	for subscription := range s.natsCleanupChan {

		select {
		case <-ctx.Done():
			return
		default:
			for topic, sub := range subscription.subs {
				err := sub.Unsubscribe()
				//s.logger.Debug("unsubscribing from NATS topics", zap.String("channel", topic))
				if err != nil {
					s.logger.Error("nats topic unsubscription error", zap.String("channel", topic), zap.Error(err))
				}
				s.prometheusBridgeChan <- -1
			}
		}

	}

}

func (s *Server) removeClientPermanently(ctx context.Context) {
	for client := range s.removeClientChan {
		client, ok := s.klients.Load(client.conn)
		connection := client.(*Client)

		select {
		case <-ctx.Done():
			return
		default:

			go func() {
				if ok {
					s.natsCleanupChan <- UnsubscribeNatsTopics{subs: connection.subs}
					for topic, _ := range connection.subs {
						s.mu.Lock()
						delete(connection.subs, topic)
						s.mu.Unlock()
					}

					s.prometheusBridgeChan <- -len(connection.subs)

					for _, ch := range connection.bridges {
						close(ch)
					}

					s.mu.Lock()
					s.klients.Delete(connection.conn)
					s.mu.Unlock()

					s.mu.Lock()
					if connection.closeCh != nil {
						close(connection.closeCh)
						connection.closeCh = nil
					}
					if connection.writeCh != nil {
						close(connection.writeCh)
						connection.writeCh = nil
					}
					s.mu.Unlock()

				} else {
					s.logger.Debug("connection of the client could not find")
				}
			}()

		}

	}
}

func (s *Server) handleRequest(client *Client, req Request) {
	if !s.validateRequest(req) {
		return
	}
	req.Action = strings.ToUpper(req.Action)
	topics := strings.Split(req.Topics, ",")
	for _, topic := range topics {
		topic = strings.TrimSpace(topic)

		parts := strings.Split(topic, "@")
		if len(parts) != 2 {
			s.sendError(client.conn, http.StatusBadRequest, fmt.Sprintf("Invalid topic format: %s", topic))
			return
		}
		topicName, asset := strings.ToLower(parts[0]), strings.ToUpper(parts[1])

		if topicName == "markets" {
			if asset != "DATA" && asset != "SNAPSHOT" {
				s.sendError(client.conn, http.StatusBadRequest, fmt.Sprintf("Invalid channel or topic for markets data: %s@%s", topicName, asset))
				return
			}
		} else if topicName == "all" && s.isValidTopic(asset) {
			if topicName == "all" {
				for _, serverChannel := range s.channels {
					natsTopic := fmt.Sprintf("%s.%s", serverChannel, asset)
					if serverChannel == "markets" {
						continue
					}
					if req.Action == "SUBSCRIBE" {
						s.requestChannel <- SubsRequest{Client: client, Action: "SUBSCRIBE", NatsTopic: natsTopic}

					} else if req.Action == "UNSUBSCRIBE" {
						s.requestChannel <- SubsRequest{Client: client, Action: "UNSUBSCRIBE", NatsTopic: natsTopic}
					}

				}
			}
		} else if !s.isValidChannel(topicName) || !s.isValidTopic(asset) {
			s.sendError(client.conn, http.StatusBadRequest, fmt.Sprintf("Invalid channel or topic name: %s@%s", topicName, asset))
			return
		}

		natsTopic := fmt.Sprintf("%s.%s", topicName, asset)

		if req.Action == "SUBSCRIBE" {
			s.requestChannel <- SubsRequest{Client: client, Action: "SUBSCRIBE", NatsTopic: natsTopic}

		} else if req.Action == "UNSUBSCRIBE" {
			s.requestChannel <- SubsRequest{Client: client, Action: "UNSUBSCRIBE", NatsTopic: natsTopic}
		}

	}
}

func (s *Server) isValidChannel(channel string) bool {
	for _, serverChannel := range s.channels {
		if strings.EqualFold(serverChannel, channel) {
			//s.logger.Debug("Channel exists in channels", zap.Bool("channel exists", true))
			return true
		}
	}

	s.logger.Debug("Channel does not exist", zap.Bool("exists", false))
	return false
}

func (s *Server) isValidTopic(topic string) bool {
	for _, serverTopic := range s.assets {

		if strings.EqualFold(serverTopic, topic) {
			//s.logger.Debug("Topic exists in assets", zap.Bool("topci exists", true))
			return true
		}
	}
	s.logger.Debug("Topic does NOT exist", zap.String("topic", topic))

	return false
}

func (s *Server) prometheusTotalConnections(ctx context.Context) {
	for action := range s.prometheusConnectionChan {
		select {
		case <-ctx.Done():
			return
		default:

			if action {
				atomic.AddInt32(&s.activeConns, 1)
				activeConnectionsGauge.Inc()
			} else {
				atomic.AddInt32(&s.activeConns, -1)
				activeConnectionsGauge.Dec()
			}
		}
	}

}

func (s *Server) prometheusTotalBridges(ctx context.Context) {
	var currentTotalBridges int
	var currentWorkWeightGauge int

	for count := range s.prometheusBridgeChan {
		select {
		case <-ctx.Done():
			return
		default:
			if currentTotalBridges >= 0 {
				currentTotalBridges += count

				currentWorkWeightGauge += count
				totalBridgesGauge.Set(float64(currentTotalBridges))

				total := int(s.activeConns)

				deflection := (float64(total) * float64(currentTotalBridges)) / float64(s.maxExtremum)

				totalWorkWeightGauge.Set(deflection * 100)
			}

		}
	}
}
