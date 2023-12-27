package users

import (
	"github.com/denizumutdereli/stream-admin/internal/dsl"
	"github.com/go-playground/validator"
)

var validateUserKyc *validator.Validate

type UserKYC struct {
	ID                 int64   `gorm:"primaryKey;type:bigint" json:"id" validate:"required"`
	UserID             int64   `gorm:"type:bigint;not null" json:"user_id" validate:"required"`
	FirstName          string  `gorm:"type:varchar;not null" json:"first_name" validate:"required"`
	LastName           string  `gorm:"type:varchar;not null" json:"last_name" validate:"required"`
	BirthDate          string  `gorm:"type:varchar;not null" json:"birth_date" validate:"required"`
	IDCardNumber       string  `gorm:"type:varchar;not null" json:"id_card_number" validate:"required"`
	IDCardSerial       string  `gorm:"type:varchar" json:"id_card_serial,omitempty"`
	CreatedAt          int64   `gorm:"type:bigint" json:"created_at"`
	UpdatedAt          int64   `gorm:"type:bigint" json:"updated_at"`
	KYCProvider        string  `gorm:"type:varchar" json:"kyc_provider,omitempty"`
	DeletedAt          int64   `gorm:"type:bigint" json:"deleted_at"`
	Nationality        string  `gorm:"type:varchar" json:"nationality,omitempty"`
	Status             string  `gorm:"type:text" json:"status,omitempty"`
	RejectStep         string  `gorm:"type:text" json:"reject_step,omitempty"`
	RejectType         string  `gorm:"type:text" json:"reject_type,omitempty"`
	Score              float64 `gorm:"type:float8" json:"score,omitempty"`
	DocumentExpiryDate int     `gorm:"type:int4" json:"document_expiry_date,omitempty"`
	NeedManualCheck    int16   `gorm:"type:smallint" json:"need_manual_check,omitempty"`
}

type UserKYCSearch struct {
	ID                 *int64         `form:"id"`
	UserID             *int64         `form:"user_id"`
	FirstName          *string        `form:"first_name"`
	LastName           *string        `form:"last_name"`
	BirthDate          *string        `form:"birth_date"`
	IDCardNumber       *string        `form:"id_card_number"`
	IDCardSerial       *string        `form:"id_card_serial"`
	CreatedAt          *int64         `form:"created_at"`
	UpdatedAt          *int64         `form:"updated_at"`
	KYCProvider        *string        `form:"kyc_provider"`
	Nationality        *string        `form:"nationality"`
	Status             *string        `form:"status"`
	RejectStep         *string        `form:"reject_step"`
	RejectType         *string        `form:"reject_type"`
	Score              *float64       `form:"score"`
	DocumentExpiryDate *int           `form:"document_expiry_date"`
	NeedManualCheck    *int16         `form:"need_manual_check"`
	KYCFiles           *[]UserKYCFile `gorm:"-"`
	UserFiles          *[]FilesOfKYCs `gorm:"-"`
	dsl.DSLFields      `gorm:"-" json:"-"`
}

func ValidateUserKYC(kyc *UserKYC) error {
	return validateUserKyc.Struct(kyc)
}
