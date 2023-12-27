package service

import (
	"context"
	"net/http"

	appErrors "github.com/denizumutdereli/stream-admin/internal/common"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/users"
	userRepos "github.com/denizumutdereli/stream-admin/internal/repository/users"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
)

type UsersService interface {
	GetSearchUserParameters() ([]types.SearchParameters, appErrors.Error)
	GetSearchKYCParameters() ([]types.SearchParameters, appErrors.Error)
	GetUsers(paginationParams *types.PaginationParams, searchParams *models.UserSearch) (*database.PaginatedResult, appErrors.Error)
	// GetUserDetailsBuilder(userId int, includeDetails *userTypes.UserDetailsIncluding, paginationParams *types.PaginationParams) (*database.DataResult, error)
	GetKYC(paginationParams *types.PaginationParams, searchParams *models.UserKYCSearch) (*database.PaginatedResult, appErrors.Error)
}

type usersService struct {
	ctx    context.Context
	cancel context.CancelFunc
	config *config.Config
	logger *zap.Logger
	redis  *transport.RedisManager
	repo   userRepos.UsersRepository
	// orderRepo        ordersRepos.OrdersRepository
	// transactionRepo  transactionsRepos.TransactionRepository
	connectionStatus bool
}

func NewUsersService(appContext *types.ExchangeConfig, repo *userRepos.UsersRepository) UsersService {
	service := &usersService{
		config: appContext.Config,
		redis:  appContext.Redis,
		logger: appContext.Logger,
		repo:   *repo,
		//repoRegistry:     repoRegistry,
		connectionStatus: true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	service.ctx = ctx
	service.cancel = cancel

	return service
}

// func (s *usersService) injectExternalOrderRepository() error {
// 	if s.orderRepo == nil {
// 		ordersRepo, err := s.repoRegistry.GetOrdersRepository()
// 		if err != nil {
// 			return fmt.Errorf("error injecting order repository to users service: %w", err)
// 		}
// 		s.orderRepo = ordersRepo
// 	}
// 	return nil
// }

// func (s *usersService) injectExternalTransactionsRepository() error {
// 	if s.transactionRepo == nil {
// 		transactionsRepo, err := s.repoRegistry.GetTransactionsRepository()
// 		if err != nil {
// 			return fmt.Errorf("error injecting order repository to users service: %w", err)
// 		}
// 		s.transactionRepo = transactionsRepo
// 	}
// 	return nil
// }

func (s *usersService) GetSearchUserParameters() ([]types.SearchParameters, appErrors.Error) {
	data, err := s.repo.GetSearchUserParameters()
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}

func (s *usersService) GetSearchKYCParameters() ([]types.SearchParameters, appErrors.Error) {
	data, err := s.repo.GetSearchKYCParameters()
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}

func (s *usersService) GetUsers(paginationParams *types.PaginationParams, searchParams *models.UserSearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.repo.GetUsers(paginationParams, searchParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}

// func (s *usersService) GetUserDetailsBuilder(userId int, includeDetails *userTypes.UserDetailsIncluding, paginationParams *types.PaginationParams) (*database.DataResult, error) {
// 	if err := s.injectExternalOrderRepository(); err != nil {
// 		return nil, err
// 	}

// 	if err := s.injectExternalTransactionsRepository(); err != nil {
// 		return nil, err
// 	}
// 	return s.repo.GetUserDetailsBuilder(userId, includeDetails, paginationParams, s.orderRepo, s.transactionRepo)
// }

func (s *usersService) GetKYC(paginationParams *types.PaginationParams, searchParams *models.UserKYCSearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.repo.GetKYC(paginationParams, searchParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}
