package users

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"

	"github.com/denizumutdereli/stream-admin/internal/outbox"
	"github.com/denizumutdereli/stream-admin/internal/repository/scopes"
	"github.com/denizumutdereli/stream-admin/internal/types"

	"github.com/twinj/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminUsersRepository interface {
	CreateInitialSuperAdmin() error

	CreateAdminUser(admin_user *models.AdministratorUser) (bool, error)
	UpdateAdminUser(adminRolePolicy *models.AdministratorUser) (*models.AdministratorUser, error)
	DeleteAdminUser(policyID string) (bool, error)
	GetAdminUsers(paginationParams *types.PaginationParams, searchParams *models.AdministratorUserSearch) (*database.PaginatedResult, error)

	FindAdminUserByAdminUsername(username string) (models.AdministratorUser, error)
	GetAdminUserByVerificationCode(verificationCode string) (models.AdministratorUser, error)
	FindAdminUserByID(userid string) (models.AdministratorUser, error)
	GetAdminActiveUsersVPNAddresses() ([]string, error)
}

type AdministratorUsersOutboxMessage struct {
	outbox.OutboxMessage
}

type repoConfig struct {
	ServicePrefix   string
	AdminUsersTable string
}

type adminUserRepository struct {
	ctx      context.Context
	cancel   context.CancelFunc
	database *gorm.DB
	config   *config.Config
	logger   *zap.Logger

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

func NewGORMAdminUsersRepository(database *gorm.DB, servicePrefix string, config *config.Config, builders builders.BuilderService) (AdminUsersRepository, error) {
	database.AutoMigrate(&models.AdministratorUser{}, &models.AdministratorRole{}, &AdministratorUsersOutboxMessage{})
	repoConfig := &repoConfig{
		ServicePrefix:   servicePrefix,
		AdminUsersTable: servicePrefix + "_users",
	}

	err := config.PrefixService.RegisterServiceTables(servicePrefix, []string{repoConfig.AdminUsersTable})
	if err != nil {
		return nil, err
	}

	repository := &adminUserRepository{
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
		repoConfig.AdminUsersTable+"_outbox_messages")

	repository.outboxManager = *outbox

	go func() {
		repository.outboxManager.ProcessMessages()
	}()

	return repository, nil
}

func (u *adminUserRepository) CreateInitialSuperAdmin() error {
	var adminUser models.AdministratorUser
	result := u.database.Where("username = ?", "spadmin").First(&adminUser)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		superAdminRole := models.AdministratorRole{
			RoleID:   uuid.NewV4().String(),
			RoleName: "SuperAdmin",
		}
		result = u.database.FirstOrCreate(&superAdminRole, superAdminRole)

		if result.Error != nil {
			return result.Error
		}

		adminUser = models.AdministratorUser{
			UserID:       uuid.NewV4().String(),
			Username:     "spadmin",
			EmployerName: "Deniz Umut Dereli",
			Password:     string(hashedPassword),
			PhoneNumber:  "+905363925261",
			LastLogin:    time.Now(),
			UserRole:     superAdminRole.RoleID,
			Status:       models.UserStatusVerified,
			VpnAddr:      "127.0.0.1", // TODO: replace on production server
		}

		result = u.database.Create(&adminUser)
		if result.Error != nil {
			return result.Error
		}
	} else if result.Error != nil {
		return result.Error
	}

	return nil
}

func (u *adminUserRepository) CreateAdminUser(adminUser *models.AdministratorUser) (bool, error) {
	var role models.AdministratorRole
	roleResult := u.database.First(&role, "role_id = ?", adminUser.UserRole)
	if errors.Is(roleResult.Error, gorm.ErrRecordNotFound) {
		return true, errors.New("role does not exist")
	} else if roleResult.Error != nil {
		return true, roleResult.Error
	}

	var existingUser models.AdministratorUser
	userResult := u.database.Where("employer_name = ? OR username = ?", adminUser.EmployerName, adminUser.Username).First(&existingUser)
	if userResult.Error == nil {
		return false, errors.New("user already exists with the given employer name or username")
	} else if !errors.Is(userResult.Error, gorm.ErrRecordNotFound) {
		return true, userResult.Error
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return true, err
	}

	adminUser.Password = string(hashedPassword)

	// if existingUser.UserRole == string(models.SuperAdmin) {
	// 	return false, errors.New("super admin users can not be modified or added")
	// }

	createResult := u.database.Create(adminUser)
	if createResult.Error != nil {
		return true, createResult.Error
	}

	return true, nil
}

func (u *adminUserRepository) UpdateAdminUser(adminUser *models.AdministratorUser) (*models.AdministratorUser, error) {
	if adminUser.UserID == "" {
		return nil, errors.New("missing admin user ID")
	}

	var duplicateAdminUser *models.AdministratorUser
	result := u.database.Where("username = ? AND user_id <> ?", adminUser.Username, adminUser.UserID).First(&duplicateAdminUser)
	if result.Error == nil {
		return nil, errors.New("username already in use")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	var existingUser models.AdministratorUser
	result = u.database.First(&existingUser, "user_id = ?", adminUser.UserID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("admin user not found")
	} else if result.Error != nil {
		return nil, result.Error
	}

	// extra layers and confirmation when an admin user modified. Turning to unverified (except for superadmin)

	// if existingUser.UserRole == string(models.SuperAdmin) {
	// 	return nil, errors.New("super admin users can not be modified or added")
	// }

	result = u.database.Model(&existingUser).Preload("Role").Updates(adminUser)
	if result.Error != nil {
		return nil, result.Error
	}

	response := &models.AdministratorUser{
		UserID:       existingUser.UserID,
		EmployerName: existingUser.EmployerName,
		Username:     existingUser.Username,
		PhoneNumber:  existingUser.PhoneNumber,
		VpnAddr:      existingUser.VpnAddr,
		UserRole:     existingUser.UserRole,
		Role:         existingUser.Role,
		Status:       existingUser.Status,
		CreatedAt:    existingUser.CreatedAt,
		UpdatedAt:    existingUser.UpdatedAt,
	}

	return response, nil
}

func (u *adminUserRepository) DeleteAdminUser(userID string) (bool, error) {
	if userID == "" {
		return false, errors.New("missing admin user ID")
	}

	var existingUser models.AdministratorUser
	result := u.database.First(&existingUser, "user_id = ?", userID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, errors.New("admin user not found")
	} else if result.Error != nil {
		return false, result.Error
	}

	// TODO: soft delete hooks

	result = u.database.Delete(&existingUser)
	if result.Error != nil {
		return false, result.Error
	}

	return true, nil
}

func (u *adminUserRepository) GetAdminUsers(paginationParams *types.PaginationParams, searchParams *models.AdministratorUserSearch) (*database.PaginatedResult, error) {
	var data []*models.AdministratorUser
	var count int64

	db := u.database.Debug().Table(u.repoConfig.AdminUsersTable).
		Preload("Role")

	whereScope := scopes.ApplySearchFilters(searchParams, u.repoConfig.AdminUsersTable, u.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := u.database.Table(u.repoConfig.AdminUsersTable).Scopes(
		whereScope,
	)

	if err := countQuery.Count(&count).Error; err != nil {
		u.logger.Error("error counting data:", zap.Error(err))
	}

	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query = query.Offset(offset).Limit(paginationParams.Limit)

	if err := query.Find(&data).Error; err != nil {
		return nil, err
	}

	paginatedResults := database.PaginateTheResults(data, count, offset, paginationParams.Page, paginationParams.Limit)

	return paginatedResults, nil
}

func (u *adminUserRepository) FindAdminUserByAdminUsername(username string) (models.AdministratorUser, error) {
	var admin_user models.AdministratorUser
	result := u.database.Debug().Where("username = ?", username).First(&admin_user)
	return admin_user, result.Error
}

func (u *adminUserRepository) GetAdminUserByVerificationCode(verificationCode string) (models.AdministratorUser, error) {
	var admin_user models.AdministratorUser
	result := u.database.Where("verification_code = ?", verificationCode).First(&admin_user)
	return admin_user, result.Error
}

func (u *adminUserRepository) FindAdminUserByID(userid string) (models.AdministratorUser, error) {
	var admin_user models.AdministratorUser
	result := u.database.Where("user_id = ?", userid).First(&admin_user)
	return admin_user, result.Error
}

func (u *adminUserRepository) GetAdminActiveUsersVPNAddresses() ([]string, error) {
	u.vpnAddrsCache.RLock()
	cached, exists := u.vpnAddrsCache.data["vpnAddresses"]
	if exists && time.Since(cached.fetchedAt) < time.Duration(u.config.DefaultCacheQueryTimeInSeconds)*time.Second {
		u.vpnAddrsCache.RUnlock()
		return cached.vpnAddresses, nil
	}
	u.vpnAddrsCache.RUnlock()

	var adminUsers []models.AdministratorUser
	result := u.database.Where("status = ?", "verified").Find(&adminUsers)
	if result.Error != nil {
		return nil, result.Error
	}

	var vpnAddrs []string
	for _, user := range adminUsers {
		vpnAddrs = append(vpnAddrs, user.VpnAddr)
	}

	u.vpnAddrsCache.Lock()
	u.vpnAddrsCache.data["vpnAddresses"] = struct {
		vpnAddresses []string
		fetchedAt    time.Time
	}{
		vpnAddresses: vpnAddrs,
		fetchedAt:    time.Now(),
	}
	u.vpnAddrsCache.Unlock()

	return vpnAddrs, nil
}
