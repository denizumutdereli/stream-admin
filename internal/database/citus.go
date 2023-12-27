package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type CitusDSN struct {
	Host   string
	Port   string
	User   string
	Pass   string
	DBName string
}

func NewCitusDB(dsn *CitusDSN) (*gorm.DB, error) {

	connectionDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Istanbul", dsn.Host, dsn.User, dsn.Pass, dsn.DBName, dsn.Port)

	a, err := gorm.Open(postgres.Open(connectionDSN), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}
	return a, nil
}
