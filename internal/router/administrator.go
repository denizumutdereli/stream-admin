package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (rc *routerController) adminAuthRoutes(defaultGroup *gin.RouterGroup) {
	serviceHandler, err := rc.handlers.GetAdminAuthHandler()
	if err != nil {
		rc.logger.Error("unable to get admin handler", zap.Error(err))
		return
	}

	if serviceHandler == nil {
		rc.logger.Error("service adminAuth handler is nil", zap.Error(err))
		return
	}

	routes := []RouteDefinition{
		{
			Method:      http.MethodPost,
			Path:        "/auth",
			HandlerFunc: serviceHandler.Login,
		},
		{
			Method:      http.MethodPost,
			Path:        "/verify",
			HandlerFunc: serviceHandler.VerifyOTPAndLogin,
			Middlewares: rc.attachMiddlewaresDirect(rc.optGuardMiddleware),
		},
		{
			Method:      http.MethodPost,
			Path:        "/refresh",
			HandlerFunc: serviceHandler.RefreshToken,
		},
		{
			Method:      http.MethodPost,
			Path:        "/logout",
			HandlerFunc: serviceHandler.Logout,
			Middlewares: rc.attachMiddlewaresDirect(rc.guardMiddleware),
		},
	}

	rc.registerRoutesToGroup(defaultGroup, routes)
}

func (rc *routerController) adminServiceRoutes(adminGroup *gin.RouterGroup) {
	serviceHandler, err := rc.handlers.GetAdminRestHandler()
	if err != nil {
		rc.logger.Error("unable to get admin handler", zap.Error(err))
		return
	}

	if serviceHandler == nil {
		rc.logger.Error("service adminService handler is nil", zap.Error(err))
		return
	}

	routes := []RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "/live",
			HandlerFunc: serviceHandler.Live,
		},
		{
			Method:      http.MethodGet,
			Path:        "/read",
			HandlerFunc: serviceHandler.Read,
		},
		{
			Method:      http.MethodGet,
			Path:        "/metrics",
			HandlerFunc: serviceHandler.Metrics,
			Middlewares: rc.attachMiddlewaresDirect(rc.isSuperAdmin),
		},
		{
			Method:      http.MethodGet,
			Path:        "/configs",
			HandlerFunc: serviceHandler.Configs,
		},
	}

	rc.registerRoutesToGroup(adminGroup, routes)
}

func (rc *routerController) setupSuperAdminRoutes(defaultGroup *gin.RouterGroup) {
	// TODO: secure with some other way
	superGroup := defaultGroup.Group("/superxyz")
	serviceHandler, err := rc.handlers.GetAdminUsersHandler()
	if err != nil {
		rc.logger.Error("unable to get admin users handler", zap.Error(err))
		return
	}

	if serviceHandler == nil {
		rc.logger.Error("service adminUsers handler is nil", zap.Error(err))
		return
	}

	routes := []RouteDefinition{
		{
			Method:      http.MethodPost,
			Path:        "/superadmin",
			HandlerFunc: serviceHandler.CreateSuperAdmin,
			//Middlewares: rc.attachMiddlewaresDirect(rc.setupGuardMiddleware), // TODO: fix placeholder middlewares
		},
	}

	rc.registerRoutesToGroup(superGroup, routes)
}

func (rc *routerController) setupAdminRolesRoutes(adminGroup *gin.RouterGroup) {

	serviceHandler, err := rc.handlers.GetAdminUserRolesHandler()
	if err != nil {
		rc.logger.Error("unable to get admin user roles handler", zap.Error(err))
		return
	}

	if serviceHandler == nil {
		rc.logger.Error("service adminUsers handler is nil", zap.Error(err))
		return
	}

	adminUsers := adminGroup.Group("/roles")

	routes := []RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "/",
			HandlerFunc: serviceHandler.GetAdminRoles,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware, rc.isSuperAdmin),
		},
		{
			Method:      http.MethodPost,
			Path:        "/create",
			HandlerFunc: serviceHandler.CreateAdminRole,
			Middlewares: rc.attachMiddlewaresDirect(rc.isSuperAdmin),
		},
		{
			Method:      http.MethodPut,
			Path:        "/attach",
			HandlerFunc: serviceHandler.AttachPoliciesToRole,
			Middlewares: rc.attachMiddlewaresDirect(rc.isSuperAdmin),
		},
	}

	rc.registerRoutesToGroup(adminUsers, routes)
}

func (rc *routerController) setupAdminPolicyRoutes(adminGroup *gin.RouterGroup) {

	serviceHandler, err := rc.handlers.GetAdminPolicyHandler()
	if err != nil {
		rc.logger.Error("unable to get admin policy handler", zap.Error(err))
		return
	}

	if serviceHandler == nil {
		rc.logger.Error("service adminPolicy handler is nil", zap.Error(err))
		return
	}

	adminPolicy := adminGroup.Group("/policy")

	routes := []RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "/",
			HandlerFunc: serviceHandler.GetAdminRolePolicies,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware, rc.isSuperAdmin),
		},
		{
			Method:      http.MethodPost,
			Path:        "/create",
			HandlerFunc: serviceHandler.CreateAdminRolePolicy,
			Middlewares: rc.attachMiddlewaresDirect(rc.isSuperAdmin),
		},
		{
			Method:      http.MethodPut,
			Path:        "/update",
			HandlerFunc: serviceHandler.UpdateAdminRolePolicy,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware, rc.isSuperAdmin),
		},
		{
			Method:      http.MethodDelete,
			Path:        "/delete/:policy_id",
			HandlerFunc: serviceHandler.DeleteAdminRolePolicy,
			Middlewares: rc.attachMiddlewaresDirect(rc.isSuperAdmin),
		},
	}

	rc.registerRoutesToGroup(adminPolicy, routes)
}

func (rc *routerController) setupAdminUsersRoutes(adminGroup *gin.RouterGroup) {
	serviceHandler, err := rc.handlers.GetAdminUsersHandler()
	if err != nil {
		rc.logger.Error("unable to get admin users handler", zap.Error(err))
		return
	}

	if serviceHandler == nil {
		rc.logger.Error("service adminUsers handler is nil", zap.Error(err))
		return
	}

	adminUsers := adminGroup.Group("/users")

	routes := []RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "/",
			HandlerFunc: serviceHandler.GetAdminUsers,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware, rc.isSuperAdmin),
		},
		{
			Method:      http.MethodPost,
			Path:        "/verify",
			HandlerFunc: serviceHandler.VerifyAdminUser,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware, rc.optGuardMiddleware, rc.isSuperAdmin),
		},
		{
			Method:      http.MethodPost,
			Path:        "/create",
			HandlerFunc: serviceHandler.CreateAdminUser,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware, rc.isSuperAdmin),
		},
		{
			Method:      http.MethodPut,
			Path:        "/update",
			HandlerFunc: serviceHandler.UpdateAdminUser,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware, rc.isSuperAdmin),
		},
		{
			Method:      http.MethodDelete,
			Path:        "/delete/:user_id",
			HandlerFunc: serviceHandler.DeleteAdminUser,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware, rc.notDeleteOwnUser, rc.isSuperAdmin),
		},
	}

	rc.registerRoutesToGroup(adminUsers, routes)
}

func (rc *routerController) setupAdminLogsRoutes(adminGroup *gin.RouterGroup) {

	serviceHandler, err := rc.handlers.GetAdminLogsHandler()
	if err != nil {
		rc.logger.Error("unable to get admin logs handler", zap.Error(err))
		return
	}

	if serviceHandler == nil {
		rc.logger.Error("service adminLogs handler is nil", zap.Error(err))
		return
	}

	routes := []RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "/logs",
			HandlerFunc: serviceHandler.GetAll,
			Middlewares: rc.attachMiddlewaresDirect(rc.paginationMiddleware, rc.isSuperAdmin),
		},
	}

	rc.registerRoutesToGroup(adminGroup, routes)
}
