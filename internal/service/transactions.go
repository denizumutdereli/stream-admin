package service

import (
	"context"
	"net/http"

	appErrors "github.com/denizumutdereli/stream-admin/internal/common"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/transactions"
	repos "github.com/denizumutdereli/stream-admin/internal/repository/transactions"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
)

type TransactionService interface {
	GetSearchParameters() ([]types.SearchParameters, appErrors.Error)
	GetFiatTransactions(paginationParams *types.PaginationParams, searchParams *models.FiatTransactionsSearch) (*database.PaginatedResult, appErrors.Error)
	GetCryptoTransactions(paginationParams *types.PaginationParams, searchParams *models.CryptoTransactionsSearch) (*database.PaginatedResult, appErrors.Error)
	GetCryptoWallets(paginationParams *types.PaginationParams, searchParams *models.CryptoWalletsSearch) (*database.PaginatedResult, appErrors.Error)
}

type transactionService struct {
	ctx              context.Context
	cancel           context.CancelFunc
	config           *config.Config
	logger           *zap.Logger
	redis            *transport.RedisManager
	repo             repos.TransactionRepository
	connectionStatus bool
}

func NewTransactionService(appContext *types.ExchangeConfig, repo *repos.TransactionRepository) TransactionService {
	service := &transactionService{
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

func (s *transactionService) GetSearchParameters() ([]types.SearchParameters, appErrors.Error) {
	data, err := s.repo.GetSearchParameters()
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}

func (s *transactionService) GetFiatTransactions(paginationParams *types.PaginationParams, searchParams *models.FiatTransactionsSearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.repo.GetFiatTransactions(paginationParams, searchParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}

func (s *transactionService) GetCryptoTransactions(paginationParams *types.PaginationParams, searchParams *models.CryptoTransactionsSearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.repo.GetCryptoTransactions(paginationParams, searchParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}

func (s *transactionService) GetCryptoWallets(paginationParams *types.PaginationParams, searchParams *models.CryptoWalletsSearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.repo.GetCryptoWallets(paginationParams, searchParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}
