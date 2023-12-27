package router

import (
	"fmt"

	"github.com/denizumutdereli/stream-admin/internal/middleware"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"
	"github.com/gin-gonic/gin"
)

func (rc *routerController) rolesMiddleware() middleware.RolesMiddleware {
	rolesMiddleware := middleware.NewRolesMiddleware(rc.config, rc.adminLogService(), rc.adminRoleService())
	return rolesMiddleware
}

func (rc *routerController) isDefaultUser() gin.HandlerFunc {
	return rc.rolesMiddleware().AuthorizeRole(string(models.RegularUser))
}

func (rc *routerController) isSuperAdmin() gin.HandlerFunc {
	fmt.Println(string(models.SuperAdmin), "--->")
	return rc.rolesMiddleware().AuthorizeRole(string(models.SuperAdmin))
}
