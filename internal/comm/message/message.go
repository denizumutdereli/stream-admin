package message

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/workers"
	"go.uber.org/zap"
)

type ContextMessages interface {
	SetContextualMessage(contextMessage *types.ContextualMessage) error
	GetContextualMessage(userID, messageType string, delivered ...bool) (types.ContextualMessage, error)
	BuildTopicName(userId string) string
	SendToNats(natsSubject string, contextMessage []byte) error
	SetListener(userID string) chan types.ContextualMessage
	GetListener(userID string) (chan types.ContextualMessage, bool)
	RemoveListener(userID string)
	CleanupListeners()
}

type contextMessages struct {
	ctx            context.Context
	cancel         context.CancelFunc
	config         *config.Config
	logger         *zap.Logger
	redis          *transport.RedisManager
	nats           *transport.NatsManager
	workerPool     *workers.WorkerPool
	mutex          sync.RWMutex
	listenersMap   map[string]chan types.ContextualMessage
	listenersMutex sync.RWMutex
}

func NewAdminContextMessageService(config *config.Config, redis *transport.RedisManager, nats *transport.NatsManager) ContextMessages {
	service := &contextMessages{
		config:       config,
		redis:        redis,
		nats:         nats,
		logger:       config.Logger,
		workerPool:   workers.NewWorkerPool(5),
		listenersMap: make(map[string]chan types.ContextualMessage),
	}

	ctx, cancel := context.WithCancel(context.Background())
	service.ctx = ctx
	service.cancel = cancel

	return service
}

func (a *contextMessages) SetContextualMessage(contextMessage *types.ContextualMessage) error {
	var err error
	var redisTimeout *int

	if !contextMessage.RedisDelivery && !contextMessage.NatsDelivery {
		return nil
	}

	operationCtx, cancel := context.WithTimeout(context.Background(), time.Duration(a.config.DefaultFuncsTimeOutInSeconds)*time.Second)
	defer cancel()

	redisKey := fmt.Sprintf("contextual_message:%s:%s", contextMessage.MessageType, contextMessage.UserId)

	workerCount := 0

	if contextMessage.RedisDelivery {
		if contextMessage.RedisTimeoutInMinutes != nil {
			redisTimeout = contextMessage.RedisTimeoutInMinutes
		} else {
			redisTimeout = &a.config.DefaultPanelAccessTokenTimeOut
		}

		workerCount++
		a.workerPool.Submit(func() error {
			select {
			case <-operationCtx.Done():
				return operationCtx.Err()
			default:
				a.mutex.Lock()
				defer a.mutex.Unlock()
				err := a.redis.SetKeyValue(operationCtx, redisKey, contextMessage,
					time.Duration(*redisTimeout)*time.Minute)
				if err != nil {
					return err
				}
				return nil
			}

		})
	}

	if contextMessage.NatsDelivery {
		workerCount++
		a.workerPool.Submit(func() error {
			select {
			case <-operationCtx.Done():
				return operationCtx.Err()
			default:
				messageJSON, err := json.Marshal(contextMessage)
				if err != nil {
					a.logger.Error("Failed to marshal context message to JSON", zap.Error(err))
					return err
				}

				topic := a.BuildTopicName(contextMessage.UserId)
				return a.SendToNats(topic, messageJSON)
			}

		})

	}

	for i := 0; i < workerCount; i++ {
		select {
		case err := <-a.workerPool.ErrorQueue:
			if err != nil {
				return err
			}
		case <-operationCtx.Done():
			return operationCtx.Err()
		}
	}

	if err != nil {
		a.logger.Error("Error sending context messages to nats and redis cluster", zap.Error(err))
		return err
	}

	workerCount = 0

	return nil
}

func (a *contextMessages) GetContextualMessage(userID, messageType string, delivered ...bool) (types.ContextualMessage, error) {
	var contextualMessageData types.ContextualMessage

	defaultDelivery := false
	if len(delivered) > 0 {
		defaultDelivery = delivered[0]
	}

	if a.redis == nil {
		a.logger.Error("No redis connection available for contextual messages")
		return types.ContextualMessage{}, errors.New("no redis connection available for contextual messages")
	}

	redisKey := fmt.Sprintf("contextual_message:%s:%s", messageType, userID)

	err := a.redis.GetKeyValue(a.ctx, redisKey, &contextualMessageData)
	if err != nil {
		return types.ContextualMessage{}, err
	}

	if defaultDelivery {
		err := a.redis.DeleteKey(a.ctx, redisKey)
		if err != nil {
			return types.ContextualMessage{}, err
		}
	}

	return contextualMessageData, nil
}

func (a *contextMessages) BuildTopicName(userId string) string {
	return fmt.Sprintf("%s.contextmessages", userId) // TODO: decide later on whether the topic name should be with user specific or general for react filtering purposes
}

func (a *contextMessages) SendToNats(natsSubject string, contextMessage []byte) error {

	err := a.nats.Publish(natsSubject, contextMessage)
	if err != nil {
		a.logger.Error("Failed to publish message", zap.Error(err))
		return err
	}

	return nil
}

func (a *contextMessages) SetListener(userID string) chan types.ContextualMessage {
	a.listenersMutex.Lock()
	defer a.listenersMutex.Unlock()

	if a.listenersMap == nil {
		a.listenersMap = make(map[string]chan types.ContextualMessage)
	}

	listenerChan, exists := a.listenersMap[userID]
	if !exists {
		listenerChan = make(chan types.ContextualMessage)
		a.listenersMap[userID] = listenerChan
	}

	return listenerChan
}

func (a *contextMessages) GetListener(userID string) (chan types.ContextualMessage, bool) {
	a.listenersMutex.RLock()
	defer a.listenersMutex.RUnlock()

	listenerChan, exists := a.listenersMap[userID]
	return listenerChan, exists
}

func (a *contextMessages) RemoveListener(userID string) {
	a.listenersMutex.Lock()
	defer a.listenersMutex.Unlock()

	if listenerChan, exists := a.listenersMap[userID]; exists {
		close(listenerChan)
		delete(a.listenersMap, userID)
	}
}

func (a *contextMessages) CleanupListeners() {
	a.listenersMutex.Lock()
	defer a.listenersMutex.Unlock()

	for userID, listenerChan := range a.listenersMap {
		close(listenerChan)
		delete(a.listenersMap, userID)
	}
}
