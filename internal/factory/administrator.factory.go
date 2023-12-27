package factory

import (
	"context"

	administratorAuthHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/auth"
	administratorLogsHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/logs"
	administratorPolicyHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/policy"
	administratorUserRolesHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/roles"
	administratorUserHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/user"

	"github.com/denizumutdereli/stream-admin/internal/handler"
	"go.uber.org/zap"
)

func (f *serviceFactory) NewAdminLogsService(ctx context.Context) (*administratorLogsHandler.AdminLogsRestHandler, error) {

	serviceName := "admin-action-monitoring"
	servicePrefix, exists := f.config.PrefixService.GetServicePrefix(serviceName)
	if !exists {
		f.logger.Fatal("No prefix found for service:", zap.String("serviceName", serviceName))
	}

	repo, err := f.registry.repos.RegisterAdminLogsRepository(servicePrefix)
	if err != nil {
		f.logger.Fatal("service repository error:", zap.Error(err))
		return nil, err
	}

	service, err := f.registry.services.RegisterAdminLogsService(repo)
	if err != nil {
		f.logger.Fatal("service registry error:", zap.Error(err))
		return nil, err
	}

	handler, err := f.registry.handlers.RegisterAdminLogsHandler(service)
	if err != nil {
		f.logger.Error("Failed to register and get admin auth handler")
		return nil, err
	}

	return handler, nil
}

func (f *serviceFactory) NewAdminAuthService(ctx context.Context) (*administratorAuthHandler.AdminAuthHandler, error) {
	serviceName := "admin-auth"
	servicePrefix, exists := f.config.PrefixService.GetServicePrefix(serviceName)
	if !exists {
		f.logger.Fatal("No prefix found for service:", zap.String("serviceName", serviceName))
	}

	adminAuthrepo, err := f.registry.repos.RegisterAdminAuthRepository(servicePrefix, f.contextMessages)
	if err != nil {
		f.logger.Fatal("service repository error:", zap.Error(err))
		return nil, err
	}

	adminUsersRepo, err := f.registry.repos.RegisterAdminUsersRepository(servicePrefix)
	if err != nil {
		f.logger.Fatal("service repository error:", zap.Error(err))
		return nil, err
	}

	service, err := f.registry.services.RegisterAdminAuthService(adminAuthrepo, adminUsersRepo, f.caesar, f.redis, f.config)
	if err != nil {
		f.logger.Fatal("service registry error:", zap.Error(err))
		return nil, err
	}

	handler, err := f.registry.handlers.RegisterAdminAuthHandler(&service)
	if err != nil {
		f.logger.Error("Failed to register and get admin auth handler")
		return nil, err
	}

	return handler, nil
}

func (f *serviceFactory) NewAdminUserService(ctx context.Context) (*administratorUserHandler.AdminUserHandler, error) {
	serviceName := "admin-users"
	servicePrefix, exists := f.config.PrefixService.GetServicePrefix(serviceName)
	if !exists {
		f.logger.Fatal("No prefix found for service:", zap.String("serviceName", serviceName))
	}

	repo, err := f.registry.repos.RegisterAdminUsersRepository(servicePrefix)
	if err != nil {
		f.logger.Fatal("service repository error:", zap.Error(err))
		return nil, err
	}

	service, err := f.registry.services.RegisterAdminUsersService(repo, f.caesar, f.config)
	if err != nil {
		f.logger.Fatal("service registry error:", zap.Error(err))
		return nil, err
	}

	handler, err := f.registry.handlers.RegisterAdminUsersHandler(&service)
	if err != nil {
		f.logger.Error("Failed to register and get admin users handler")
		return nil, err
	}

	return handler, nil
}

func (f *serviceFactory) NewAdminUserRolesService(ctx context.Context) (*administratorUserRolesHandler.AdminUserRolesHandler, error) {
	serviceName := "admin-user-roles"
	servicePrefix, exists := f.config.PrefixService.GetServicePrefix(serviceName)
	if !exists {
		f.logger.Fatal("No prefix found for service:", zap.String("serviceName", serviceName))
	}

	repo, err := f.registry.repos.RegisterAdminUserRolesRepository(servicePrefix)
	if err != nil {
		f.logger.Fatal("service repository error:", zap.Error(err))
		return nil, err
	}

	service, err := f.registry.services.RegisterAdminUserRolesService(repo, f.caesar, f.config)
	if err != nil {
		f.logger.Fatal("service registry error:", zap.Error(err))
		return nil, err
	}

	handler, err := f.registry.handlers.RegisterAdminUserRolesHandler(&service)
	if err != nil {
		f.logger.Error("Failed to register and get admin user roles handler")
		return nil, err
	}

	return handler, nil
}

func (f *serviceFactory) NewAdminPolicyService(ctx context.Context) (*administratorPolicyHandler.AdminPolicyHandler, error) {
	serviceName := "admin-policy"
	servicePrefix, exists := f.config.PrefixService.GetServicePrefix(serviceName)
	if !exists {
		f.logger.Fatal("No prefix found for service:", zap.String("serviceName", serviceName))
	}

	repo, err := f.registry.repos.RegisterAdminPolicyRepository(servicePrefix)
	if err != nil {
		f.logger.Fatal("service repository error:", zap.Error(err))
		return nil, err
	}

	service, err := f.registry.services.RegisterAdminPolicyService(repo, f.caesar, f.config)
	if err != nil {
		f.logger.Fatal("service registry error:", zap.Error(err))
		return nil, err
	}

	handler, err := f.registry.handlers.RegisterAdminPolicyHandler(&service)
	if err != nil {
		f.logger.Error("Failed to register and get admin users handler")
		return nil, err
	}

	return handler, nil
}

func (f *serviceFactory) NewAdminServiceFactory(ctx context.Context) (*handler.AdminRestHandler, error) {

	repo, err := f.registry.repos.RegisterAdminRepository()
	if err != nil {
		f.logger.Fatal("service repository error:", zap.Error(err))
		return nil, err
	}

	service, err := f.registry.services.RegisterAdminService(repo)
	if err != nil {
		f.logger.Fatal("service registry error:", zap.Error(err))
		return nil, err
	}

	go func() {
		for leaderStatus := range f.config.IsLeader {
			isInstanceLeader := leaderStatus
			if leaderStatus {
				f.logger.Info("I am now the leader :)", zap.Bool("isLeader", isInstanceLeader))
				service.SetIsLeader(ctx, true)
			} else {
				service.SetIsLeader(ctx, false)
				f.logger.Warn("I lost my leadership :(", zap.Bool("isLeader", isInstanceLeader))
			}
		}
	}()

	handler, err := f.registry.handlers.RegisterAdminRestHandler(service)
	if err != nil {
		f.logger.Error("Failed to register and get admin rest handler")
		return nil, err
	}

	return handler, nil
}
