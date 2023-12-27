package service

import (
	"context"
	"net/http"

	appErrors "github.com/denizumutdereli/stream-admin/internal/common"

	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/assets"
	repos "github.com/denizumutdereli/stream-admin/internal/repository/assets"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
)

type AssetsService interface {
	GetSearchParameters() ([]types.SearchParameters, appErrors.Error)
	GetCoins(paginationParams *types.PaginationParams, queryParams *models.AssetsCoinsSearch) (*database.PaginatedResult, appErrors.Error)
	GetAssets(paginationParams *types.PaginationParams, queryParams *models.AssetsSearch) (*database.PaginatedResult, appErrors.Error)
	GetNetworks(paginationParams *types.PaginationParams, queryParams *models.AssetsNetworksSearch) (*database.PaginatedResult, appErrors.Error)
}

type assetsService struct {
	ctx              context.Context
	cancel           context.CancelFunc
	config           *config.Config
	logger           *zap.Logger
	redis            *transport.RedisManager
	repo             repos.AssetsRepository
	connectionStatus bool
}

func NewAssetsService(appContext *types.ExchangeConfig, repo *repos.AssetsRepository) AssetsService {
	service := &assetsService{
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

func (s *assetsService) GetSearchParameters() ([]types.SearchParameters, appErrors.Error) {
	data, err := s.repo.GetSearchParameters()
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}

func (s *assetsService) GetCoins(paginationParams *types.PaginationParams, queryParams *models.AssetsCoinsSearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.repo.GetCoins(paginationParams, queryParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}

func (s *assetsService) GetAssets(paginationParams *types.PaginationParams, queryParams *models.AssetsSearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.repo.GetAssets(paginationParams, queryParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}

func (s *assetsService) GetNetworks(paginationParams *types.PaginationParams, queryParams *models.AssetsNetworksSearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.repo.GetNetworks(paginationParams, queryParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}
