package models

import (
	"time"

	"github.com/denizumutdereli/stream-admin/internal/dsl"
)

type AdministratorLogs struct {
	ID         int64     `gorm:"primary_key;type:bigint;autoIncrement:true" json:"id"`
	LogLevel   int       `gorm:"type:smallint" json:"log_level" validate:"required"`
	UserID     string    `gorm:"type:text" json:"user_id" validate:"required"`
	UserRole   string    `gorm:"type:text" json:"user_role" validate:"required"`
	Action     string    `gorm:"type:text" json:"action" validate:"required"`
	Method     string    `gorm:"type:text" json:"method" validate:"required"`
	Ip         string    `gorm:"type:text" json:"ip" validate:"required"`
	Status     int       `gorm:"type:int" json:"status" validate:"required"`
	UserAgent  string    `gorm:"type:text" json:"user_agent" validate:"required"`
	Timestamps time.Time `gorm:"type:timestamp" json:"timestamps"`
	CreatedAt  int64     `gorm:"type:bigint" json:"created_at"`
	UpdatedAt  int64     `gorm:"type:bigint" json:"updated_at"`
	DeletedAt  int64     `gorm:"type:bigint" json:"deleted_at"`
}

type AdministratorLogsSearch struct {
	LogLevel      *int       `form:"log_level"`
	UserID        *string    `form:"user_id"`
	UserRole      *string    `form:"user_role"`
	Action        *string    `form:"action"`
	Method        *string    `form:"method"`
	Ip            *string    `form:"ip"`
	Status        *int       `form:"status"`
	UserAgent     *string    `form:"user_agent"`
	Timestamps    *time.Time `form:"timestamps"`
	CreatedAt     **int64    `form:"created_at"`
	UpdatedAt     int64      `form:"updated_at"`
	DeletedAt     *int64     `form:"deleted_at"`
	dsl.DSLFields `gorm:"-" json:"-"`
}
