package factory

import (
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

/*
for service CRUD TODO: managing multiple connection instances
*/

type DatabaseFactory struct {
	config *config.Config
	logger *zap.Logger
}

func NewDatabaseFactory(config *config.Config, logger *zap.Logger) *DatabaseFactory {
	return &DatabaseFactory{
		config: config,
		logger: logger,
	}
}

func (df *DatabaseFactory) CreateCitusDB() (*gorm.DB, error) {
	db, err := database.NewCitusDB(df.config.Database)
	if err != nil {
		df.logger.Error("Failed to create Citus database connection", zap.Error(err))
		return nil, err
	}
	df.logger.Info("Connected to Citus database")
	return db, nil
}
