package router

import (
	"fmt"

	"github.com/denizumutdereli/stream-admin/internal/middleware"
	"github.com/gin-gonic/gin"
)

/* direct middlewares ------------------------------------------------------------------------------ */

func (rc *routerController) paginationMiddleware() gin.HandlerFunc {
	paginationMiddleware := middleware.Pagination()
	return paginationMiddleware
}

func (rc *routerController) ipLimiterMiddleware() middleware.IpController {
	ipControllerMiddleware := middleware.NewIPController(rc.config, rc.adminUserService())
	return ipControllerMiddleware
}

func (rc *routerController) rateLimiterMiddleware() gin.HandlerFunc {
	fmt.Println("rate.... --->")
	rateimiterMiddleware := middleware.NewRateLimiter(rc.config)
	return rateimiterMiddleware.RateLimitMiddleware()
}

func (rc *routerController) guardMiddleware() gin.HandlerFunc {
	fmt.Println("guarding...")
	guardMiddleware := middleware.NewGuardMiddleware(rc.config, rc.authRepository(), rc.contextMessageService())
	return guardMiddleware.Guard()
}

func (rc *routerController) sessionMiddleware() middleware.SessionMiddleware {
	sessionMiddleware := middleware.NewSessionMiddleware(rc.config, rc.redis, rc.adminUserService(), rc.contextMessageService())
	return sessionMiddleware
}

/* child middlewares ------------------------------------------------------------------------------- */

func (rc *routerController) ipRangeAllowedMiddleware() gin.HandlerFunc {
	return rc.ipLimiterMiddleware().IsRangeOfIPAllowed()
}

func (rc *routerController) ipAllowedMiddleware() gin.HandlerFunc {
	return rc.ipLimiterMiddleware().IPAllowedMiddleware()
}

func (rc *routerController) checkUserLock() gin.HandlerFunc {
	return rc.sessionMiddleware().CheckUserLock()
}

func (rc *routerController) optGuardMiddleware() gin.HandlerFunc {
	return rc.sessionMiddleware().LimitOTPAttempts()
}

func (rc *routerController) notDeleteOwnUser() gin.HandlerFunc {
	return rc.sessionMiddleware().NotDeleteOwnUser()
}

/* middleware utilities ---------------------------------------------------------------------------- */

func (rc *routerController) attachMiddlewaresDirect(middlewareFuncs ...func() gin.HandlerFunc) []gin.HandlerFunc {
	var attachedMiddlewares []gin.HandlerFunc
	for _, middlewareFunc := range middlewareFuncs {
		attachedMiddlewares = append(attachedMiddlewares, middlewareFunc())
	}
	return attachedMiddlewares
}

func (rc *routerController) attachMiddlewaresToGroup(group *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	for _, middleware := range middlewares {
		group.Use(middleware)
	}
}
