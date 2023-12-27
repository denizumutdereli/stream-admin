package models

import (
	"github.com/go-playground/validator"
	"gorm.io/gorm"
)

var validateAdmin *validator.Validate

type Administrator struct {
	gorm.Model
	Username string `gorm:"size:20;unique" json:"username" validate:"required,alphanum,min=3,max=20"`
	Password string `json:"password" validate:"required"`
}

type AdminUpdate struct {
	Username string `gorm:"size:20;unique" json:"username" validate:"required,alphanum,min=3,max=20"`
}

func ValidateAdmin(admin *Administrator) error {
	return validateAdmin.Struct(admin)
}

func (a *Administrator) BeforeSave(tx *gorm.DB) error {
	return nil
}

func (a *Administrator) BeforeCreate(tx *gorm.DB) error {
	return nil
}
