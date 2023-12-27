package users

import "github.com/denizumutdereli/stream-admin/internal/types"

type OrdersDetailsDSLQuery struct {
	OrdersDSLSearch *string `json:"orders_dsl_search"`
	QueryConditions *[]types.QueryCondition
}

type UserDetailsIncluding struct {
	KYC                *bool `form:"kyc"`
	Banks              *bool `form:"banks"`
	Orders             *bool `form:"orders"`
	Logins             *bool `form:"logins"`
	FiatTransactions   *bool `form:"fiat_transactions"`
	CryptoTransactions *bool `form:"crypto_transactions"`
	CryptoWallets      *bool `form:"crypto_wallets"`
	// KYCQuery     DetailsDSLQuery
	// BanksQuery   DetailsDSLQuery
	// OrdersQuery OrdersDetailsDSLQuery
}

func NewUserDetailsIncludingWithDefaults() *UserDetailsIncluding {
	return &UserDetailsIncluding{
		KYC:   ptrBool(false),
		Banks: ptrBool(false),
		// Orders: ptrBool(false),
		Logins: ptrBool(false),
		// CryptoTransactions: ptrBool(false),
		// CryptoWallets:      ptrBool(false),
		// FiatTransactions:   ptrBool(false),
	}
}

func ptrBool(b bool) *bool {
	return &b
}
