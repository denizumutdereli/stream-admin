package registry

import (
	"errors"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	contextMessages "github.com/denizumutdereli/stream-admin/internal/comm/message"
	"github.com/denizumutdereli/stream-admin/internal/config"

	"github.com/denizumutdereli/stream-admin/internal/repository"
	administratorAuthRepo "github.com/denizumutdereli/stream-admin/internal/repository/administrator/auth"
	administratorLogsRepo "github.com/denizumutdereli/stream-admin/internal/repository/administrator/logs"
	administratorPolicyRepo "github.com/denizumutdereli/stream-admin/internal/repository/administrator/policy"
	administratorUserRolesRepo "github.com/denizumutdereli/stream-admin/internal/repository/administrator/roles"
	administratorUsersRepo "github.com/denizumutdereli/stream-admin/internal/repository/administrator/users"

	"github.com/denizumutdereli/stream-admin/internal/transport"

	"github.com/denizumutdereli/stream-admin/internal/repository/assets"
	"github.com/denizumutdereli/stream-admin/internal/repository/orders"
	"github.com/denizumutdereli/stream-admin/internal/repository/transactions"
	"github.com/denizumutdereli/stream-admin/internal/repository/users"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RepositoryRegistry interface {

	// Admin registry
	RegisterAdminRepository() (*repository.AdminRepository, error)
	RegisterAdminUsersRepository(servicePrefix string) (*administratorUsersRepo.AdminUsersRepository, error)
	RegisterAdminUserRolesRepository(servicePrefix string) (*administratorUserRolesRepo.AdminUserRolesRepository, error)
	RegisterAdminAuthRepository(servicePrefix string, contextMessage contextMessages.ContextMessages) (*administratorAuthRepo.AdminAuthRepository, error)
	RegisterAdminLogsRepository(servicePrefix string) (*administratorLogsRepo.AdminLogsRepository, error)
	RegisterAdminPolicyRepository(servicePrefix string) (*administratorPolicyRepo.AdminRolePolicyRepository, error)

	// Sub-services registry
	RegisterOrdersRepository(servicePrefix string) (*orders.OrdersRepository, error)
	RegisterTransactionsRepository(servicePrefix string) (*transactions.TransactionRepository, error)
	RegisterUsersRepository(servicePrefix string) (*users.UsersRepository, error)
	RegisterAssetsRepository(servicePrefix string) (*assets.AssetsRepository, error)

	// Admin registry getter
	GetAdminRepository() (repository.AdminRepository, error)
	GetAdminUsersRepository() (administratorUsersRepo.AdminUsersRepository, error)
	GetAdminUserRolesRepository() (administratorUserRolesRepo.AdminUserRolesRepository, error)
	GetAdminAuthRepository() (administratorAuthRepo.AdminAuthRepository, error)
	GetAdminLogsRepository() (administratorLogsRepo.AdminLogsRepository, error)
	GetAdminPolicyRepository() (administratorPolicyRepo.AdminRolePolicyRepository, error)

	// Sub-services registry getter
	GetOrdersRepository() (orders.OrdersRepository, error)
	GetUsersRepository() (users.UsersRepository, error)
	GetTransactionsRepository() (transactions.TransactionRepository, error)
	GetAssetsRepository() (assets.AssetsRepository, error)
}

type repositoryRegistry struct {
	db       *gorm.DB
	config   *config.Config
	logger   *zap.Logger
	builders builders.BuilderService
	redis    *transport.RedisManager

	admin       repository.AdminRepository
	adminUsers  administratorUsersRepo.AdminUsersRepository
	adminRoles  administratorUserRolesRepo.AdminUserRolesRepository
	adminLogs   administratorLogsRepo.AdminLogsRepository
	adminAuth   administratorAuthRepo.AdminAuthRepository
	adminPolicy administratorPolicyRepo.AdminRolePolicyRepository
	//adminContextMessages contextMessage.ContextMessages
	orders       orders.OrdersRepository
	transactions transactions.TransactionRepository
	users        users.UsersRepository
	assets       assets.AssetsRepository
}

func NewRepositoryRegistry(db *gorm.DB, config *config.Config, builders builders.BuilderService, redis *transport.RedisManager) (RepositoryRegistry, error) {

	service := &repositoryRegistry{
		db:       db,
		config:   config,
		logger:   config.Logger,
		builders: builders,
		redis:    redis,
	}

	return service, nil
}

func (r *repositoryRegistry) GetAdminRepository() (repository.AdminRepository, error) {
	return r.admin, nil
}

func (r *repositoryRegistry) GetAdminUsersRepository() (administratorUsersRepo.AdminUsersRepository, error) {
	return r.adminUsers, nil
}

func (r *repositoryRegistry) GetAdminUserRolesRepository() (administratorUserRolesRepo.AdminUserRolesRepository, error) {
	return r.adminRoles, nil
}

func (r *repositoryRegistry) GetAdminAuthRepository() (administratorAuthRepo.AdminAuthRepository, error) {
	return r.adminAuth, nil
}

func (r *repositoryRegistry) GetAdminLogsRepository() (administratorLogsRepo.AdminLogsRepository, error) {
	return r.adminLogs, nil
}

func (r *repositoryRegistry) GetAdminPolicyRepository() (administratorPolicyRepo.AdminRolePolicyRepository, error) {
	return r.adminPolicy, nil
}

func (r *repositoryRegistry) GetAdminAdminRepository() (administratorPolicyRepo.AdminRolePolicyRepository, error) {
	return r.adminPolicy, nil
}

func (r *repositoryRegistry) GetOrdersRepository() (orders.OrdersRepository, error) {
	return r.orders, nil
}
func (r *repositoryRegistry) GetUsersRepository() (users.UsersRepository, error) {
	return r.users, nil
}
func (r *repositoryRegistry) GetTransactionsRepository() (transactions.TransactionRepository, error) {
	return r.transactions, nil
}

func (r *repositoryRegistry) GetAssetsRepository() (assets.AssetsRepository, error) {
	return r.assets, nil
}

func (r *repositoryRegistry) RegisterAdminRepository() (*repository.AdminRepository, error) {
	if r.admin == nil {
		var err error
		r.logger.Debug("orders repository is not registered, registering it now")
		r.admin = repository.NewGORMAdminRepository(r.db)

		if err != nil {
			r.logger.Fatal("service repository creation error:", zap.Error(err))
		}

		if r.admin == nil {
			return nil, errors.New("failed to initialize admin repository")
		}

		return &r.admin, nil

	}
	return &r.admin, nil
}

func (r *repositoryRegistry) RegisterAdminUsersRepository(servicePrefix string) (*administratorUsersRepo.AdminUsersRepository, error) {
	if r.adminUsers == nil {
		var err error
		r.logger.Debug("admin users repository is not registered, registering it now")
		r.adminUsers, err = administratorUsersRepo.NewGORMAdminUsersRepository(r.db, servicePrefix, r.config, r.builders)

		if err != nil {
			r.logger.Fatal("service repository creation error:", zap.Error(err))
		}

		if r.adminUsers == nil {
			return nil, errors.New("failed to initialize admin users repository")
		}

		return &r.adminUsers, nil

	}
	return &r.adminUsers, nil
}

func (r *repositoryRegistry) RegisterAdminUserRolesRepository(servicePrefix string) (*administratorUserRolesRepo.AdminUserRolesRepository, error) {
	if r.adminRoles == nil {
		var err error
		r.logger.Debug("admin user roles repository is not registered, registering it now")
		r.adminRoles, err = administratorUserRolesRepo.NewGORMAdminUserRolesRepository(r.db, servicePrefix, r.config, r.builders)

		if err != nil {
			r.logger.Fatal("service repository creation error:", zap.Error(err))
		}

		if r.adminRoles == nil {
			return nil, errors.New("failed to initialize admin user roles repository")
		}

		return &r.adminRoles, nil

	}
	return &r.adminRoles, nil
}

func (r *repositoryRegistry) RegisterAdminAuthRepository(servicePrefix string, contextMessage contextMessages.ContextMessages) (*administratorAuthRepo.AdminAuthRepository, error) {
	if r.adminAuth == nil {
		var err error
		r.logger.Debug("admin auth repository is not registered, registering it now")
		r.adminAuth, err = administratorAuthRepo.NewAuthRepository(r.db, servicePrefix, r.config, r.builders, r.redis, contextMessage)

		if err != nil {
			r.logger.Fatal("service repository creation error:", zap.Error(err))
		}

		if r.adminAuth == nil {
			return nil, errors.New("failed to initialize admin auth repository")
		}

		return &r.adminAuth, nil

	}
	return &r.adminAuth, nil
}

func (r *repositoryRegistry) RegisterAdminLogsRepository(servicePrefix string) (*administratorLogsRepo.AdminLogsRepository, error) {
	if r.adminLogs == nil {
		var err error
		r.logger.Debug("admin logs repository is not registered, registering it now")
		r.adminLogs, err = administratorLogsRepo.NewGORMAdminLogs(r.db, servicePrefix, r.config, r.builders)

		if err != nil {
			r.logger.Fatal("service repository creation error:", zap.Error(err))
		}

		if r.adminLogs == nil {
			return nil, errors.New("failed to initialize admin logs repository")
		}

		return &r.adminLogs, nil

	}
	return &r.adminLogs, nil
}

func (r *repositoryRegistry) RegisterAdminPolicyRepository(servicePrefix string) (*administratorPolicyRepo.AdminRolePolicyRepository, error) {
	if r.adminPolicy == nil {
		var err error
		r.logger.Debug("admin policy repository is not registered, registering it now")
		r.adminPolicy, err = administratorPolicyRepo.NewGORMAdminPolicyRepository(r.db, servicePrefix, r.config, r.builders)

		if err != nil {
			r.logger.Fatal("service repository creation error:", zap.Error(err))
		}

		if r.adminLogs == nil {
			return nil, errors.New("failed to initialize admin logs repository")
		}

		return &r.adminPolicy, nil

	}
	return &r.adminPolicy, nil
}

/* sub-services ------------------------------------------------------------------------------------------------- */

func (r *repositoryRegistry) RegisterOrdersRepository(servicePrefix string) (*orders.OrdersRepository, error) {
	if r.orders == nil {
		var err error
		r.logger.Debug("orders repository is not registered, registering it now")
		r.orders, err = orders.NewGORMOrdersRepository(r.db, servicePrefix, r.config, r.builders)

		if err != nil {
			r.logger.Fatal("service repository creation error:", zap.Error(err))
		}

		if r.orders == nil {
			return nil, errors.New("failed to initialize orders repository")
		}

		return &r.orders, nil

	}
	return &r.orders, nil
}

func (r *repositoryRegistry) RegisterTransactionsRepository(servicePrefix string) (*transactions.TransactionRepository, error) {
	if r.transactions == nil {
		var err error
		r.logger.Debug("Transactions repository is not registered, registering it now")
		r.transactions, err = transactions.NewGORMTransactionsRepository(r.db, servicePrefix, r.config, r.builders)

		if err != nil {
			r.logger.Fatal("service repository creation error:", zap.Error(err))
		}

		if r.transactions == nil {
			return nil, errors.New("received nil transactions repository")
		}

		return &r.transactions, nil
	}
	return &r.transactions, nil
}

func (r *repositoryRegistry) RegisterUsersRepository(servicePrefix string) (*users.UsersRepository, error) {
	if r.users == nil {
		var err error
		r.logger.Debug("Users repository is not registered, registering it now")
		r.users, err = users.NewGORMUsersRepository(r.db, servicePrefix, r.config, r.builders)

		if err != nil {
			r.logger.Fatal("service repository creation error:", zap.Error(err))
		}

		if r.users == nil {
			return nil, errors.New("received nil users repository")
		}

		return &r.users, nil
	}
	return &r.users, nil
}

func (r *repositoryRegistry) RegisterAssetsRepository(servicePrefix string) (*assets.AssetsRepository, error) {
	if r.assets == nil {
		var err error
		r.logger.Debug("assets repository is not registered, registering it now")
		r.assets, err = assets.NewGORMAssetsRepository(r.db, servicePrefix, r.config, r.builders)

		if err != nil {
			r.logger.Fatal("service repository creation error:", zap.Error(err))
		}

		if r.assets == nil {
			return nil, errors.New("received nil assets repository")
		}

		return &r.assets, nil
	}
	return &r.assets, nil
}
