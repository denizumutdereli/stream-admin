package router

import (
	"github.com/denizumutdereli/stream-admin/internal/comm/message"
	"github.com/denizumutdereli/stream-admin/internal/repository/administrator/auth"
	"github.com/denizumutdereli/stream-admin/internal/service/administrator/logs"
	"github.com/denizumutdereli/stream-admin/internal/service/administrator/roles"
	"github.com/denizumutdereli/stream-admin/internal/service/administrator/users"
	"go.uber.org/zap"
)

/* service interface ---------------------------------------------- */
func (rc *routerController) adminUserService() users.AdminUserService {
	adminUserService, err := rc.services.GetAdminUsersService()

	if err != nil {
		rc.logger.Error("error getting admin users service", zap.Error(err))
	}

	if adminUserService == nil {
		rc.logger.Error("no admin users service found", zap.Error(err))
	}
	return adminUserService
}

func (rc *routerController) adminRoleService() roles.AdminUserRolesService {
	adminRoleService, err := rc.services.GetAdminUserRolesService()

	if err != nil {
		rc.logger.Error("error getting admin roles service", zap.Error(err))
	}

	if adminRoleService == nil {
		rc.logger.Error("no admin roles service found", zap.Error(err))
	}
	return adminRoleService
}

func (rc *routerController) adminLogService() logs.AdminLogsService {
	adminLogs, err := rc.services.GetAdminLogsService()
	if err != nil {
		rc.logger.Error("error getting admin logs service", zap.Error(err))
	}

	if adminLogs == nil {
		rc.logger.Error("no admin logs service found", zap.Error(err))
	}
	return adminLogs
}

func (rc *routerController) contextMessageService() message.ContextMessages {
	contextMessagesService, err := rc.services.GetAdminContextMessageService()

	if err != nil {
		rc.logger.Error("Error getting contextual messages service for DI", zap.Error(err))
	}
	return contextMessagesService
}

/* repository interface ---------------------------------------------- */
func (rc *routerController) authRepository() auth.AdminAuthRepository {
	authRepository, err := rc.repos.GetAdminAuthRepository()

	if err != nil {
		rc.logger.Error("Error getting admin auth repository on guard middleware for DI", zap.Error(err))
	}

	return authRepository
}
