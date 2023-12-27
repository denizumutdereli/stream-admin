package users

import (
	"gorm.io/gorm"
)

func (z *adminUserRepository) SearchUser(userid int, db *gorm.DB) *gorm.DB {
	db.Where("id = ? ", userid)
	return db
}

func (z *adminUserRepository) SearchRoleByName(rolename string, db *gorm.DB) *gorm.DB {
	db.Where("role_name = ? ", rolename)
	return db
}

func (z *adminUserRepository) SearchRoleById(roleId int, db *gorm.DB) *gorm.DB {
	db.Where("id = ? ", roleId)
	return db
}
