package factory

import (
	"context"
	"fmt"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/caesar"
	contextMessage "github.com/denizumutdereli/stream-admin/internal/comm/message"
	"github.com/denizumutdereli/stream-admin/internal/config"

	"github.com/denizumutdereli/stream-admin/internal/handler"
	administratorAuthHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/auth"
	administratorLogsHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/logs"
	administratorPolicyHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/policy"
	administratorUserRolesHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/roles"
	administratorUserHandler "github.com/denizumutdereli/stream-admin/internal/handler/administrator/user"
	"github.com/denizumutdereli/stream-admin/internal/registry"
	"github.com/denizumutdereli/stream-admin/internal/service/administrator/stream"
	"github.com/denizumutdereli/stream-admin/internal/wsserver"
	"gorm.io/gorm"

	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
)

/*
	Administrator streams (service names includes *stream) are proxy of user end data for showing on admin panel
*/

type ServiceFactory interface {
	BuildTransports(redis, nats, kafka bool) []error

	NewKafkaManager() (transport.KafkaManager, error)
	NewRedisManager() (*transport.RedisManager, error)
	NewNatsManager() error

	FRedis() *transport.RedisManager
	FKafka() *transport.KafkaManager
	FNats() *transport.NatsManager

	Services() registry.ServiceRegistry
	Handlers() registry.HandlersRegistry
	Repos() registry.RepositoryRegistry

	NewAdminServiceFactory(ctx context.Context) (*handler.AdminRestHandler, error)
	NewAdminLogsService(ctx context.Context) (*administratorLogsHandler.AdminLogsRestHandler, error)
	NewAdminAuthService(ctx context.Context) (*administratorAuthHandler.AdminAuthHandler, error)
	NewAdminUserService(ctx context.Context) (*administratorUserHandler.AdminUserHandler, error)
	NewAdminUserRolesService(ctx context.Context) (*administratorUserRolesHandler.AdminUserRolesHandler, error)
	NewAdminPolicyService(ctx context.Context) (*administratorPolicyHandler.AdminPolicyHandler, error)
	NewAdminContextMessageService(ctx context.Context) (contextMessage.ContextMessages, error)

	NewStreamAssetsService() (*stream.AssetsService, error)

	NewOrdersService(ctx context.Context) (*handler.OrdersRestHandler, error)
	NewUsersService(ctx context.Context) (*handler.UsersRestHandler, error)
	NewTransactionsService(ctx context.Context) (*handler.TransactionsRestHandler, error)
	NewAssetsService(ctx context.Context) (*handler.AssetsRestHandler, error)
	NewWsServer() *wsserver.Server
}

type serviceRegistry struct {
	repos    registry.RepositoryRegistry
	services registry.ServiceRegistry
	handlers registry.HandlersRegistry
}

type serviceFactory struct {
	channels []string
	database *gorm.DB
	config   *config.Config
	logger   *zap.Logger
	//adminActionLogger   *administratorLogsService.AdminLogsService
	nats                *transport.NatsManager
	redis               *transport.RedisManager
	streamAssetsManager *stream.AssetsService
	streamAssets        []string
	appContext          *types.ExchangeConfig
	builders            builders.BuilderService
	caesar              caesar.CaesarManager
	kafka               transport.KafkaManager
	contextMessages     contextMessage.ContextMessages
	wsserver            *wsserver.Server
	serverReady         chan struct{}
	registry            serviceRegistry
}

func NewServiceFactory(config *config.Config) (ServiceFactory, error) {
	factory := &serviceFactory{
		config:      config,
		logger:      config.Logger,
		channels:    config.Channels,
		serverReady: make(chan struct{}),
	}

	var err error

	transportBuildErrors := factory.BuildTransports(true, true, true)
	if len(transportBuildErrors) > 0 {
		for _, e := range transportBuildErrors {
			factory.logger.Error(e.Error(), zap.Error(e))
		}
		return nil, fmt.Errorf("failed to initialize transport layers")
	}

	factory.database, err = NewDatabaseFactory(config, factory.logger).CreateCitusDB()
	if err != nil {
		factory.logger.Fatal("Failed to create citus database connection", zap.Error(err))
		return nil, err
	}

	factory.builders = builders.NewBuilder(factory.config)
	factory.caesar = caesar.NewCaesarManager(factory.redis)

	// administrator stream sub-services
	factory.streamAssetsManager, err = factory.NewStreamAssetsService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AssetsService: %w", err)
	}

	factory.streamAssets, err = factory.streamAssetsManager.GetAssets()
	if err != nil {
		return nil, fmt.Errorf("failed to get initial assets: %w", err)
	}

	factory.appContext = &types.ExchangeConfig{
		Redis:        factory.redis,
		Config:       factory.config,
		Logger:       factory.logger,
		DB:           factory.database,
		Caesar:       &factory.caesar,
		Channels:     factory.channels,
		Nats:         factory.nats,
		StreamAssets: factory.streamAssets,
	}

	/* lazy loading for repositories  & handlers----------------------------------------------------------*/
	repositoryRegistry, err := registry.NewRepositoryRegistry(factory.database, factory.config, factory.builders, factory.redis)
	if err != nil {
		factory.logger.Fatal("Failed to create repository registry", zap.Error(err))
		return nil, err
	}
	factory.registry.repos = repositoryRegistry

	serviceRegistry, err := registry.NewServiceRegistry(factory.appContext)
	if err != nil {
		factory.logger.Fatal("Failed to create service registry", zap.Error(err))
		return nil, err
	}

	factory.registry.services = serviceRegistry

	handlersRegistry, err := registry.NewHandlersRegistry(factory.config, factory.builders)
	if err != nil {
		factory.logger.Fatal("Failed to create handlers registry", zap.Error(err))
		return nil, err
	}
	factory.registry.handlers = handlersRegistry

	/* ----------------------------------------------------------------------------------------------------*/

	wsServer := factory.NewWsServer()

	go func() {
		factory.logger.Info("WebSocket Server is starting ", zap.String("ws-port", config.WsServerPort))
		wsServer.Serve(config.WsServerPort)
		close(factory.serverReady)
	}()

	return factory, nil
}

func (f *serviceFactory) FRedis() *transport.RedisManager {
	return f.redis
}

func (f *serviceFactory) FKafka() *transport.KafkaManager {
	return &f.kafka
}

func (f *serviceFactory) FNats() *transport.NatsManager {
	return f.nats
}

func (f *serviceFactory) Services() registry.ServiceRegistry {
	return f.registry.services
}

func (f *serviceFactory) Handlers() registry.HandlersRegistry {
	return f.registry.handlers
}

func (f *serviceFactory) Repos() registry.RepositoryRegistry {
	return f.registry.repos
}
