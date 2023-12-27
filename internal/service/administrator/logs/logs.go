package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	appErrors "github.com/denizumutdereli/stream-admin/internal/common"

	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"
	logs "github.com/denizumutdereli/stream-admin/internal/repository/administrator/logs"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/denizumutdereli/stream-admin/internal/workers"

	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AdminLogsService interface {
	GetAll(paginationParams *types.PaginationParams, queryParams *models.AdministratorLogsSearch) (*database.PaginatedResult, appErrors.Error)
	LogAction(c *gin.Context, userRole, UserID string, loglevel ...int) appErrors.Error
	BuildTopicName(loglevel ...int) string
	SendToNats(natsSubject string, logMessage []byte) appErrors.Error
}

type adminLogsService struct {
	ctx              context.Context
	cancel           context.CancelFunc
	config           *config.Config
	logger           *zap.Logger
	nats             *transport.NatsManager
	redis            *transport.RedisManager
	logTopics        []string
	repo             logs.AdminLogsRepository
	workerPool       *workers.WorkerPool
	mutex            sync.RWMutex
	connectionStatus bool
}

func NewAdminLogsService(appContext *types.ExchangeConfig, repo *logs.AdminLogsRepository) AdminLogsService {
	service := &adminLogsService{
		config:           appContext.Config,
		redis:            appContext.Redis,
		logger:           appContext.Logger,
		logTopics:        appContext.Config.AdminActionLogTopics,
		nats:             appContext.Nats,
		repo:             *repo,
		workerPool:       workers.NewWorkerPool(5),
		connectionStatus: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(service.config.DefaultFuncsTimeOutInSeconds)*time.Second)
	service.ctx = ctx
	service.cancel = cancel

	return service
}

func (s *adminLogsService) GetAll(paginationParams *types.PaginationParams, queryParams *models.AdministratorLogsSearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.repo.GetAll(paginationParams, queryParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusBadRequest, "", err.Error(), err)
	}
	return data, nil
}

func (a *adminLogsService) LogAction(c *gin.Context, userRole, UserID string, loglevel ...int) appErrors.Error {

	actualLogLevel := 1
	if len(loglevel) > 0 {
		actualLogLevel = loglevel[0]
	}

	if strings.Contains(c.Request.RequestURI, "/admin/logs") {
		return nil
	}

	actionData := &models.AdministratorLogs{
		LogLevel:   actualLogLevel,
		UserID:     UserID,
		UserRole:   userRole,
		Action:     c.Request.RequestURI,
		Method:     c.Request.Method,
		Ip:         utils.GetClientIP(c),
		Status:     c.Writer.Status(),
		UserAgent:  c.Request.UserAgent(),
		Timestamps: time.Now(),
	}

	actionJSON, err := json.Marshal(actionData)
	if err != nil {
		a.logger.Error("Failed to marshal action to JSON", zap.Error(err))
		return appErrors.AppError(http.StatusServiceUnavailable, "", "Failed to marshal action to JSON", err)
	}

	operationCtx, cancel := context.WithTimeout(context.Background(), time.Duration(a.config.DefaultFuncsTimeOutInSeconds)*time.Second)
	defer cancel()

	//a.logger.Debug(userRole+" action taken", zap.ByteString("action", actionJSON))

	a.workerPool.Submit(func() error {
		select {
		case <-operationCtx.Done():
			return operationCtx.Err()
		default:
			topic := a.BuildTopicName(actualLogLevel)
			return a.SendToNats(topic, actionJSON)
		}
	})

	a.workerPool.Submit(func() error {
		select {
		case <-operationCtx.Done():
			return operationCtx.Err()
		default:
			a.mutex.Lock()
			defer a.mutex.Unlock()
			return a.repo.Create(actionData)
		}
	})

	for i := 0; i < 2; i++ {
		select {
		case err := <-a.workerPool.ErrorQueue:
			if err != nil {
				return appErrors.AppError(http.StatusServiceUnavailable, "", err.Error(), err)
			}
		case <-operationCtx.Done():
			return appErrors.AppError(http.StatusServiceUnavailable, "", operationCtx.Err().Error(), operationCtx.Err())
		}
	}

	if err != nil {
		a.logger.Error("Error sending admin logs to nats stream cluster & save to db", zap.Error(err))
		return appErrors.AppError(http.StatusServiceUnavailable, "", "Error sending admin logs to nats stream cluster & save to db", err)
	}

	return nil
}

func (a *adminLogsService) BuildTopicName(loglevel ...int) string {
	actualLogLevel := 0
	if len(loglevel) > 0 {
		actualLogLevel = loglevel[0]
	}

	if actualLogLevel < 0 || actualLogLevel >= len(a.logTopics) {
		return "adminactivities.default"
	}

	return fmt.Sprintf("adminactivities.%s", a.logTopics[actualLogLevel])
}

func (a *adminLogsService) SendToNats(natsSubject string, logMessage []byte) appErrors.Error {
	err := a.nats.Publish(natsSubject, logMessage)
	if err != nil {
		a.logger.Error("Failed to publish message", zap.Error(err))
		return appErrors.AppError(http.StatusServiceUnavailable, "", "Failed logs to publish message", err)

	}

	return nil
}
