package policy

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	rolePolicyModels "github.com/denizumutdereli/stream-admin/internal/models/administrator/policy"
	"github.com/denizumutdereli/stream-admin/internal/repository/scopes"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/twinj/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AdminRolePolicyRepository interface {
	CreateAdminRolePolicy(adminRolePolicy *rolePolicyModels.AdministratorRolePolicy) (bool, error)
	UpdateAdminRolePolicy(adminRolePolicy *rolePolicyModels.AdministratorRolePolicy) (*rolePolicyModels.AdministratorRolePolicyResponse, error)
	DeleteAdminRolePolicy(policyID string) (bool, error)
	GetAdminRolePolicies(paginationParams *types.PaginationParams, searchParams *rolePolicyModels.AdministratorRolePolicySearch) (*database.PaginatedResult, error)
}

type repoConfig struct {
	ServicePrefix        string
	AdminRolePolicyTable string
	AdminRolesTable      string
}

type adminRolePolicyRepository struct {
	ctx              context.Context
	cancel           context.CancelFunc
	database         *gorm.DB
	repoConfig       *repoConfig
	logger           *zap.Logger
	builders         builders.BuilderService
	dslSearchEnabled bool
}

func NewGORMAdminPolicyRepository(database *gorm.DB, servicePrefix string, config *config.Config, builders builders.BuilderService) (AdminRolePolicyRepository, error) {
	database.AutoMigrate(&rolePolicyModels.AdministratorRolePolicy{})
	repoConfig := &repoConfig{
		ServicePrefix:        servicePrefix,
		AdminRolePolicyTable: servicePrefix + "_role_policies",
		AdminRolesTable:      servicePrefix + "_roles",
	}

	err := config.PrefixService.RegisterServiceTables(servicePrefix, []string{repoConfig.AdminRolePolicyTable, repoConfig.AdminRolesTable})
	if err != nil {
		return nil, err
	}

	repository := &adminRolePolicyRepository{database: database, repoConfig: repoConfig, logger: config.Logger, builders: builders, dslSearchEnabled: true}
	ctx, cancel := context.WithCancel(context.Background())
	repository.ctx = ctx
	repository.cancel = cancel

	return repository, nil
}

func (u *adminRolePolicyRepository) CreateAdminRolePolicy(adminRolePolicy *rolePolicyModels.AdministratorRolePolicy) (bool, error) {
	var existingPolicy rolePolicyModels.AdministratorRolePolicy
	result := u.database.Where("title = ?", adminRolePolicy.Title).First(&existingPolicy)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		adminRolePolicy.PolicyID = uuid.NewV4().String()

		var subPolicies []types.SubRolePolicies
		err := json.Unmarshal([]byte(adminRolePolicy.SubPolicies), &subPolicies)
		if err != nil {
			u.logger.Error("error unmarshalling sub policies:", zap.Error(err))
			return false, err
		}
		adminRolePolicy.SubPolicyRules = subPolicies

		result = u.database.Create(adminRolePolicy)
		if result.Error != nil {
			return false, result.Error
		}
		return true, nil
	} else if result.Error != nil {
		return false, result.Error
	}

	return false, nil
}

func (u *adminRolePolicyRepository) UpdateAdminRolePolicy(adminRolePolicy *rolePolicyModels.AdministratorRolePolicy) (*rolePolicyModels.AdministratorRolePolicyResponse, error) {
	if adminRolePolicy.PolicyID == "" {
		return nil, errors.New("missing policy ID")
	}

	var duplicatePolicy rolePolicyModels.AdministratorRolePolicy
	result := u.database.Where("title = ? AND policy_id <> ?", adminRolePolicy.Title, adminRolePolicy.PolicyID).First(&duplicatePolicy)
	if result.Error == nil {
		return nil, errors.New("title already in use by another policy")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	var existingPolicy rolePolicyModels.AdministratorRolePolicy
	result = u.database.First(&existingPolicy, "policy_id = ?", adminRolePolicy.PolicyID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("policy not found")
	} else if result.Error != nil {
		return nil, result.Error
	}

	result = u.database.Model(&existingPolicy).Updates(adminRolePolicy)
	if result.Error != nil {
		return nil, result.Error
	}

	var subPolicies []types.SubRolePolicies
	err := json.Unmarshal([]byte(existingPolicy.SubPolicies), &subPolicies)
	if err != nil {
		u.logger.Error("error unmarshalling sub policies:", zap.Error(err))
		return nil, err
	}

	response := &rolePolicyModels.AdministratorRolePolicyResponse{
		PolicyID:       existingPolicy.PolicyID,
		Readonly:       existingPolicy.Readonly,
		Target:         existingPolicy.Target,
		Title:          existingPolicy.Title,
		SubPolicyRules: subPolicies,
		Status:         existingPolicy.Status,
		CreatedAt:      existingPolicy.CreatedAt,
		UpdatedAt:      existingPolicy.UpdatedAt,
	}

	return response, nil
}

func (u *adminRolePolicyRepository) DeleteAdminRolePolicy(policyID string) (bool, error) {
	if policyID == "" {
		return false, errors.New("missing policy ID")
	}

	var existingPolicy rolePolicyModels.AdministratorRolePolicy
	result := u.database.First(&existingPolicy, "policy_id = ?", policyID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, errors.New("policy not found")
	} else if result.Error != nil {
		return false, result.Error
	}

	result = u.database.Delete(&existingPolicy)
	if result.Error != nil {
		return false, result.Error
	}

	return true, nil
}

func (u *adminRolePolicyRepository) GetAdminRolePolicies(paginationParams *types.PaginationParams, searchParams *rolePolicyModels.AdministratorRolePolicySearch) (*database.PaginatedResult, error) {
	var data []*rolePolicyModels.AdministratorRolePolicy
	var count int64

	db := u.database.Debug().Table(u.repoConfig.AdminRolePolicyTable)

	whereScope := scopes.ApplySearchFilters(searchParams, u.repoConfig.AdminRolePolicyTable, u.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := u.database.Table(u.repoConfig.AdminRolePolicyTable).Where("deleted_at IS NULL").Scopes(whereScope)

	if err := countQuery.Count(&count).Error; err != nil {
		u.logger.Error("error counting data:", zap.Error(err))
	}

	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query = query.Offset(offset).Limit(paginationParams.Limit)

	if err := query.Find(&data).Error; err != nil {
		return nil, err
	}

	var responseData []*rolePolicyModels.AdministratorRolePolicyResponse

	for _, policy := range data {
		var subPolicies []types.SubRolePolicies
		err := json.Unmarshal([]byte(policy.SubPolicies), &subPolicies)
		if err != nil {
			u.logger.Error("error unmarshalling sub policies:", zap.Error(err))
			continue
		}

		responseItem := &rolePolicyModels.AdministratorRolePolicyResponse{
			PolicyID:       policy.PolicyID,
			Readonly:       policy.Readonly,
			Target:         policy.Target,
			Title:          policy.Title,
			SubPolicyRules: subPolicies,
			Status:         policy.Status,
			CreatedAt:      policy.CreatedAt,
			UpdatedAt:      policy.UpdatedAt,
		}
		responseData = append(responseData, responseItem)
	}

	paginatedResults := database.PaginateTheResults(responseData, count, offset, paginationParams.Page, paginationParams.Limit)
	return paginatedResults, nil

}
