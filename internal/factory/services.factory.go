package factory

import (
	"context"

	"github.com/denizumutdereli/stream-admin/internal/handler"
	"go.uber.org/zap"
)

func (f *serviceFactory) NewOrdersService(ctx context.Context) (*handler.OrdersRestHandler, error) {
	serviceName := "orders"
	servicePrefix, exists := f.config.PrefixService.GetServicePrefix(serviceName)
	if !exists {
		f.logger.Fatal("No prefix found for service:", zap.String("serviceName", serviceName))
	}

	repo, err := f.registry.repos.RegisterOrdersRepository(servicePrefix)
	if err != nil {
		f.logger.Fatal("service repository error:", zap.Error(err))
		return nil, err
	}

	service, err := f.registry.services.RegisterOrdersService(repo)
	if err != nil {
		f.logger.Fatal("service registry error:", zap.Error(err))
		return nil, err
	}

	handler, err := f.registry.handlers.RegisterOrdersRestHandler(service)
	if err != nil {
		f.logger.Error("Failed to register and get orders rest handler")
		return nil, err
	}

	return handler, nil
}

func (f *serviceFactory) NewTransactionsService(ctx context.Context) (*handler.TransactionsRestHandler, error) {
	serviceName := "transactions"
	servicePrefix, exists := f.config.PrefixService.GetServicePrefix(serviceName)
	if !exists {
		f.logger.Fatal("No prefix found for service:", zap.String("serviceName", serviceName))
	}

	repo, err := f.registry.repos.RegisterTransactionsRepository(servicePrefix)
	if err != nil {
		f.logger.Fatal("service repository error:", zap.Error(err))
		return nil, err
	}

	service, err := f.registry.services.RegisterTransactionsService(repo)
	if err != nil {
		f.logger.Fatal("service registry error:", zap.Error(err))
		return nil, err
	}

	handler, err := f.registry.handlers.RegisterTransactionRestHandler(service)
	if err != nil {
		f.logger.Error("Failed to register and get txs service rest handler")
		return nil, err
	}

	return handler, nil
}

func (f *serviceFactory) NewUsersService(ctx context.Context) (*handler.UsersRestHandler, error) {
	serviceName := "users"
	servicePrefix, exists := f.config.PrefixService.GetServicePrefix(serviceName)
	if !exists {
		f.logger.Fatal("No prefix found for service:", zap.String("serviceName", serviceName))
	}

	repo, err := f.registry.repos.RegisterUsersRepository(servicePrefix)
	if err != nil {
		f.logger.Fatal("service repository error:", zap.Error(err))
		return nil, err
	}

	service, err := f.registry.services.RegisterUsersService(repo)
	if err != nil {
		f.logger.Fatal("service registry error:", zap.Error(err))
		return nil, err
	}

	handler, err := f.registry.handlers.RegisterUsersHandler(service)
	if err != nil {
		f.logger.Error("Failed to register and get users service rest handler")
		return nil, err
	}

	return handler, nil
}

func (f *serviceFactory) NewAssetsService(ctx context.Context) (*handler.AssetsRestHandler, error) {
	serviceName := "assets"
	servicePrefix, exists := f.config.PrefixService.GetServicePrefix(serviceName)
	if !exists {
		f.logger.Fatal("No prefix found for service:", zap.String("serviceName", serviceName))
	}

	repo, err := f.registry.repos.RegisterAssetsRepository(servicePrefix)
	if err != nil {
		f.logger.Fatal("service repository error:", zap.Error(err))
		return nil, err
	}

	service, err := f.registry.services.RegisterAssetsService(repo)
	if err != nil {
		f.logger.Fatal("service registry error:", zap.Error(err))
		return nil, err
	}

	handler, err := f.registry.handlers.RegisterAssetsRestHandler(service)
	if err != nil {
		f.logger.Error("Failed to register and get users service rest handler")
		return nil, err
	}

	return handler, nil
}
