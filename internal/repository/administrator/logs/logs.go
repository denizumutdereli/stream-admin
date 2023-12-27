package logs

import (
	"context"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"
	"github.com/denizumutdereli/stream-admin/internal/repository/scopes"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RepoConfig struct {
	ServicePrefix string
	LogsTable     string
}

type AdminLogsRepository interface {
	Create(adminlog *models.AdministratorLogs) error
	GetAll(paginationParams *types.PaginationParams, searchParams *models.AdministratorLogsSearch) (*database.PaginatedResult, error)
}

type adminLogsRepository struct {
	ctx              context.Context
	cancel           context.CancelFunc
	database         *gorm.DB
	repoConfig       *RepoConfig
	logger           *zap.Logger
	builders         builders.BuilderService
	dslSearchEnabled bool
}

func NewGORMAdminLogs(database *gorm.DB, servicePrefix string, config *config.Config, builders builders.BuilderService) (AdminLogsRepository, error) {
	database.AutoMigrate(&models.AdministratorLogs{})
	repoConfig := &RepoConfig{
		ServicePrefix: servicePrefix,
		LogsTable:     servicePrefix + "_logs"}

	err := config.PrefixService.RegisterServiceTables(servicePrefix, []string{repoConfig.LogsTable})
	if err != nil {
		return nil, err
	}

	repository := &adminLogsRepository{database: database, repoConfig: repoConfig, logger: config.Logger, builders: builders, dslSearchEnabled: true}
	ctx, cancel := context.WithCancel(context.Background())
	repository.ctx = ctx
	repository.cancel = cancel

	return repository, nil
}

func (z *adminLogsRepository) Create(adminlog *models.AdministratorLogs) error {
	return z.database.Create(adminlog).Error
}

func (z *adminLogsRepository) GetAll(paginationParams *types.PaginationParams, searchParams *models.AdministratorLogsSearch) (*database.PaginatedResult, error) {
	var adminLogs []*models.AdministratorLogs
	var count int64

	db := z.database.Debug().Table(z.repoConfig.LogsTable)

	whereScope := scopes.ApplySearchFilters(searchParams, z.repoConfig.LogsTable, z.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := z.database.Table(z.repoConfig.LogsTable).Where("deleted_at IS NULL").Scopes(whereScope)

	if err := countQuery.Count(&count).Error; err != nil {
		z.logger.Error("error counting admin user logs:", zap.Error(err))
	}

	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query = query.Offset(offset).Limit(paginationParams.Limit)

	if err := query.Find(&adminLogs).Error; err != nil {
		return nil, err
	}

	paginatedResults := database.PaginateTheResults(adminLogs, count, offset, paginationParams.Page, paginationParams.Limit)

	return paginatedResults, nil
}
