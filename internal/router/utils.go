package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (rc *routerController) GetRouter() *gin.Engine {
	return rc.router
}

func (rc *routerController) GetRoutes() []RouteInfo {
	return rc.routes
}

func (rc *routerController) registerRoutes() {
	var routes []RouteInfo
	for _, route := range rc.router.Routes() {
		routes = append(routes, RouteInfo{
			Method:  route.Method,
			Path:    route.Path,
			Handler: route.Handler,
		})
	}
	rc.routes = routes
}

func (rc *routerController) newRouteGroup(groupPrefix, groupName string, middlewares ...gin.HandlerFunc) *gin.RouterGroup {
	var group *gin.RouterGroup
	if existingGroup, exists := rc.groups[groupPrefix]; exists {
		group = existingGroup
	} else {
		group = rc.router.Group(groupPrefix)
		// Attach middlewares if provided
		for _, middleware := range middlewares {
			group.Use(middleware)
		}
		rc.groups[groupPrefix] = group
		rc.groupNames[groupName] = groupPrefix
	}
	return group
}

func (rc *routerController) registerGroup(group *gin.RouterGroup, parentGroup *gin.RouterGroup) {
	if parentGroup != nil {
		nestedGroup := parentGroup.Group(group.BasePath())
		*group = *nestedGroup
	} else {
		rc.router.Group(group.BasePath())
	}
}

func (rc *routerController) getGroup(groupName string) *gin.RouterGroup {
	if groupPrefix, exists := rc.groupNames[groupName]; exists {
		return rc.groups[groupPrefix]
	}
	rc.logger.Error("group name not found", zap.String("group", groupName))
	return nil
}

func (rc *routerController) registerRoutesToGroup(group *gin.RouterGroup, routes []RouteDefinition) {
	methodToHandler := map[string]func(string, ...gin.HandlerFunc) gin.IRoutes{
		http.MethodGet:     group.GET,
		http.MethodPost:    group.POST,
		http.MethodPut:     group.PUT,
		http.MethodDelete:  group.DELETE,
		http.MethodOptions: group.OPTIONS,
	}

	for _, route := range routes {
		handlers := append(route.Middlewares, route.HandlerFunc)
		if handlerFunc, ok := methodToHandler[route.Method]; ok {
			handlerFunc(route.Path, handlers...)
		} else {
			rc.logger.Error("HTTP method not supported", zap.String("method", route.Method))
		}
	}
}
