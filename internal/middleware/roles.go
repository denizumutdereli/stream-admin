package middleware

import (
	"fmt"
	"net/http"

	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"
	administratorLogsService "github.com/denizumutdereli/stream-admin/internal/service/administrator/logs"
	"github.com/denizumutdereli/stream-admin/internal/service/administrator/roles"
	"github.com/denizumutdereli/stream-admin/internal/utils"

	"github.com/denizumutdereli/stream-admin/internal/config"

	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RoleCallbackFunction func(c *gin.Context, claims *types.AccessTokenClaims)

type RolesMiddleware interface {
	AuthorizeRole(allowedRoles ...string) gin.HandlerFunc
	GetRoleInfo(roleID string) (*models.AdministratorRole, error)

	defaultCallback(c *gin.Context, claims *types.AccessTokenClaims)
	unauthorizedCallback(c *gin.Context, claims *types.AccessTokenClaims)
	superAdminCallback(c *gin.Context, claims *types.AccessTokenClaims)
	getAccessTokenClaims(c *gin.Context) (*types.AccessTokenClaims, error)
}

type rolesMiddleware struct {
	config         *config.Config
	logger         *zap.Logger
	rolesCallbacks map[string]RoleCallbackFunction
	adminLogger    administratorLogsService.AdminLogsService
	roleService    roles.AdminUserRolesService
}

func NewRolesMiddleware(config *config.Config, adminLogger administratorLogsService.AdminLogsService, roleService roles.AdminUserRolesService) RolesMiddleware {

	rolesMidleware := &rolesMiddleware{config: config, logger: config.Logger, adminLogger: adminLogger, roleService: roleService}
	rolesMidleware.rolesCallbacks = map[string]RoleCallbackFunction{
		"superAdmin": rolesMidleware.superAdminCallback,
		"user":       rolesMidleware.defaultCallback,
	}

	return rolesMidleware
}

func (r *rolesMiddleware) GetRoleInfo(roleID string) (*models.AdministratorRole, error) {
	userRole, err := r.roleService.GetAdminRoleByID(roleID)
	if err != nil {
		return &models.AdministratorRole{}, fmt.Errorf("user role not found in context roleID: %s", roleID)
	}

	return &userRole, nil
}

func (r *rolesMiddleware) AuthorizeRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := r.getAccessTokenClaims(c)
		if err != nil {
			r.logger.Error("error getting access token claims from context", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}

		// Get user role info
		activeUserRole, err := r.GetRoleInfo(claims.UserRole)

		if err != nil {
			r.logger.Error("error getting user role info. Passing now.")
			r.unauthorizedCallback(c, claims)
			return
		}

		utils.ClearScreen()
		fmt.Println("-----", claims.UserRole, "-----------------------------------------------------\n", activeUserRole)
		if activeUserRole.RoleName == string(models.SuperAdmin) { // TODO: fixed id layer
			r.superAdminCallback(c, claims)
			c.Next()
			return
		}

		isAllowed := false
		for _, allowedRole := range allowedRoles {
			if claims.UserRole == allowedRole {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			r.unauthorizedCallback(c, claims)
			return
		}

		if callback, found := r.rolesCallbacks[claims.UserRole]; found {
			callback(c, claims)
		} else {
			r.logger.Warn("no callback defined for role", zap.String("role", claims.UserRole), zap.String("user_id", claims.UserID))
			r.defaultCallback(c, claims)
		}

		// r.unauthorizedCallback(c, claims)
	}
}

func (r *rolesMiddleware) defaultCallback(c *gin.Context, claims *types.AccessTokenClaims) {
	r.logger.Info("Default user action taken", zap.String("user", claims.UserID), zap.String("action", c.Request.RequestURI))
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You do not have the necessary permissions to access this resource. Contact your system administrator if you believe this is an error."})
}

func (r *rolesMiddleware) unauthorizedCallback(c *gin.Context, claims *types.AccessTokenClaims) {
	r.logger.Info("Unauthorized user action taken", zap.String("user", claims.UserID), zap.String("action", c.Request.RequestURI))
	r.adminLogger.LogAction(c, claims.UserRole, claims.UserID, 1)
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You do not have the necessary permissions to access this resource. Contact your system administrator if you believe this is an error."})
}

func (r *rolesMiddleware) superAdminCallback(c *gin.Context, claims *types.AccessTokenClaims) {
	r.logger.Info("Super Admin action taken", zap.String("user", claims.UserID), zap.String("action", c.Request.RequestURI))
	//r.logger.Debug("inserting log" + "-----------------------------------------------")
	r.adminLogger.LogAction(c, claims.UserRole, claims.UserID, 0)
}

func (r *rolesMiddleware) getAccessTokenClaims(c *gin.Context) (*types.AccessTokenClaims, error) {
	userID, exists := c.Get(string(types.ContextUserIDKey))
	if !exists {
		return nil, fmt.Errorf("user ID not found in context")
	}

	fmt.Println(userID)

	role, exists := c.Get(string(types.ContextRoleKey))
	if !exists {
		return nil, fmt.Errorf("role not found in context")
	}

	userAgent, exists := c.Get(string(types.ContextUserAgent))
	if !exists {
		return nil, fmt.Errorf("user agent not found in context")
	}

	return &types.AccessTokenClaims{
		UserID:    userID.(string),
		UserRole:  role.(string),
		UserAgent: userAgent.(string),
	}, nil
}
