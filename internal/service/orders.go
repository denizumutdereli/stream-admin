package service

import (
	"context"
	"net/http"

	appErrors "github.com/denizumutdereli/stream-admin/internal/common"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/orders"
	repos "github.com/denizumutdereli/stream-admin/internal/repository/orders"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
)

type OrdersService interface {
	GetAll(paginationParams *types.PaginationParams, queryParams *models.OrderSearch) (*database.PaginatedResult, appErrors.Error)
}

type ordersService struct {
	ctx              context.Context
	cancel           context.CancelFunc
	config           *config.Config
	logger           *zap.Logger
	redis            *transport.RedisManager
	repo             repos.OrdersRepository
	connectionStatus bool
}

func NewOrdersService(appContext *types.ExchangeConfig, repo *repos.OrdersRepository) OrdersService {
	service := &ordersService{
		config:           appContext.Config,
		redis:            appContext.Redis,
		logger:           appContext.Logger,
		repo:             *repo,
		connectionStatus: true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	service.ctx = ctx
	service.cancel = cancel

	return service
}

func (s *ordersService) GetAll(paginationParams *types.PaginationParams, queryParams *models.OrderSearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.repo.GetAll(paginationParams, queryParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}
