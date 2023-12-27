package repository

import (
	"gorm.io/gorm"
)

type AdminRepository interface {
	IsConnected() bool
}

type GORMAdminRepository struct {
	database *gorm.DB
}

func NewGORMAdminRepository(database *gorm.DB) AdminRepository {
	//database.AutoMigrate(&models.Admin{})

	return &GORMAdminRepository{database}
}

func (r *GORMAdminRepository) IsConnected() bool {
	sqlDB, err := r.database.DB()
	if err != nil {
		return false
	}

	return sqlDB.Ping() == nil
}
