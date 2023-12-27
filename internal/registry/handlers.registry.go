package registry

import (
	"errors"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/handler"
	"github.com/denizumutdereli/stream-admin/internal/service"
	"go.uber.org/zap"

	administratorAuthHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/auth"
	administratorLogsHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/logs"
	administratorPolicyHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/policy"
	administratorUserRolesHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/roles"
	administratorUserHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/user"

	administratorAuthService "github.com/denizumutdereli/stream-admin/internal/service/administrator/auth"
	administratorUsersService "github.com/denizumutdereli/stream-admin/internal/service/administrator/users"
	administratorLogsService "github.com/denizumutdereli/stream-admin/internal/service/administrator/logs"
	administratorPolicyService "github.com/denizumutdereli/stream-admin/internal/service/administrator/policy"
	administratorRolesService "github.com/denizumutdereli/stream-admin/internal/service/administrator/roles"
)

type HandlersRegistry interface {
	/* Administrator ---------------------------------------------------------------------------- */
	RegisterAdminRestHandler(service service.AdminService) (*handler.AdminRestHandler, error)
	RegisterAdminUsersHandler(service *administratorUsersService.AdminUserService) (*administratorUserHandler.AdminUserHandler, error)
	RegisterAdminUserRolesHandler(service *administratorRolesService.AdminUserRolesService) (*administratorUserRolesHandler.AdminUserRolesHandler, error)
	RegisterAdminAuthHandler(service *administratorAuthService.AdminAuthService) (*administratorAuthHandler.AdminAuthHandler, error)
	RegisterAdminLogsHandler(service administratorLogsService.AdminLogsService) (*administratorLogsHandler.AdminLogsRestHandler, error)
	RegisterAdminPolicyHandler(service *administratorPolicyService.AdminPolicyService) (*administratorPolicyHandler.AdminPolicyHandler, error)

	GetAdminRestHandler() (handler.AdminRestHandler, error)
	GetAdminUsersHandler() (administratorUserHandler.AdminUserHandler, error)
	GetAdminUserRolesHandler() (administratorUserRolesHandler.AdminUserRolesHandler, error)
	GetAdminAuthHandler() (administratorAuthHandler.AdminAuthHandler, error)
	GetAdminLogsHandler() (administratorLogsHandler.AdminLogsRestHandler, error)
	GetAdminPolicyHandler() (administratorPolicyHandler.AdminPolicyHandler, error)
	/* ------------------------------------------------------------------------------------------- */

	RegisterOrdersRestHandler(service service.OrdersService) (*handler.OrdersRestHandler, error)
	RegisterTransactionRestHandler(service service.TransactionService) (*handler.TransactionsRestHandler, error)
	RegisterUsersHandler(service service.UsersService) (*handler.UsersRestHandler, error)
	RegisterAssetsRestHandler(service service.AssetsService) (*handler.AssetsRestHandler, error)

	GetOrdersRestHandler() (handler.OrdersRestHandler, error)
	GetTransactionsRestHandler() (handler.TransactionsRestHandler, error)
	GetUsersRestHandler() (handler.UsersRestHandler, error)
	GetAssetsRestHandler() (handler.AssetsRestHandler, error)
}

type handlersRegistry struct {
	config                *config.Config
	logger                *zap.Logger
	builders              builders.BuilderService
	adminRestHandler      handler.AdminRestHandler
	adminUsersHandler     administratorUserHandler.AdminUserHandler
	adminUserRolesHandler administratorUserRolesHandler.AdminUserRolesHandler
	adminAuthHandler      administratorAuthHandler.AdminAuthHandler
	adminLogsHandler      administratorLogsHandler.AdminLogsRestHandler
	adminPolicyHandler    administratorPolicyHandler.AdminPolicyHandler
	ordersHandler         handler.OrdersRestHandler
	transactionsHandler   handler.TransactionsRestHandler
	usersHandler          handler.UsersRestHandler
	assetsHandler         handler.AssetsRestHandler
}

func NewHandlersRegistry(config *config.Config, builders builders.BuilderService) (HandlersRegistry, error) {

	service := &handlersRegistry{
		config:   config,
		logger:   config.Logger,
		builders: builders,
	}

	return service, nil
}

/* Administrator --------------------------------------------------------------------------- */

func (h *handlersRegistry) RegisterAdminRestHandler(service service.AdminService) (*handler.AdminRestHandler, error) {
	if h.adminRestHandler == nil {
		var err error
		h.logger.Debug("Admin rest handler is not registered, registering it now")

		handler := handler.NewAdminRestHandler(service, h.config) // TODO: return error as well

		if err != nil {
			h.logger.Fatal("service handler creation error:", zap.Error(err))
			return nil, err
		}

		if handler == nil {
			return nil, errors.New("received nil adminRest handler")
		}

		h.adminRestHandler = handler
		return &handler, nil
	}

	return &h.adminRestHandler, nil
}

func (h *handlersRegistry) RegisterAdminUsersHandler(service *administratorUsersService.AdminUserService) (*administratorUserHandler.AdminUserHandler, error) {
	if h.adminUsersHandler == nil {
		var err error
		h.logger.Debug("Admin users handler is not registered, registering it now")

		handler := administratorUserHandler.NewAdminUserHandler(service, h.config, h.builders) // TODO: return error as well

		if err != nil {
			h.logger.Fatal("service handler creation error:", zap.Error(err))
			return nil, err
		}

		if handler == nil {
			return nil, errors.New("received nil adminUsers handler")
		}

		h.adminUsersHandler = handler
		return &handler, nil
	}

	return &h.adminUsersHandler, nil
}

func (h *handlersRegistry) RegisterAdminUserRolesHandler(service *administratorRolesService.AdminUserRolesService) (*administratorUserRolesHandler.AdminUserRolesHandler, error) {
	if h.adminUserRolesHandler == nil {
		var err error
		h.logger.Debug("Admin users handler is not registered, registering it now")

		handler := administratorUserRolesHandler.NewAdminUserRolesHandler(service, h.config, h.builders) // TODO: return error as well

		if err != nil {
			h.logger.Fatal("service handler creation error:", zap.Error(err))
			return nil, err
		}

		if handler == nil {
			return nil, errors.New("received nil adminUsers handler")
		}

		h.adminUserRolesHandler = handler
		return &handler, nil
	}

	return &h.adminUserRolesHandler, nil
}

func (h *handlersRegistry) RegisterAdminAuthHandler(service *administratorAuthService.AdminAuthService) (*administratorAuthHandler.AdminAuthHandler, error) {
	if h.adminAuthHandler == nil {
		var err error
		h.logger.Debug("Admin auth handler is not registered, registering it now")

		handler := administratorAuthHandler.NewAdminAuthHandler(service, h.config, h.builders) // TODO: return error as well

		if err != nil {
			h.logger.Fatal("service handler creation error:", zap.Error(err))
			return nil, err
		}

		if handler == nil {
			return nil, errors.New("received nil adminAuth handler")
		}

		h.adminAuthHandler = handler
		return &handler, nil
	}

	return &h.adminAuthHandler, nil
}

func (h *handlersRegistry) RegisterAdminLogsHandler(service administratorLogsService.AdminLogsService) (*administratorLogsHandler.AdminLogsRestHandler, error) {
	if h.adminLogsHandler == nil {
		var err error
		h.logger.Debug("Admin logs handler is not registered, registering it now")

		handler := administratorLogsHandler.NewAdminLogsRestHandler(service, h.config, h.builders) // TODO: return error as well

		if err != nil {
			h.logger.Fatal("service handler creation error:", zap.Error(err))
			return nil, err
		}

		if handler == nil {
			return nil, errors.New("received nil adminLogs handler")
		}

		h.adminLogsHandler = handler
		return &handler, nil
	}

	return &h.adminLogsHandler, nil
}

func (h *handlersRegistry) RegisterAdminPolicyHandler(service *administratorPolicyService.AdminPolicyService) (*administratorPolicyHandler.AdminPolicyHandler, error) {
	if h.adminPolicyHandler == nil {
		var err error
		h.logger.Debug("Admin policy handler is not registered, registering it now")

		handler := administratorPolicyHandler.NewAdminPolicyHandler(service, h.config, h.builders) // TODO: return error as well

		if err != nil {
			h.logger.Fatal("service handler creation error:", zap.Error(err))
			return nil, err
		}

		if handler == nil {
			return nil, errors.New("received nil adminLogs handler")
		}

		h.adminPolicyHandler = handler
		return &handler, nil
	}

	return &h.adminPolicyHandler, nil
}

func (h *handlersRegistry) GetAdminRestHandler() (handler.AdminRestHandler, error) {
	return h.adminRestHandler, nil
}

func (h *handlersRegistry) GetAdminUsersHandler() (administratorUserHandler.AdminUserHandler, error) {
	return h.adminUsersHandler, nil
}

func (h *handlersRegistry) GetAdminUserRolesHandler() (administratorUserRolesHandler.AdminUserRolesHandler, error) {
	return h.adminUserRolesHandler, nil
}

func (h *handlersRegistry) GetAdminAuthHandler() (administratorAuthHandler.AdminAuthHandler, error) {
	return h.adminAuthHandler, nil
}

func (h *handlersRegistry) GetAdminLogsHandler() (administratorLogsHandler.AdminLogsRestHandler, error) {
	return h.adminLogsHandler, nil
}

func (h *handlersRegistry) GetAdminPolicyHandler() (administratorPolicyHandler.AdminPolicyHandler, error) {
	return h.adminPolicyHandler, nil
}

/* ---------------------------------------------------------------------------------------- */

func (h *handlersRegistry) RegisterOrdersRestHandler(service service.OrdersService) (*handler.OrdersRestHandler, error) {
	if h.ordersHandler == nil {
		var err error
		h.logger.Debug("Orders service handler is not registered, registering it now")

		handler := handler.NewOrdersRestHandler(service, h.config, h.builders) // TODO: return error as well

		if err != nil {
			h.logger.Fatal("service handler creation error:", zap.Error(err))
			return nil, err
		}

		if handler == nil {
			return nil, errors.New("received serviceOrders nil handler")
		}

		h.ordersHandler = handler
		return &handler, nil
	}

	return &h.ordersHandler, nil
}

func (h *handlersRegistry) RegisterTransactionRestHandler(service service.TransactionService) (*handler.TransactionsRestHandler, error) {
	if h.transactionsHandler == nil {
		var err error
		h.logger.Debug("Transactions service handler is not registered, registering it now")

		handler := handler.NewTransactionsRestHandler(service, h.config, h.builders)

		if err != nil {
			h.logger.Fatal("service handler creation error:", zap.Error(err))
			return nil, err
		}

		if handler == nil {
			return nil, errors.New("received serviceTransactions nil handler")
		}

		h.transactionsHandler = handler
		return &handler, nil
	}

	return &h.transactionsHandler, nil
}

func (h *handlersRegistry) RegisterUsersHandler(service service.UsersService) (*handler.UsersRestHandler, error) {
	if h.usersHandler == nil {
		var err error
		h.logger.Debug("Users service handler is not registered, registering it now")

		handler := handler.NewUsersRestHandler(service, h.config, h.builders) // TODO: return error as well

		if err != nil {
			h.logger.Fatal("service handler creation error:", zap.Error(err))
			return nil, err
		}

		if handler == nil {
			return nil, errors.New("received serviceUsers nil handler")
		}

		h.usersHandler = handler
		return &handler, nil
	}

	return &h.usersHandler, nil
}

func (h *handlersRegistry) RegisterAssetsRestHandler(service service.AssetsService) (*handler.AssetsRestHandler, error) {
	if h.assetsHandler == nil {
		var err error
		h.logger.Debug("Users service handler is not registered, registering it now")

		handler := handler.NewAssetsRestHandler(service, h.config, h.builders) // TODO: return error as well

		if err != nil {
			h.logger.Fatal("service handler creation error:", zap.Error(err))
			return nil, err
		}

		if handler == nil {
			return nil, errors.New("received serviceAssets nil handler")
		}

		h.assetsHandler = handler
		return &handler, nil
	}

	return &h.assetsHandler, nil
}

func (h *handlersRegistry) GetOrdersRestHandler() (handler.OrdersRestHandler, error) {
	return h.ordersHandler, nil
}

func (h *handlersRegistry) GetTransactionsRestHandler() (handler.TransactionsRestHandler, error) {
	return h.transactionsHandler, nil
}

func (h *handlersRegistry) GetUsersRestHandler() (handler.UsersRestHandler, error) {
	return h.usersHandler, nil
}

func (h *handlersRegistry) GetAssetsRestHandler() (handler.AssetsRestHandler, error) {
	return h.assetsHandler, nil
}
