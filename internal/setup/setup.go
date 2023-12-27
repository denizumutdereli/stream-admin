package setup

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/factory"
	"go.uber.org/zap"
)

func SetupApp(cfg *config.Config, logger *zap.Logger) (factory.ServiceFactory, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serviceFactory, err := factory.NewServiceFactory(cfg)
	if err != nil {
		logger.Fatal("Error creating service factory", zap.Error(err))
		return nil, err
	}

	var wg sync.WaitGroup

	initFunctions := []func(context.Context) error{
		func(ctx context.Context) error {
			_, err := serviceFactory.NewAdminServiceFactory(ctx)
			return err
		},
		func(ctx context.Context) error {
			_, err := serviceFactory.NewAdminAuthService(ctx)
			return err
		},
		func(ctx context.Context) error {
			_, err := serviceFactory.NewAdminUserService(ctx)
			return err
		},
		func(ctx context.Context) error {
			_, err := serviceFactory.NewAdminUserRolesService(ctx)
			return err
		},
		func(ctx context.Context) error {
			_, err := serviceFactory.NewAdminContextMessageService(ctx)
			return err
		},
		func(ctx context.Context) error {
			_, err := serviceFactory.NewAdminLogsService(ctx) // BUG: there is a sequence registration problem. I will look.
			return err
		},
		func(ctx context.Context) error {
			_, err := serviceFactory.NewAdminPolicyService(ctx)
			return err
		},
		func(ctx context.Context) error {
			_, err := serviceFactory.NewOrdersService(ctx)
			return err
		},
		func(ctx context.Context) error {
			_, err := serviceFactory.NewUsersService(ctx)
			return err
		},
		func(ctx context.Context) error {
			_, err := serviceFactory.NewTransactionsService(ctx)
			return err
		},
		func(ctx context.Context) error {
			_, err := serviceFactory.NewAssetsService(ctx)
			return err
		},
	}

	for _, initFunc := range initFunctions {
		wg.Add(1)
		go func(f func(context.Context) error) {
			defer wg.Done()
			if err := f(ctx); err != nil {
				logger.Error("Failed to initialize service", zap.Error(err))
			}
		}(initFunc)
	}

	wg.Wait()

	setupSignalHandling(ctx, cancel, logger)

	return serviceFactory, nil
}

func setupSignalHandling(ctx context.Context, cancel context.CancelFunc, logger *zap.Logger) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

		// cleanup...

		cancel()
	}()
}
