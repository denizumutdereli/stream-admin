package models

import (
	"github.com/denizumutdereli/stream-admin/internal/dsl"
	"github.com/go-playground/validator"
)

var validateAssetsCoins *validator.Validate

type AssetsCoins struct {
	ID                  int64  `gorm:"primaryKey;type:bigint" json:"id" validate:"required"`
	Name                string `gorm:"type:varchar(255)" json:"name" validate:"required"`
	Title               string `gorm:"type:varchar(255)" json:"title" validate:"required"`
	Icon                string `gorm:"type:varchar(255)" json:"icon" validate:"omitempty"`
	IsLegalMoney        int    `gorm:"type:smallint" json:"is_legal_money" validate:"required"`
	WithdrawAllEnable   int    `gorm:"type:smallint" json:"withdraw_all_enable" validate:"required"`
	WithdrawProvider    string `gorm:"type:varchar(255)" json:"withdraw_provider" validate:"required"`
	DepositProvider     string `gorm:"type:varchar(255)" json:"deposit_provider" validate:"required"`
	AssetDetailProvider string `gorm:"type:varchar(255)" json:"asset_detail_provider" validate:"required"`
	Type                string `gorm:"type:varchar(255)" json:"type" validate:"required"`
	CreatedAt           int64  `gorm:"type:bigint" json:"created_at"`
	UpdatedAt           int64  `gorm:"type:bigint" json:"updated_at"`
	DeletedAt           int64  `gorm:"type:bigint" json:"deleted_at"`
}

type AssetsCoinsSearch struct {
	ID                  *int64  `form:"id"`
	Name                *string `form:"name"`
	Title               *string `form:"title"`
	IsLegalMoney        *int    `form:"is_legal_money"`
	WithdrawAllEnable   *int    `form:"withdraw_all_enable"`
	WithdrawProvider    *string `form:"withdraw_provider"`
	DepositProvider     *string `form:"deposit_provider"`
	AssetDetailProvider *string `form:"asset_detail_provider"`
	Type                *string `form:"type"`
	CreatedAt           *int64  `form:"created_at"`
	UpdatedAt           *int64  `form:"updated_at"`
	DeletedAt           *int64  `form:"deleted_at"`
	DSLSearch           *string `form:"dsl_search"`
	dsl.DSLFields       `gorm:"-" json:"-"`
}

func ValidateAssetsCoins(coin *AssetsCoins) error {
	return validateAssetsCoins.Struct(coin)
}
