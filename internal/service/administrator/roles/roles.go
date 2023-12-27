package roles

import (
	"context"
	"net/http"

	"github.com/denizumutdereli/stream-admin/internal/caesar"
	appErrors "github.com/denizumutdereli/stream-admin/internal/common"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"
	roles "github.com/denizumutdereli/stream-admin/internal/repository/administrator/roles"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
)

type AdminUserRolesService interface {
	CreateAdminRole(ctx context.Context, admin_role *models.AdministratorRole) (*models.AdministratorRole, appErrors.Error)
	GetAdminRoles(paginationParams *types.PaginationParams, queryParams *models.AdministratorRoleSearch) (*database.PaginatedResult, appErrors.Error)
	AttachPoliciesToRole(ctx context.Context, roleID string, policyIDs []string) (*models.RoleWithPolicyTitles, appErrors.Error)
	GetAdminRoleByID(roleID string) (models.AdministratorRole, appErrors.Error)
}

type adminUserRolesService struct {
	ctx       context.Context
	cancel    context.CancelFunc
	rolesRepo roles.AdminUserRolesRepository
	config    *config.Config
	logger    *zap.Logger
	caesar    caesar.CaesarManager
}

func NewAdminUserRolesService(rolesRepo *roles.AdminUserRolesRepository, caesar caesar.CaesarManager, config *config.Config) AdminUserRolesService {
	service := &adminUserRolesService{
		rolesRepo: *rolesRepo,
		config:    config,
		logger:    config.Logger,
		caesar:    caesar,
	}

	ctx, cancel := context.WithCancel(context.Background())
	service.ctx = ctx
	service.cancel = cancel

	return service
}

func (s *adminUserRolesService) CreateAdminRole(ctx context.Context, adminRole *models.AdministratorRole) (*models.AdministratorRole, appErrors.Error) {
	isNew, err := s.rolesRepo.CreateAdminRole(adminRole)

	if !isNew {
		return nil, appErrors.AppError(http.StatusConflict, "", "role already exists", nil)
	}

	if err != nil {
		s.logger.Error("error in creating or fetching admin role", zap.Error(err))
		return nil, appErrors.AppError(http.StatusInternalServerError, "", "error in role processing", err)
	}

	return adminRole, nil
}

func (s *adminUserRolesService) GetAdminRoles(paginationParams *types.PaginationParams, queryParams *models.AdministratorRoleSearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.rolesRepo.GetAdminRoles(paginationParams, queryParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}

func (s *adminUserRolesService) AttachPoliciesToRole(ctx context.Context, roleID string, policyIDs []string) (*models.RoleWithPolicyTitles, appErrors.Error) {
	data, err := s.rolesRepo.AttachPoliciesToRole(roleID, policyIDs)
	if err != nil {
		s.logger.Error("error attaching policies to role", zap.Error(err))
		return nil, appErrors.AppError(http.StatusInternalServerError, "", "error attaching policies to role", err)
	}
	return data, nil
}

func (s *adminUserRolesService) GetAdminRoleByID(roleID string) (models.AdministratorRole, appErrors.Error) {
	data, err := s.rolesRepo.GetAdminRoleByID(roleID)
	if err != nil {
		return models.AdministratorRole{}, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}
