package policy

import (
	"context"
	"net/http"

	"github.com/denizumutdereli/stream-admin/internal/caesar"
	appErrors "github.com/denizumutdereli/stream-admin/internal/common"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	rolePolicyModels "github.com/denizumutdereli/stream-admin/internal/models/administrator/policy"

	policyRepo "github.com/denizumutdereli/stream-admin/internal/repository/administrator/policy"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
)

type AdminPolicyService interface {
	// admin role policies
	CreateAdminRolePolicy(ctx context.Context, admin_role *rolePolicyModels.AdministratorRolePolicy) (*rolePolicyModels.AdministratorRolePolicy, appErrors.Error)
	UpdateAdminRolePolicy(ctx context.Context, adminRolePolicy *rolePolicyModels.AdministratorRolePolicy) (*rolePolicyModels.AdministratorRolePolicyResponse, appErrors.Error)
	DeleteAdminRolePolicy(ctx context.Context, policyID string) appErrors.Error
	GetAdminRolePolicies(paginationParams *types.PaginationParams, queryParams *rolePolicyModels.AdministratorRolePolicySearch) (*database.PaginatedResult, appErrors.Error)
}

type adminPolicyService struct {
	ctx    context.Context
	cancel context.CancelFunc
	repo   policyRepo.AdminRolePolicyRepository
	config *config.Config
	logger *zap.Logger
	caesar caesar.CaesarManager
}

func NewAdminPolicyService(repo *policyRepo.AdminRolePolicyRepository, caesar caesar.CaesarManager, config *config.Config) AdminPolicyService {
	service := &adminPolicyService{
		repo:   *repo,
		config: config,
		logger: config.Logger,
		caesar: caesar,
	}

	ctx, cancel := context.WithCancel(context.Background())
	service.ctx = ctx
	service.cancel = cancel

	return service
}

/* Admin roles ------------------------------------------------------------------------------------------------------ */

func (s *adminPolicyService) CreateAdminRolePolicy(ctx context.Context, adminRolePolicy *rolePolicyModels.AdministratorRolePolicy) (*rolePolicyModels.AdministratorRolePolicy, appErrors.Error) {

	if err := s.validateRolesSubPolicies(&adminRolePolicy.SubPolicies); err != nil {
		return nil, appErrors.AppError(http.StatusBadRequest, "", "role policy validation has failed", err)
	}

	isNew, err := s.repo.CreateAdminRolePolicy(adminRolePolicy)
	if err != nil {
		s.logger.Error("error in creating or fetching admin role policy", zap.Error(err))
		return nil, appErrors.AppError(http.StatusInternalServerError, "", "error in role policy processing", err)
	}

	if !isNew {
		return nil, appErrors.AppError(http.StatusBadRequest, "", "role policy already exists", nil)
	}

	return adminRolePolicy, nil
}

func (s *adminPolicyService) UpdateAdminRolePolicy(ctx context.Context, adminRolePolicy *rolePolicyModels.AdministratorRolePolicy) (*rolePolicyModels.AdministratorRolePolicyResponse, appErrors.Error) {

	if err := s.validateRolesSubPolicies(&adminRolePolicy.SubPolicies); err != nil {
		return nil, appErrors.AppError(http.StatusBadRequest, "", "role policy validation has failed", err)
	}

	response, err := s.repo.UpdateAdminRolePolicy(adminRolePolicy)
	if err != nil {
		s.logger.Error("error updating admin role policy", zap.Error(err))
		return nil, appErrors.AppError(http.StatusInternalServerError, "", "error updating role policy", err)
	}

	return response, nil
}

func (s *adminPolicyService) DeleteAdminRolePolicy(ctx context.Context, policyID string) appErrors.Error {
	_, err := s.repo.DeleteAdminRolePolicy(policyID)
	if err != nil {
		s.logger.Error("error deleting admin role policy", zap.Error(err))
		return appErrors.AppError(http.StatusInternalServerError, "", "error deleting role policy", err)
	}

	return nil
}

func (s *adminPolicyService) GetAdminRolePolicies(paginationParams *types.PaginationParams, queryParams *rolePolicyModels.AdministratorRolePolicySearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.repo.GetAdminRolePolicies(paginationParams, queryParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}
