package users

import (
	"github.com/denizumutdereli/stream-admin/internal/database"
	"github.com/denizumutdereli/stream-admin/internal/dsl"
	"github.com/go-playground/validator"
)

var validateUsers *validator.Validate

type User struct {
	ID                  int64  `gorm:"primaryKey;type:bigint" json:"id" validate:"required"`
	Email               string `gorm:"type:varchar(255)" json:"email" validate:"required,email"`
	Phone               string `gorm:"type:varchar(255)" json:"phone"`
	Password            string `gorm:"type:varchar(255)" json:"password" validate:"required"`
	ReferralUserID      int64  `gorm:"type:bigint" json:"referral_user_id"`
	ReferralKey         string `gorm:"type:varchar(12)" json:"referral_key"`
	Country             string `gorm:"type:varchar(255)" json:"country"`
	RegisterIP          string `gorm:"type:varchar(255)" json:"register_ip"`
	RegisterDevice      string `gorm:"type:varchar(255)" json:"register_device"`
	RegisterLocation    string `gorm:"type:varchar(200)" json:"register_location"`
	ProfilePicture      string `gorm:"type:varchar(255)" json:"profile_picture"`
	Status              string `gorm:"type:varchar(255)" json:"status" validate:"required"`
	G2FAEnabled         int    `gorm:"type:smallint" json:"g2fa_enabled" validate:"required"`
	G2FASecret          string `gorm:"type:varchar(255)" json:"g2fa_secret"`
	Ban                 int64  `gorm:"type:bigint" json:"ban"`
	LastKYCLevel        string `gorm:"type:varchar(50)" json:"last_kyc_level"`
	LastKYCDate         int64  `gorm:"type:bigint" json:"last_kyc_date"`
	KYCStatus           int    `gorm:"type:smallint" json:"kyc_status"`
	KYCPublished        int    `gorm:"type:smallint" json:"kyc_published"`
	MobileDisabledUntil string `gorm:"type:varchar(255)" json:"mobile_disabled_until"` //timestampz fix!
	WebDisabledUntil    int64  `gorm:"type:bigint" json:"web_disabled_until"`
	Agreements          string `gorm:"type:text" json:"agreements"`
	CreatedAt           int64  `gorm:"type:bigint" json:"created_at"`
	UpdatedAt           int64  `gorm:"type:bigint" json:"updated_at"`
	DeletedAt           int64  `gorm:"type:bigint" json:"deleted_at"`
}

type UserSearch struct {
	ID                  *int64  `form:"id"`
	Email               *string `form:"email"`
	Phone               *string `form:"phone"`
	ReferralUserID      *int64  `form:"referral_user_id"`
	ReferralKey         *string `form:"referral_key"`
	Country             *string `form:"country"`
	RegisterIP          *string `form:"register_ip"`
	RegisterDevice      *string `form:"register_device"`
	RegisterLocation    *string `form:"register_location"`
	Status              *string `form:"status"`
	G2FAEnabled         *int    `form:"g2fa_enabled"`
	G2FASecret          *string `form:"g2fa_secret"`
	Ban                 *int64  `form:"ban"`
	LastKYCLevel        *string `form:"last_kyc_level"`
	LastKYCDate         *int64  `form:"last_kyc_date"`
	KYCStatus           *int    `form:"kyc_status"`
	KYCPublished        *int    `form:"kyc_published"`
	CreatedAt           *int64  `form:"created_at"`
	UpdatedAt           *int64  `form:"updated_at"`
	MobileDisabledUntil *string `form:"mobile_disabled_until"`
	WebDisabledUntil    *int64  `form:"web_disabled_until"`
	DeletedAt           *int64  `form:"deleted_at"`
	dsl.DSLFields       `gorm:"-" json:"-"`
}

type UserSettings struct {
	UserID        int64   `json:"user_id"`
	Language      *string `form:"language,omitempty"`
	Theme         *string `form:"theme,omitempty"`
	Currency      *string `form:"currency,omitempty"`
	FavoritePairs *string `form:"favorite_pairs,omitempty"`
}

type UserBankAccounts struct {
	UserID    *string `json:"user_id"` // int vs string - very nice!
	BankTag   *string `form:"bank_tag,omitempty"`
	BankName  *string `form:"bank_name,omitempty"`
	TxCurreny *string `form:"tx_currency,omitempty"`
	OwnerName *string `form:"owner_name,omitempty"`
	Iban      *string `form:"iban,omitempty"`
	SwiftCode *string `form:"swift_code,omitempty"`
	IsActive  *int    `form:"is_active,omitempty"`
	Type      *string `form:"type,omitempty"`
	CreatedAt *string `form:"create_at,omitempty"`  //unixtime fix!!
	UpdatedAt *string `form:"updated_at,omitempty"` //unixtime fix!!
	DeletedAt *string `form:"deleted_at,omitempty"` //unixtime fix!!
}

type SearchWithSettings struct {
	UserSearch
	UserSettings

	//KycData *[]UserKYCSearch `gorm:"-"`
	//Files   *[]FilesOfKYCs   `gorm:"-"`
}

type SearchWithFullJoins struct {
	UserSearch
	UserSettings
	KycData            *[]UserKYCSearch          `gorm:"-"`
	UserBanks          *[]UserBankAccounts       `gorm:"-"`
	UserOrders         *database.PaginatedResult `gorm:"-"`
	FiatTransactions   *database.PaginatedResult `gorm:"-"`
	CryptoTransactions *database.PaginatedResult `gorm:"-"`
	CryptoWallets      *database.PaginatedResult `gorm:"-"`

	//KycData *[]UserKYCSearch `gorm:"-"`
	//Files   *[]FilesOfKYCs   `gorm:"-"`
}

func ValidateUsers(user *User) error {
	return validateUsers.Struct(user)
}
