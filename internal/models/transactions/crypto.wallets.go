package transactions

import (
	"github.com/denizumutdereli/stream-admin/internal/dsl"
	"github.com/go-playground/validator"
)

var validateCryptoWallets *validator.Validate

type CryptoWallets struct {
	ID        int64  `gorm:"primaryKey;type:bigint" json:"id" validate:"required"`
	UserID    int64  `gorm:"type:bigint" json:"user_id" validate:"required,gte=0"`
	Coin      string `gorm:"type:varchar(255)" json:"coin" validate:"required"`
	Network   string `gorm:"type:varchar(255)" json:"network" validate:"required"`
	Address   string `gorm:"type:varchar(255)" json:"address" validate:"omitempty"`
	CreatedAt int64  `gorm:"type:bigint" json:"created_at"`
	UpdatedAt int64  `gorm:"type:bigint" json:"updated_at"`
}

type CryptoWalletsSearch struct {
	ID            *int64  `form:"id" json:"id"`
	UserID        *int64  `form:"user_id" json:"user_id"`
	Coin          *string `form:"coin" json:"coin"`
	Network       *string `form:"network" json:"network"`
	Address       *string `form:"address" json:"address"`
	CreatedAt     *int64  `form:"created_at" json:"created_at"`
	UpdatedAt     *int64  `form:"updated_at" json:"updated_at"`
	dsl.DSLFields `gorm:"-" json:"-"`
}

func ValidateCryptoWallets(wallet *CryptoWallets) error {
	return validateCryptoWallets.Struct(wallet)
}
