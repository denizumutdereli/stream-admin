package roles

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"
	rolePolicyModels "github.com/denizumutdereli/stream-admin/internal/models/administrator/policy"

	"github.com/denizumutdereli/stream-admin/internal/outbox"
	"github.com/denizumutdereli/stream-admin/internal/repository/scopes"
	"github.com/denizumutdereli/stream-admin/internal/types"

	"github.com/twinj/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AdminUserRolesRepository interface {
	CreateAdminRole(adminRole *models.AdministratorRole) (bool, error)
	GetAdminRoles(paginationParams *types.PaginationParams, searchParams *models.AdministratorRoleSearch) (*database.PaginatedResult, error)
	AttachPoliciesToRole(roleID string, policyIDs []string) (*models.RoleWithPolicyTitles, error)
	GetAdminRoleByID(roleID string) (models.AdministratorRole, error)
}

type AdministratorRolesOutboxMessage struct {
	outbox.OutboxMessage
}

type repoConfig struct {
	ServicePrefix       string
	AdminUserRolesTable string
}

type adminUserRolesRepository struct {
	ctx              context.Context
	cancel           context.CancelFunc
	database         *gorm.DB
	config           *config.Config
	logger           *zap.Logger
	repoConfig       *repoConfig
	builders         builders.BuilderService
	dslSearchEnabled bool
	vpnAddrsCache    struct {
		sync.RWMutex
		data map[string]struct {
			vpnAddresses []string
			fetchedAt    time.Time
		}
	}
	outboxManager outbox.OutboxManager
}

func NewGORMAdminUserRolesRepository(database *gorm.DB, servicePrefix string, config *config.Config, builders builders.BuilderService) (AdminUserRolesRepository, error) {
	database.AutoMigrate(&models.AdministratorRole{}, &AdministratorRolesOutboxMessage{})
	repoConfig := &repoConfig{
		ServicePrefix:       servicePrefix,
		AdminUserRolesTable: servicePrefix + "_roles"}

	err := config.PrefixService.RegisterServiceTables(servicePrefix, []string{repoConfig.AdminUserRolesTable})
	if err != nil {
		return nil, err
	}

	repository := &adminUserRolesRepository{
		config:           config,
		logger:           config.Logger,
		database:         database,
		repoConfig:       repoConfig,
		builders:         builders,
		dslSearchEnabled: true,
		vpnAddrsCache: struct {
			sync.RWMutex
			data map[string]struct {
				vpnAddresses []string
				fetchedAt    time.Time
			}
		}{data: make(map[string]struct {
			vpnAddresses []string
			fetchedAt    time.Time
		})},
	}

	ctx, cancel := context.WithCancel(context.Background())
	repository.ctx = ctx
	repository.cancel = cancel

	outbox := outbox.NewOutboxManager(
		repository.config,
		database, outbox.DefaultDispatcherSettings(),
		repoConfig.AdminUserRolesTable+"_outbox_messages")

	repository.outboxManager = *outbox

	go func() {
		repository.outboxManager.ProcessMessages()
	}()

	return repository, nil
}

func (u *adminUserRolesRepository) CreateAdminRole(adminRole *models.AdministratorRole) (bool, error) {
	var existingRole models.AdministratorRole
	result := u.database.Where("role_name = ?", adminRole.RoleName).First(&existingRole)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		adminRole.RoleID = uuid.NewV4().String()
		result = u.database.Create(adminRole)
		if result.Error != nil {
			return false, result.Error
		}
		return true, nil
	} else if result.Error != nil {
		return false, result.Error
	}

	*adminRole = existingRole
	return false, nil
}

func (u *adminUserRolesRepository) GetAdminRoles(paginationParams *types.PaginationParams, searchParams *models.AdministratorRoleSearch) (*database.PaginatedResult, error) {
	var data []*models.AdministratorRole
	var count int64

	db := u.database.Debug().Table(u.repoConfig.AdminUserRolesTable)

	whereScope := scopes.ApplySearchFilters(searchParams, u.repoConfig.AdminUserRolesTable, u.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := u.database.Table(u.repoConfig.AdminUserRolesTable).Scopes(
		whereScope,
	)

	if err := countQuery.Count(&count).Error; err != nil {
		u.logger.Error("error counting data:", zap.Error(err))
	}

	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query = query.Offset(offset).Limit(paginationParams.Limit)

	if err := query.Preload("Policies").Find(&data).Error; err != nil {
		return nil, err
	}

	paginatedResults := database.PaginateTheResults(data, count, offset, paginationParams.Page, paginationParams.Limit)

	return paginatedResults, nil

}

func (u *adminUserRolesRepository) AttachPoliciesToRole(roleID string, policyIDs []string) (*models.RoleWithPolicyTitles, error) {
	numPolicies := len(policyIDs)
	if numPolicies < 1 {
		return nil, errors.New("at least one policy is required")
	}
	if numPolicies > 10 {
		return nil, errors.New("cannot attach more than 10 policies")
	}

	var role models.AdministratorRole
	if err := u.database.Preload("Policies").First(&role, "role_id = ?", roleID).Error; err != nil {
		return nil, errors.New("role not found")
	}

	var policies []*rolePolicyModels.AdministratorRolePolicy
	if err := u.database.Where("policy_id IN ?", policyIDs).Find(&policies).Error; err != nil {
		return nil, err
	}

	if len(policies) != numPolicies {
		return nil, errors.New("one or more policies not found")
	}

	if err := u.database.Model(&role).Association("Policies").Replace(policies); err != nil {
		return nil, errors.New("error while attaching policies: " + err.Error())
	}

	var policyTitles []models.PolicyTitle
	if err := u.database.Model(&rolePolicyModels.AdministratorRolePolicy{}).Where("policy_id IN ?", policyIDs).Select("policy_id", "title").Find(&policyTitles).Error; err != nil {
		return nil, err
	}

	roleWithPolicyTitles := models.RoleWithPolicyTitles{
		RoleID:   role.RoleID,
		RoleName: role.RoleName,
		Policies: policyTitles,
	}

	return &roleWithPolicyTitles, nil
}

func (u *adminUserRolesRepository) GetAdminRoleByID(roleID string) (models.AdministratorRole, error) {
	var admin_role models.AdministratorRole

	result := u.database.Debug().Where("role_id = ?", roleID).First(&admin_role)

	// TODO: add role status
	//if admin_role.Status = models.S

	return admin_role, result.Error
}
