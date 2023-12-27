package users

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (z *usersRepository) SearchUser(userid int, db *gorm.DB) *gorm.DB {
	db.Where("id = ? ", userid)
	return db
}

func (z *usersRepository) JoinWithUserSettings(db *gorm.DB) *gorm.DB {
	query := fmt.Sprintf("LEFT JOIN %s ON %s.id = %s.user_id",
		z.repoConfig.UserSettingsTable,
		z.repoConfig.UserTable,
		z.repoConfig.UserSettingsTable,
	)
	return db.Joins(query)
}

func (z *usersRepository) JoinWithUserKYCData(db *gorm.DB) *gorm.DB {
	query := fmt.Sprintf("LEFT JOIN %s ON %s.id = %s.user_id",
		z.repoConfig.KycTable,
		z.repoConfig.UserTable,
		z.repoConfig.KycTable,
	)
	return db.Joins(query)
}

func (z *usersRepository) GroupByUserIDWithSettings(db *gorm.DB) *gorm.DB {
	return db.Group(fmt.Sprintf("%s.id, %s.user_id", z.repoConfig.UserTable, z.repoConfig.UserSettingsTable))
}

func (z *usersRepository) SelectFieldsWithSettings(fields []string, targetStruct interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var err error
		var selectFields string

		defaultFields := fmt.Sprintf("%s.*", z.repoConfig.UserTable)

		if len(fields) > 0 {
			selectFields, err = z.builders.SelectFields(fields, z.repoConfig.UserSettingsTable, targetStruct)
			if err != nil {
				z.logger.Error("select fields failed", zap.Error(err))
			}
		}

		selectedFieldArrays := []string{defaultFields, selectFields}

		return db.Select(strings.Join(selectedFieldArrays, ", "))
	}
}
