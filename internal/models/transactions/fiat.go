package transactions

import (
	"github.com/denizumutdereli/stream-admin/internal/dsl"
	"github.com/go-playground/validator"
)

var validateFiatTransactions *validator.Validate

type FiatTransactions struct {
	ID                 int64   `gorm:"primaryKey;type:bigint" json:"id" validate:"required"`
	UserID             int64   `gorm:"type:bigint" json:"user_id" validate:"required,gte=0"`
	TxID               string  `gorm:"type:varchar(255)" json:"tx_id" validate:"omitempty"`
	Currency           string  `gorm:"type:varchar(255)" json:"currency" validate:"required"`
	BankName           string  `gorm:"type:varchar(255)" json:"bank_name" validate:"omitempty"`
	Name               string  `gorm:"type:varchar(255)" json:"name" validate:"omitempty"`
	Iban               string  `gorm:"type:varchar(255)" json:"iban" validate:"required"`
	NationalId         string  `gorm:"type:varchar(255)" json:"national_id" validate:"required"`
	Amount             float64 `gorm:"type:numeric" json:"amount" validate:"required,gte=0"`
	TransactionOrderID string  `gorm:"type:varchar(255)" json:"transaction_order_id" validate:"omitempty"`
	Status             string  `gorm:"type:varchar(255)" json:"status" validate:"required"`
	Description        string  `gorm:"type:varchar(255)" json:"description" validate:"omitempty"`
	Type               string  `gorm:"type:varchar(255)" json:"type" validate:"required"`
	Timestamp          int64   `gorm:"type:bigint" json:"timestamp"`
	CreatedAt          int64   `gorm:"type:bigint" json:"created_at"`
	UpdatedAt          int64   `gorm:"type:bigint" json:"updated_at"`
}

type FiatTransactionsSearch struct {
	ID                 *int64   `form:"id" json:"id"`
	UserID             *int64   `form:"user_id" json:"user_id"`
	TxID               *string  `form:"tx_id" json:"tx_id"`
	Currency           *string  `form:"currency" json:"currency"`
	BankName           *string  `form:"bank_name" json:"bank_name"`
	Name               *string  `form:"name" json:"name"`
	Iban               *string  `form:"iban" json:"iban"`
	NationalId         *string  `form:"national_id" json:"national_id"`
	Amount             *float64 `form:"amount" json:"amount"`
	TransactionOrderID *string  `form:"transaction_order_id" json:"transaction_order_id"`
	Status             *string  `form:"status" json:"status"`
	Description        *string  `form:"description" json:"description"`
	Type               *string  `form:"type" json:"type"`
	Timestamp          *int64   `form:"timestamp" json:"timestamp"`
	CreatedAt          *int64   `form:"created_at" json:"created_at"`
	UpdatedAt          *int64   `form:"updated_at" json:"updated_at"`
	dsl.DSLFields      `gorm:"-" json:"-"`
}

func ValidateFiatTransactionsr(tx *FiatTransactions) error {
	return validateFiatTransactions.Struct(tx)
}
