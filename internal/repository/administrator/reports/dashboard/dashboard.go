package dashboard

import (
	"context"

	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator/reports"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetAll(paginationParams *types.PaginationParams, searchParams *models.AdministratorDashboard) (*database.PaginatedResult, error)
	Create(dashboard *models.AdministratorDashboard) error
	GetByID(id uint) (*models.AdministratorDashboard, error)
	Update(dashboard *models.AdministratorDashboard) error
	Delete(id uint) error
}

type repoConfig struct {
	DashboardTable string
}

type dashboardRepository struct {
	ctx        context.Context
	cancel     context.CancelFunc
	database   *gorm.DB
	repoConfig *repoConfig
	logger     *zap.Logger
}

func NewGORMDashboardRepository(database *gorm.DB, servicePrefix string, logger *zap.Logger) (DashboardRepository, error) {
	database.AutoMigrate(&models.AdministratorDashboard{})
	repoConfig := &repoConfig{
		DashboardTable: servicePrefix + "_dashboards"}

	repository := &dashboardRepository{database: database, repoConfig: repoConfig, logger: logger}
	ctx, cancel := context.WithCancel(context.Background())
	repository.ctx = ctx
	repository.cancel = cancel

	return repository, nil
}

func (r *dashboardRepository) GetAll(paginationParams *types.PaginationParams, searchParams *models.AdministratorDashboard) (*database.PaginatedResult, error) {
	var data []*models.AdministratorDashboard
	var count int64

	db := r.database.Debug().Table(r.repoConfig.DashboardTable)

	query := db.Scopes()

	countQuery := r.database.Table(r.repoConfig.DashboardTable).Where("deleted_at IS NULL").Scopes()

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

func (r *dashboardRepository) Create(dashboard *models.AdministratorDashboard) error {
	return r.database.Table(r.repoConfig.DashboardTable).Create(dashboard).Error
}

func (r *dashboardRepository) GetByID(id uint) (*models.AdministratorDashboard, error) {
	var dashboard models.AdministratorDashboard
	if err := r.database.Table(r.repoConfig.DashboardTable).Where("id = ?", id).First(&dashboard).Error; err != nil {
		return nil, err
	}
	return &dashboard, nil
}

func (r *dashboardRepository) Update(dashboard *models.AdministratorDashboard) error {
	return r.database.Table(r.repoConfig.DashboardTable).Save(dashboard).Error
}

func (r *dashboardRepository) Delete(id uint) error {
	return r.database.Table(r.repoConfig.DashboardTable).Where("id = ?", id).Delete(&models.AdministratorDashboard{}).Error
}
