package schedule

import (
	"context"

	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator/reports"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ScheduleRepository interface {
	GetAll(paginationParams *types.PaginationParams, searchParams *models.AdministratorDashboardSchedule) (*database.PaginatedResult, error)
	Create(dashboard *models.AdministratorDashboardSchedule) error
	GetByID(id uint) (*models.AdministratorDashboardSchedule, error)
	Update(dashboard *models.AdministratorDashboardSchedule) error
	Delete(id uint) error
}

type repoConfig struct {
	ScheduleTable string
}

type scheduleRepository struct {
	ctx        context.Context
	cancel     context.CancelFunc
	database   *gorm.DB
	repoConfig *repoConfig
	logger     *zap.Logger
}

func NewGORMScheduleRepository(database *gorm.DB, servicePrefix string, logger *zap.Logger) (ScheduleRepository, error) {
	database.AutoMigrate(&models.AdministratorDashboardSchedule{})
	repoConfig := &repoConfig{
		ScheduleTable: servicePrefix + "_dashboard_schedules"}

	repository := &scheduleRepository{database: database, repoConfig: repoConfig, logger: logger}
	ctx, cancel := context.WithCancel(context.Background())
	repository.ctx = ctx
	repository.cancel = cancel

	return repository, nil
}

func (r *scheduleRepository) GetAll(paginationParams *types.PaginationParams, searchParams *models.AdministratorDashboardSchedule) (*database.PaginatedResult, error) {
	var data []*models.AdministratorDashboardSchedule
	var count int64

	db := r.database.Debug().Table(r.repoConfig.ScheduleTable)

	query := db.Scopes()

	countQuery := r.database.Table(r.repoConfig.ScheduleTable).Where("deleted_at IS NULL").Scopes()

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

func (r *scheduleRepository) Create(schedule *models.AdministratorDashboardSchedule) error {
	return r.database.Table(r.repoConfig.ScheduleTable).Create(schedule).Error
}

func (r *scheduleRepository) GetByID(id uint) (*models.AdministratorDashboardSchedule, error) {
	var schedule models.AdministratorDashboardSchedule
	if err := r.database.Table(r.repoConfig.ScheduleTable).Where("id = ?", id).First(&schedule).Error; err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *scheduleRepository) Update(schedule *models.AdministratorDashboardSchedule) error {
	return r.database.Table(r.repoConfig.ScheduleTable).Save(schedule).Error
}

func (r *scheduleRepository) Delete(id uint) error {
	return r.database.Table(r.repoConfig.ScheduleTable).Where("id = ?", id).Delete(&models.AdministratorDashboardSchedule{}).Error
}
