package transactions

import (
	"github.com/denizumutdereli/stream-admin/internal/dsl"
	"github.com/go-playground/validator"
)

var validateCryptoTransactions *validator.Validate

type CryptoTransactions struct {
	ID                 int64   `gorm:"primaryKey;type:bigint" json:"id" validate:"required"`
	UserID             int64   `gorm:"type:bigint" json:"user_id" validate:"required,gte=0"`
	Coin               string  `gorm:"type:varchar(255)" json:"coin" validate:"required"`
	Network            string  `gorm:"type:varchar(255)" json:"network" validate:"required"`
	SourceAddress      string  `gorm:"type:varchar(255)" json:"source_address" validate:"omitempty"`
	DestinationAddress string  `gorm:"type:varchar(255)" json:"destination_address" validate:"omitempty"`
	Amount             float64 `gorm:"type:numeric" json:"amount" validate:"required,gte=0"`
	Memo               string  `gorm:"type:varchar(255)" json:"memo" validate:"omitempty"`
	TxID               string  `gorm:"type:varchar(255)" json:"tx_id" validate:"omitempty"`
	Status             string  `gorm:"type:varchar(255)" json:"status" validate:"required"`
	Type               string  `gorm:"type:varchar(255)" json:"type" validate:"required"`
	TransactionOrderID string  `gorm:"type:varchar(255)" json:"transaction_order_id" validate:"required"`
	Currency           string  `gorm:"type:varchar(255)" json:"currency" validate:"omitempty"`
	CurrencyValue      float64 `gorm:"type:numeric" json:"currency_value" validate:"required,gte=0"`
	Fee                float64 `gorm:"type:numeric" json:"fee" validate:"required,gte=0"`
	CreatedAt          int64   `gorm:"type:bigint" json:"created_at"`
	UpdatedAt          int64   `gorm:"type:bigint" json:"updated_at"`
}

type CryptoTransactionsSearch struct {
	ID                 *int64   `form:"id" json:"id"`
	UserID             *int64   `form:"user_id" json:"user_id"`
	Coin               *string  `form:"coin" json:"coin"`
	Network            *string  `form:"network" json:"network"`
	SourceAddress      *string  `form:"source_address" json:"source_address"`
	DestinationAddress *string  `form:"destination_address" json:"destination_address"`
	Amount             *float64 `form:"amount" json:"amount"`
	Memo               *string  `form:"memo" json:"memo"`
	TxID               *string  `form:"tx_id" json:"tx_id"`
	Status             *string  `form:"status" json:"status"`
	Type               *string  `form:"type" json:"type"`
	TransactionOrderID *string  `form:"transaction_order_id" json:"transaction_order_id"`
	Currency           *string  `form:"currency" json:"currency"`
	CurrencyValue      *float64 `form:"currency_value" json:"currency_value"`
	Fee                *float64 `form:"fee" json:"fee"`
	CreatedAt          *int64   `form:"created_at" json:"created_at"`
	UpdatedAt          *int64   `form:"updated_at" json:"updated_at"`
	dsl.DSLFields      `gorm:"-" json:"-"`
}

func ValidateCryptoTransactions(tx *CryptoTransactions) error {
	return validateCryptoTransactions.Struct(tx)
}
