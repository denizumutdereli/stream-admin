package models

import (
	"github.com/denizumutdereli/stream-admin/internal/dsl"
	"github.com/go-playground/validator"
)

var validateAssetsNetworks *validator.Validate

type AssetsNetworks struct {
	ID                      int64   `gorm:"primaryKey;type:bigint" json:"id" validate:"required"`
	Name                    string  `gorm:"type:varchar(255)" json:"name" validate:"required"`
	AddressRegex            string  `gorm:"type:text" json:"address_regex" validate:"omitempty"`
	Coin                    string  `gorm:"type:varchar(255)" json:"coin" validate:"required"`
	DepositEnable           int     `gorm:"type:smallint" json:"deposit_enable" validate:"required"`
	WithdrawEnable          int     `gorm:"type:smallint" json:"withdraw_enable" validate:"required"`
	MemoRegex               string  `gorm:"type:text" json:"memo_regex" validate:"omitempty"`
	MinConfirm              int     `gorm:"type:smallint" json:"min_confirm" validate:"required"`
	Network                 string  `gorm:"type:varchar(255)" json:"network" validate:"required"`
	WithdrawIntegerMultiple float64 `gorm:"type:numeric" json:"withdraw_integer_multiple" validate:"required"`
	WithdrawMax             float64 `gorm:"type:numeric" json:"withdraw_max" validate:"required"`
	WithdrawMin             float64 `gorm:"type:numeric" json:"withdraw_min" validate:"required"`
	SameAddress             int     `gorm:"type:smallint" json:"same_address" validate:"required"`
	WithdrawFee             float64 `gorm:"type:numeric" json:"withdraw_fee" validate:"required"`
	SpecialTips             string  `gorm:"type:text" json:"special_tips" validate:"omitempty"`
	WithdrawDesc            string  `gorm:"type:text" json:"withdraw_desc" validate:"omitempty"`
	Type                    string  `gorm:"type:varchar(255)" json:"type" validate:"omitempty"`
	UnlockConfirm           int     `gorm:"type:smallint" json:"unclock_confirm" validate:"required"`
	CreatedAt               int64   `gorm:"type:bigint" json:"created_at"`
	UpdatedAt               int64   `gorm:"type:bigint" json:"updated_at"`
	DeletedAt               int64   `gorm:"type:bigint" json:"deleted_at"`
}

type AssetsNetworksSearch struct {
	ID                      *int64   `form:"id"`
	Name                    *string  `form:"name"`
	AddressRegex            *string  `form:"address_regex"`
	Coin                    *string  `form:"coin"`
	DepositEnable           *int     `form:"deposit_enable"`
	WithdrawEnable          *int     `form:"withdraw_enable"`
	MemoRegex               *string  `form:"memo_regex"`
	MinConfirm              *int     `form:"min_confirm"`
	Network                 *string  `form:"network"`
	WithdrawIntegerMultiple *float64 `form:"withdraw_integer_multiple"`
	WithdrawMax             *float64 `form:"withdraw_max"`
	WithdrawMin             *float64 `form:"withdraw_min"`
	SameAddress             *int     `form:"same_address"`
	WithdrawFee             *float64 `form:"withdraw_fee"`
	SpecialTips             *string  `form:"special_tips"`
	WithdrawDesc            *string  `form:"withdraw_desc"`
	Type                    *string  `form:"type"`
	UnlockConfirm           *int     `form:"unclock_confirm"`
	CreatedAt               *int64   `form:"created_at"`
	UpdatedAt               *int64   `form:"updated_at"`
	DeletedAt               *int64   `form:"deleted_at"`
	dsl.DSLFields           `gorm:"-" json:"-"`
}

func ValidateAssetsNetworks(coin *AssetsNetworks) error {
	return validateAssetsNetworks.Struct(coin)
}
