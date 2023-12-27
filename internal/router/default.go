package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (rc *routerController) setupDefaultRoutes(defaultGroup *gin.RouterGroup) {
	defaultRoutes := []RouteDefinition{
		{
			Method: http.MethodGet,
			Path:   "/",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Welcome to the " + rc.config.AppName})
			},
		},
		{
			Method: http.MethodOptions,
			Path:   "/*options",
			HandlerFunc: func(c *gin.Context) {
				c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // TODO from config
				c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type")
				c.Status(http.StatusNoContent)
			},
		},
	}

	rc.registerRoutesToGroup(defaultGroup, defaultRoutes)

	rc.router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Method or route not found in: " + rc.config.AppName})
	})
}
