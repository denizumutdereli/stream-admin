package models

import (
	"time"
)

type AdministratorAuth struct {
	ConfigID  string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
