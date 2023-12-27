package roles

import (
	"gorm.io/gorm"
)

func (z *adminUserRolesRepository) SearchRoleByName(rolename string, db *gorm.DB) *gorm.DB {
	db.Where("role_name = ? ", rolename)
	return db
}

func (z *adminUserRolesRepository) SearchRoleById(roleId int, db *gorm.DB) *gorm.DB {
	db.Where("id = ? ", roleId)
	return db
}
