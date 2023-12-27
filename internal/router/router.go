package router

import (
	"fmt"

	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/registry"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RouteInfo struct {
	Method  string
	Path    string
	Handler string
}

type RouteDefinition struct {
	Method      string
	Path        string
	HandlerFunc gin.HandlerFunc
	Middlewares []gin.HandlerFunc
}

type RouterController interface {
	GetRouter() *gin.Engine
	GetRoutes() []RouteInfo
	newRouteGroup(groupPrefix, groupName string, middlewares ...gin.HandlerFunc) *gin.RouterGroup
	registerGroup(group *gin.RouterGroup, parentGroup *gin.RouterGroup)
	getGroup(name string) *gin.RouterGroup

	registerRoutesToGroup(group *gin.RouterGroup, routes []RouteDefinition)

	// RegisterMiddlewareDirect(name string, middleware gin.HandlerFunc)
}

type routerController struct {
	config      *config.Config
	logger      *zap.Logger
	redis       *transport.RedisManager
	router      *gin.Engine
	routes      []RouteInfo
	handlers    registry.HandlersRegistry
	services    registry.ServiceRegistry
	repos       registry.RepositoryRegistry
	groups      map[string]*gin.RouterGroup
	groupNames  map[string]string
	middlewares map[string]gin.HandlerFunc
}

func NewRouterController(cfg *config.Config, redis *transport.RedisManager, handlers registry.HandlersRegistry, services registry.ServiceRegistry, repos registry.RepositoryRegistry) RouterController {
	rc := &routerController{
		config:      cfg,
		logger:      cfg.Logger,
		redis:       redis,
		router:      gin.Default(),
		handlers:    handlers,
		services:    services,
		repos:       repos,
		groups:      make(map[string]*gin.RouterGroup),
		groupNames:  make(map[string]string),
		middlewares: make(map[string]gin.HandlerFunc),
	}

	// Configure CORS middleware
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.AllowedOrigins
	corsConfig.AllowAllOrigins = cfg.AllowAllOrigins
	corsConfig.AllowMethods = cfg.AllowedRestMethods
	corsConfig.AllowHeaders = cfg.AllowedRestHeaders

	rc.router.Use(cors.New(corsConfig))

	gin.ForceConsoleColor()
	//router settings

	rc.router.SetTrustedProxies(nil)
	rc.router.RemoveExtraSlash = true
	rc.router.RedirectTrailingSlash = true

	rc.router.Use(
		rc.rateLimiterMiddleware(),
		rc.ipRangeAllowedMiddleware(),
		rc.ipAllowedMiddleware(),
	)

	utils.ClearScreen()
	rc.setupDefaultInterface()
	rc.setupAdminInterface()

	// register routes for furher using
	rc.registerRoutes()

	fmt.Println(rc.GetRoutes(), "--------->>>")

	return rc
}

func (rc *routerController) setupDefaultInterface() {
	defaultRouteGroup := rc.router.Group("/")

	rc.setupDefaultRoutes(defaultRouteGroup)
	rc.adminAuthRoutes(defaultRouteGroup)
	rc.setupSuperAdminRoutes(defaultRouteGroup)
}

func (rc *routerController) setupAdminInterface() {
	adminGroup := rc.router.Group("/admin")

	rc.attachMiddlewaresToGroup(adminGroup, rc.guardMiddleware(), rc.checkUserLock(), rc.isSuperAdmin())

	// administrator interface
	rc.adminServiceRoutes(adminGroup)
	rc.setupAdminRolesRoutes(adminGroup)
	rc.setupAdminUsersRoutes(adminGroup)
	rc.setupAdminLogsRoutes(adminGroup)
	rc.setupAdminPolicyRoutes(adminGroup)

	servicesGroup := rc.router.Group("/service")

	rc.attachMiddlewaresToGroup(servicesGroup, rc.guardMiddleware(), rc.checkUserLock(), rc.isDefaultUser())

	// service interface
	rc.serviceOrdersRoutes(servicesGroup)
	rc.serviceUsersRoutes(servicesGroup)
	rc.serviceKYCRoutes(servicesGroup)
	rc.serviceTransactionRoutes(servicesGroup)
	rc.serviceAssetsRoutes(servicesGroup)
}
