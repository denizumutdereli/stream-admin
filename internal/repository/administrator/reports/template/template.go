package template

import (
	"context"

	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator/reports"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DashboardTemplateRepository interface {
	GetAll(paginationParams *types.PaginationParams, searchParams *models.AdministratorDashboardTemplate) (*database.PaginatedResult, error)
	Create(template *models.AdministratorDashboardTemplate) error
	GetByID(id uint) (*models.AdministratorDashboardTemplate, error)
	Update(template *models.AdministratorDashboardTemplate) error
	Delete(id uint) error
}

type repoConfig struct {
	TemplateTable string
}

type dashboardTemplateRepository struct {
	ctx        context.Context
	cancel     context.CancelFunc
	database   *gorm.DB
	repoConfig *repoConfig
	logger     *zap.Logger
}

func NewGORMDashboardTemplateRepository(database *gorm.DB, servicePrefix string, logger *zap.Logger) (DashboardTemplateRepository, error) {
	database.AutoMigrate(&models.AdministratorDashboardTemplate{})
	repoConfig := &repoConfig{
		TemplateTable: servicePrefix + "_dashboard_templates"}

	repository := &dashboardTemplateRepository{database: database, repoConfig: repoConfig, logger: logger}
	ctx, cancel := context.WithCancel(context.Background())
	repository.ctx = ctx
	repository.cancel = cancel

	return repository, nil
}

func (r *dashboardTemplateRepository) GetAll(paginationParams *types.PaginationParams, searchParams *models.AdministratorDashboardTemplate) (*database.PaginatedResult, error) {
	var data []*models.AdministratorDashboardTemplate
	var count int64

	db := r.database.Debug().Table(r.repoConfig.TemplateTable)

	query := db.Scopes()

	countQuery := r.database.Table(r.repoConfig.TemplateTable).Where("deleted_at IS NULL").Scopes()

	if err := countQuery.Count(&count).Error; err != nil {
		r.logger.Error("error counting dashboard templates:", zap.Error(err))
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

func (r *dashboardTemplateRepository) Create(template *models.AdministratorDashboardTemplate) error {
	return r.database.Table(r.repoConfig.TemplateTable).Create(template).Error
}

func (r *dashboardTemplateRepository) GetByID(id uint) (*models.AdministratorDashboardTemplate, error) {
	var template models.AdministratorDashboardTemplate
	if err := r.database.Table(r.repoConfig.TemplateTable).Where("id = ?", id).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *dashboardTemplateRepository) Update(template *models.AdministratorDashboardTemplate) error {
	return r.database.Table(r.repoConfig.TemplateTable).Save(template).Error
}

func (r *dashboardTemplateRepository) Delete(id uint) error {
	return r.database.Table(r.repoConfig.TemplateTable).Where("id = ?", id).Delete(&models.AdministratorDashboardTemplate{}).Error
}
