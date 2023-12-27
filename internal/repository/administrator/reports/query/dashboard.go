package query

import (
	"context"

	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator/reports"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type QueryRepository interface {
	GetAll(paginationParams *types.PaginationParams, searchParams *models.AdministratorDashboardQuery) (*database.PaginatedResult, error)
	Create(template *models.AdministratorDashboardQuery) error
	GetByID(id uint) (*models.AdministratorDashboardQuery, error)
	Update(template *models.AdministratorDashboardQuery) error
	Delete(id uint) error
}

type repoConfig struct {
	QueriesTable string
}

type queryRepository struct {
	ctx        context.Context
	cancel     context.CancelFunc
	database   *gorm.DB
	repoConfig *repoConfig
	logger     *zap.Logger
}

func NewGORMQueryRepository(database *gorm.DB, servicePrefix string, logger *zap.Logger) (QueryRepository, error) {
	database.AutoMigrate(&models.AdministratorDashboardQuery{})
	repoConfig := &repoConfig{
		QueriesTable: servicePrefix + "_dashboard_queries"}

	repository := &queryRepository{database: database, repoConfig: repoConfig, logger: logger}
	ctx, cancel := context.WithCancel(context.Background())
	repository.ctx = ctx
	repository.cancel = cancel

	return repository, nil
}

func (r *queryRepository) GetAll(paginationParams *types.PaginationParams, searchParams *models.AdministratorDashboardQuery) (*database.PaginatedResult, error) {
	var data []*models.AdministratorDashboardQuery
	var count int64

	db := r.database.Debug().Table(r.repoConfig.QueriesTable)

	query := db.Scopes()

	countQuery := r.database.Table(r.repoConfig.QueriesTable).Where("deleted_at IS NULL").Scopes()

	if err := countQuery.Count(&count).Error; err != nil {
		r.logger.Error("error counting data:", zap.Error(err))
		return nil, err
	}

	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query = query.Offset(offset).Limit(paginationParams.Limit)

	if err := query.Find(&data).Error; err != nil {
		return nil, err
	}

	paginatedResults := database.PaginateTheResults(data, count, offset, paginationParams.Page, paginationParams.Limit)

	return paginatedResults, nil
}

func (r *queryRepository) Create(template *models.AdministratorDashboardQuery) error {
	return r.database.Table(r.repoConfig.QueriesTable).Create(template).Error
}

func (r *queryRepository) GetByID(id uint) (*models.AdministratorDashboardQuery, error) {
	var query models.AdministratorDashboardQuery
	if err := r.database.Table(r.repoConfig.QueriesTable).Where("id = ?", id).First(&query).Error; err != nil {
		return nil, err
	}
	return &query, nil
}

func (r *queryRepository) Update(query *models.AdministratorDashboardQuery) error {
	return r.database.Table(r.repoConfig.QueriesTable).Save(query).Error
}

func (r *queryRepository) Delete(id uint) error {
	return r.database.Table(r.repoConfig.QueriesTable).Where("id = ?", id).Delete(&models.AdministratorDashboardQuery{}).Error
}
