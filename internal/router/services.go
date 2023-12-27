package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (rc *routerController) serviceOrdersRoutes(servicesGroup *gin.RouterGroup) {

	serviceHandler, err := rc.handlers.GetOrdersRestHandler()
	if err != nil {
		rc.logger.Error("unable to get service handler", zap.Error(err))
		return
	}

	if serviceHandler == nil {
		rc.logger.Error("service handler is nil", zap.Error(err))
		return
	}

	serviceGroup := servicesGroup.Group("/orders")

	routes := []RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: serviceHandler.GetAll,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware),
		},
		{
			Method:      http.MethodGet,
			Path:        "/params",
			HandlerFunc: serviceHandler.GetAll, // TODO: params func
		},
	}
	rc.registerRoutesToGroup(serviceGroup, routes)
	rc.registerGroup(serviceGroup, servicesGroup)

}

func (rc *routerController) serviceUsersRoutes(servicesGroup *gin.RouterGroup) {

	serviceHandler, err := rc.handlers.GetUsersRestHandler()
	if err != nil {
		rc.logger.Error("unable to get service handler", zap.Error(err))
		return
	}

	if serviceHandler == nil {
		rc.logger.Error("service handler is nil", zap.Error(err))
		return
	}

	serviceGroup := servicesGroup.Group("/users")
	routes := []RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: serviceHandler.GetUsers,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware),
		},
		// {
		// 	Method:      http.MethodGet,
		// 	Path:        "/detail/:user_id",
		// 	HandlerFunc: serviceHandler.GetUserDetailsBuilder,
		// 	Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware),
		// },
		{
			Method:      http.MethodGet,
			Path:        "/params",
			HandlerFunc: serviceHandler.GetSearchUserParameters,
		},
	}
	rc.registerRoutesToGroup(serviceGroup, routes)
	rc.registerGroup(serviceGroup, servicesGroup)

}

func (rc *routerController) serviceKYCRoutes(servicesGroup *gin.RouterGroup) {

	serviceHandler, err := rc.handlers.GetUsersRestHandler()
	if err != nil {
		rc.logger.Error("unable to get service handler", zap.Error(err))
		return
	}

	if serviceHandler == nil {
		rc.logger.Error("service handler is nil", zap.Error(err))
		return
	}

	serviceGroup := servicesGroup.Group("/kyc")
	routes := []RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: serviceHandler.GetKYC,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware),
		},
		{
			Method:      http.MethodGet,
			Path:        "/params",
			HandlerFunc: serviceHandler.GetSearchKYCParameters,
		},
	}
	rc.registerRoutesToGroup(serviceGroup, routes)
	rc.registerGroup(serviceGroup, servicesGroup)

}

func (rc *routerController) serviceTransactionRoutes(servicesGroup *gin.RouterGroup) {

	serviceHandler, err := rc.handlers.GetTransactionsRestHandler()
	if err != nil {
		rc.logger.Error("unable to get service handler", zap.Error(err))
		return
	}

	if serviceHandler == nil {
		rc.logger.Error("service handler is nil", zap.Error(err))
		return
	}

	serviceGroup := servicesGroup.Group("/transactions")
	routes := []RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "/fiat",
			HandlerFunc: serviceHandler.GetFiatTransactions,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware),
		},
		{
			Method:      http.MethodGet,
			Path:        "/crypto",
			HandlerFunc: serviceHandler.GetCryptoTransactions,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware),
		},
		{
			Method:      http.MethodGet,
			Path:        "/wallets",
			HandlerFunc: serviceHandler.GetCryptoWallets,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware),
		},
		{
			Method:      http.MethodGet,
			Path:        "/params",
			HandlerFunc: serviceHandler.GetSearchParameters,
		},
	}
	rc.registerRoutesToGroup(serviceGroup, routes)
	rc.registerGroup(serviceGroup, servicesGroup)

}

func (rc *routerController) serviceAssetsRoutes(servicesGroup *gin.RouterGroup) {

	serviceHandler, err := rc.handlers.GetAssetsRestHandler()
	if err != nil {
		rc.logger.Error("unable to get service handler", zap.Error(err))
		return
	}

	if serviceHandler == nil {
		rc.logger.Error("service handler is nil", zap.Error(err))
		return
	}

	serviceGroup := servicesGroup.Group("/assets")
	routes := []RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "/coins",
			HandlerFunc: serviceHandler.GetCoins,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware),
		},
		{
			Method:      http.MethodGet,
			Path:        "/assets",
			HandlerFunc: serviceHandler.GetAssets,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware),
		},
		{
			Method:      http.MethodGet,
			Path:        "/networks",
			HandlerFunc: serviceHandler.GetNetworks,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware),
		},
		{
			Method:      http.MethodGet,
			Path:        "/params",
			HandlerFunc: serviceHandler.GetSearchParameters,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware),
		},
	}
	rc.registerRoutesToGroup(serviceGroup, routes)
	rc.registerGroup(serviceGroup, servicesGroup)

}
