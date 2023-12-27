package registry

import (
	"github.com/denizumutdereli/stream-admin/internal/caesar"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/repository"
	"github.com/denizumutdereli/stream-admin/internal/repository/administrator/auth"
	"github.com/denizumutdereli/stream-admin/internal/repository/administrator/logs"
	"github.com/denizumutdereli/stream-admin/internal/repository/assets"
	"github.com/denizumutdereli/stream-admin/internal/repository/orders"
	"github.com/denizumutdereli/stream-admin/internal/repository/transactions"
	"github.com/denizumutdereli/stream-admin/internal/repository/users"

	contextMessage "github.com/denizumutdereli/stream-admin/internal/comm/message"

	adminPolicyRepo "github.com/denizumutdereli/stream-admin/internal/repository/administrator/policy"
	adminUserRolesRepo "github.com/denizumutdereli/stream-admin/internal/repository/administrator/roles"
	adminUsersRepo "github.com/denizumutdereli/stream-admin/internal/repository/administrator/users"

	"github.com/denizumutdereli/stream-admin/internal/service"
	administratorAuthService "github.com/denizumutdereli/stream-admin/internal/service/administrator/auth"
	administratorLogsService "github.com/denizumutdereli/stream-admin/internal/service/administrator/logs"
	administratorPolicyService "github.com/denizumutdereli/stream-admin/internal/service/administrator/policy"
	administratorRolesService "github.com/denizumutdereli/stream-admin/internal/service/administrator/roles"
	administratorUsersService "github.com/denizumutdereli/stream-admin/internal/service/administrator/users"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
)

type ServiceRegistry interface {
	RegisterAdminLogsService(repo *logs.AdminLogsRepository) (administratorLogsService.AdminLogsService, error)
	RegisterAdminAuthService(authRepo *auth.AdminAuthRepository, userRepo *adminUsersRepo.AdminUsersRepository, caesar caesar.CaesarManager, redis *transport.RedisManager, config *config.Config) (administratorAuthService.AdminAuthService, error)
	RegisterAdminUsersService(userRepo *adminUsersRepo.AdminUsersRepository, caesar caesar.CaesarManager, config *config.Config) (administratorUsersService.AdminUserService, error)
	RegisterAdminUserRolesService(userRolesRepo *adminUserRolesRepo.AdminUserRolesRepository, caesar caesar.CaesarManager, config *config.Config) (administratorRolesService.AdminUserRolesService, error)
	RegisterAdminPolicyService(policyRepo *adminPolicyRepo.AdminRolePolicyRepository, caesar caesar.CaesarManager, config *config.Config) (administratorPolicyService.AdminPolicyService, error)
	RegisterAdminContextMessageService(config *config.Config, redis *transport.RedisManager, nats *transport.NatsManager) (contextMessage.ContextMessages, error)
	RegisterAdminService(repo *repository.AdminRepository) (service.AdminService, error)

	GetAdminLogsService() (administratorLogsService.AdminLogsService, error)
	GetAdminAuthService() (administratorAuthService.AdminAuthService, error)
	GetAdminUsersService() (administratorUsersService.AdminUserService, error)
	GetAdminUserRolesService() (administratorRolesService.AdminUserRolesService, error)
	GetAdminPolicyService() (administratorPolicyService.AdminPolicyService, error)
	GetAdminContextMessageService() (contextMessage.ContextMessages, error)
	GetAdminService() (service.AdminService, error)

	RegisterAssetsService(repo *assets.AssetsRepository) (service.AssetsService, error)
	RegisterUsersService(repo *users.UsersRepository) (service.UsersService, error)
	RegisterTransactionsService(repo *transactions.TransactionRepository) (service.TransactionService, error)
	RegisterOrdersService(repo *orders.OrdersRepository) (service.OrdersService, error)

	GetAssetsService() (service.AssetsService, error)
	GetUsersService() (service.UsersService, error)
	GetTransactionsService() (service.TransactionService, error)
	GetOrdersService() (service.OrdersService, error)
}

type serviceRegistry struct {
	config     *config.Config
	logger     *zap.Logger
	appContext *types.ExchangeConfig

	// administrator services
	administratorLogsService           administratorLogsService.AdminLogsService
	administratorAuthService           administratorAuthService.AdminAuthService
	administratorUsersService          administratorUsersService.AdminUserService
	administratorUserRolesService      administratorRolesService.AdminUserRolesService
	administratorPolicyService         administratorPolicyService.AdminPolicyService
	administratorContextMessageService contextMessage.ContextMessages
	administratorService               service.AdminService

	// sub-services
	assetsService      service.AssetsService
	usersService       service.UsersService
	transactionService service.TransactionService
	ordersService      service.OrdersService
}

func NewServiceRegistry(appContext *types.ExchangeConfig) (ServiceRegistry, error) {

	service := &serviceRegistry{
		config:     appContext.Config,
		logger:     appContext.Logger,
		appContext: appContext,
	}

	return service, nil
}

/* Administrator ------------------------------------------------------------------ */

func (s *serviceRegistry) RegisterAdminAuthService(authRepo *auth.AdminAuthRepository, userRepo *adminUsersRepo.AdminUsersRepository, caesar caesar.CaesarManager, redis *transport.RedisManager, config *config.Config) (administratorAuthService.AdminAuthService, error) {
	if s.administratorAuthService == nil {
		service := administratorAuthService.NewAuthService(authRepo, userRepo, caesar, redis, s.config)
		s.administratorAuthService = service
		return service, nil
	}
	return s.administratorAuthService, nil
}

func (s *serviceRegistry) RegisterAdminLogsService(repo *logs.AdminLogsRepository) (administratorLogsService.AdminLogsService, error) {
	if s.administratorLogsService == nil {
		service := administratorLogsService.NewAdminLogsService(s.appContext, repo)
		s.administratorLogsService = service
		return service, nil
	}
	return s.administratorLogsService, nil
}

func (s *serviceRegistry) RegisterAdminUsersService(userRepo *adminUsersRepo.AdminUsersRepository, caesar caesar.CaesarManager, config *config.Config) (administratorUsersService.AdminUserService, error) {
	if s.administratorUsersService == nil {
		service := administratorUsersService.NewAdminUsersService(userRepo, caesar, s.appContext.Redis, s.config)
		s.administratorUsersService = service
		return service, nil
	}
	return s.administratorUsersService, nil
}

func (s *serviceRegistry) RegisterAdminUserRolesService(userRolesRepo *adminUserRolesRepo.AdminUserRolesRepository, caesar caesar.CaesarManager, config *config.Config) (administratorRolesService.AdminUserRolesService, error) {
	if s.administratorUserRolesService == nil {
		service := administratorRolesService.NewAdminUserRolesService(userRolesRepo, caesar, s.config)
		s.administratorUserRolesService = service
		return service, nil
	}
	return s.administratorUserRolesService, nil
}

func (s *serviceRegistry) RegisterAdminPolicyService(policyRepo *adminPolicyRepo.AdminRolePolicyRepository, caesar caesar.CaesarManager, config *config.Config) (administratorPolicyService.AdminPolicyService, error) {
	if s.administratorPolicyService == nil {
		service := administratorPolicyService.NewAdminPolicyService(policyRepo, caesar, s.config)
		s.administratorPolicyService = service
		return service, nil
	}
	return s.administratorPolicyService, nil
}

func (s *serviceRegistry) RegisterAdminContextMessageService(config *config.Config, redis *transport.RedisManager, nats *transport.NatsManager) (contextMessage.ContextMessages, error) {
	if s.administratorContextMessageService == nil {
		service := contextMessage.NewAdminContextMessageService(config, redis, nats)
		s.administratorContextMessageService = service
		return service, nil
	}
	return s.administratorContextMessageService, nil
}

func (s *serviceRegistry) RegisterAdminService(repo *repository.AdminRepository) (service.AdminService, error) {
	if s.administratorService == nil {
		service := service.NewAdminService(s.appContext, repo)
		s.administratorService = service
		return service, nil
	}
	return s.administratorService, nil
}

func (s *serviceRegistry) GetAdminLogsService() (administratorLogsService.AdminLogsService, error) {
	return s.administratorLogsService, nil
}

func (s *serviceRegistry) GetAdminAuthService() (administratorAuthService.AdminAuthService, error) {
	return s.administratorAuthService, nil
}

func (s *serviceRegistry) GetAdminUsersService() (administratorUsersService.AdminUserService, error) {
	return s.administratorUsersService, nil
}

func (s *serviceRegistry) GetAdminUserRolesService() (administratorRolesService.AdminUserRolesService, error) {
	return s.administratorUserRolesService, nil
}

func (s *serviceRegistry) GetAdminPolicyService() (administratorPolicyService.AdminPolicyService, error) {
	return s.administratorPolicyService, nil
}

func (s *serviceRegistry) GetAdminContextMessageService() (contextMessage.ContextMessages, error) {
	return s.administratorContextMessageService, nil
}

func (s *serviceRegistry) GetAdminService() (service.AdminService, error) {
	return s.administratorService, nil
}

/* -------------------------------------------------------------------------------- */

func (s *serviceRegistry) RegisterAssetsService(repo *assets.AssetsRepository) (service.AssetsService, error) {
	if s.assetsService == nil {
		service := service.NewAssetsService(s.appContext, repo)
		s.assetsService = service
		return service, nil
	}
	return s.assetsService, nil
}

func (s *serviceRegistry) RegisterUsersService(repo *users.UsersRepository) (service.UsersService, error) {
	if s.usersService == nil {
		service := service.NewUsersService(s.appContext, repo)
		s.usersService = service
		return service, nil
	}
	return s.usersService, nil
}

func (s *serviceRegistry) RegisterTransactionsService(repo *transactions.TransactionRepository) (service.TransactionService, error) {
	if s.transactionService == nil {
		service := service.NewTransactionService(s.appContext, repo)
		s.transactionService = service
		return service, nil
	}
	return s.transactionService, nil
}

func (s *serviceRegistry) RegisterOrdersService(repo *orders.OrdersRepository) (service.OrdersService, error) {
	if s.ordersService == nil {
		service := service.NewOrdersService(s.appContext, repo)
		s.ordersService = service
		return service, nil
	}
	return s.ordersService, nil
}

func (s *serviceRegistry) GetAssetsService() (service.AssetsService, error) {
	return s.assetsService, nil
}

func (s *serviceRegistry) GetUsersService() (service.UsersService, error) {
	return s.usersService, nil
}

func (s *serviceRegistry) GetTransactionsService() (service.TransactionService, error) {
	return s.transactionService, nil
}

func (s *serviceRegistry) GetOrdersService() (service.OrdersService, error) {
	return s.ordersService, nil
}
