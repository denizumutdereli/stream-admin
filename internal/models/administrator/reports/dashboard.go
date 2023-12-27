package reports

import (
	"time"

	"github.com/go-playground/validator"
)

var validate *validator.Validate

type AdministratorDashboardTemplate struct {
	ID          uint       `json:"id" validate:"required" gorm:"primaryKey"`
	Name        string     `json:"name" validate:"required,max=100" gorm:"type:varchar(100);not null;unique"`
	Description string     `json:"description" validate:"omitempty,max=1000" gorm:"type:text"`
	DataTypes   string     `json:"data_types" validate:"omitempty,max=500" gorm:"type:text"`
	IsLive      bool       `json:"is_live" gorm:"default:false"`
	CreatedBy   uint       `json:"created_by" validate:"required"`
	UpdatedBy   uint       `json:"updated_by" validate:"required"`
	DeletedBy   *uint      `json:"deleted_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

type AdministratorDashboard struct {
	ID                 uint                            `json:"id" validate:"required" gorm:"primaryKey"`
	UserID             uint                            `json:"user_id" validate:"required" gorm:"constraint:OnUpdate:SET NULL,OnDelete:SET NULL;"`
	TemplateID         uint                            `json:"template_id" validate:"required" gorm:"constraint:OnUpdate:SET NULL,OnDelete:SET NULL;"`
	Name               string                          `json:"name" validate:"required,max=100" gorm:"type:varchar(100);not null;unique"`
	Description        string                          `json:"description" validate:"omitempty,max=1000" gorm:"type:text"`
	IncludedQueries    []AdministratorDashboardQuery   `json:"included_queries" gorm:"many2many:tab_included_queries;constraint:OnDelete:CASCADE;"`
	ExcludedQueries    []AdministratorDashboardQuery   `json:"excluded_queries" gorm:"many2many:tab_excluded_queries;constraint:OnDelete:CASCADE;"`
	Tabs               []AdministratorDashboardTab     `json:"tabs" gorm:"foreignKey:DashboardID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SelectedScheduleID *uint                           `json:"selected_schedule_id"`
	SelectedSchedule   *AdministratorDashboardSchedule `json:"selected_schedule" gorm:"foreignKey:SelectedScheduleID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CreatedBy          uint                            `json:"created_by" validate:"required"`
	UpdatedBy          uint                            `json:"updated_by" validate:"required"`
	DeletedBy          *uint                           `json:"deleted_by"`
	CreatedAt          time.Time                       `json:"created_at"`
	UpdatedAt          time.Time                       `json:"updated_at"`
	DeletedAt          *time.Time                      `json:"deleted_at"`
}

type AdministratorDashboardTab struct {
	ID              uint                          `json:"id" validate:"required" gorm:"primaryKey"`
	DashboardID     uint                          `json:"dashboard_id" validate:"required"`
	Order           int                           `json:"order" validate:"required" gorm:"not null"`
	Name            string                        `json:"name" validate:"required,max=100" gorm:"type:varchar(100);not null;unique"`
	IncludedQueries []AdministratorDashboardQuery `json:"included_queries" gorm:"many2many:tab_included_queries;constraint:OnDelete:CASCADE;"`
	ExcludedQueries []AdministratorDashboardQuery `json:"excluded_queries" gorm:"many2many:tab_excluded_queries;constraint:OnDelete:CASCADE;"`
	CreatedAt       time.Time                     `json:"created_at"`
	UpdatedAt       time.Time                     `json:"updated_at"`
}

type AdministratorDashboardQuery struct {
	ID           uint      `json:"id" validate:"required" gorm:"primaryKey"`
	Name         string    `json:"name" validate:"required,max=100" gorm:"type:varchar(100);not null;unique"`
	Filters      string    `json:"filters" validate:"omitempty,max=1000" gorm:"type:text"`
	DSLFilters   string    `json:"dsl_filters" validate:"omitempty,json" gorm:"type:jsonb"`
	IsPredefined bool      `json:"is_predefined" gorm:"default:false"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AdministratorDashboardSchedule struct {
	ID        uint   `json:"id" validate:"required" gorm:"primaryKey"`
	Mechanics string `json:"mechanics" validate:"required,json" gorm:"type:jsonb"`
}

type AdministratorDashboardScheduleMechanics struct {
	DateInterval   *string `json:"date_interval"`
	SendingChannel *string `json:"sending_channel"`
	CronExpression *string `json:"cron_expression" validate:"omitempty,cronexpr" gorm:"type:varchar(100)"`
}

func ValidateData(data *interface{}) error {
	return validate.Struct(data)
}
