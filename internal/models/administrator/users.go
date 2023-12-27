package models

import (
	"net"
	"strings"
	"time"
	"unicode"

	"github.com/denizumutdereli/stream-admin/internal/dsl"
	rolePolicyModels "github.com/denizumutdereli/stream-admin/internal/models/administrator/policy"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

/* TODO: separate the role to individiual model folder */

type (
	UserStatus       string
	DefaultUserRoles string
)

const (
	UserStatusPending  UserStatus = "pending"
	UserStatusPaused   UserStatus = "paused"
	UserStatusVerified UserStatus = "verified"

	SuperAdmin  DefaultUserRoles = "superAdmin"
	RegularUser DefaultUserRoles = "regular"
)

var validateAdminUsers *validator.Validate

type AdministratorUser struct {
	UserID           string             `gorm:"primaryKey;varchar(255)" json:"user_id"`
	EmployerName     string             `gorm:"unique;not null;varchar(255)" json:"employer_name" validate:"required,min=3"`
	Username         string             `gorm:"unique;not null;varchar(255)" json:"username" validate:"required,email"`
	Password         string             `gorm:"not null;varchar(255)" json:"-"`
	PhoneNumber      string             `gorm:"not null;varchar(255)" json:"phone_number" validate:"required,phone"`
	VerificationCode string             `gorm:"varchar(255)" json:"-"`
	VpnAddr          string             `gorm:"type:varchar(255)" json:"vpn_addr" validate:"vpnaddr"`
	UserRole         string             `gorm:"not null;type:varchar(255)" json:"user_role" validate:"required"`
	Role             *AdministratorRole `gorm:"foreignKey:UserRole;references:RoleID;onUpdate:CASCADE;onDelete:SET NULL"`
	Status           UserStatus         `gorm:"type:varchar(255);default:'pending'" json:"status" validate:"userStatus"`
	LastLogin        time.Time          `gorm:"-" json:"last_login" validate:"omitempty"`
	CreatedAt        time.Time          `json:"CreatedAt"`
	UpdatedAt        time.Time          `json:"UpdatedAt"`
}

type AdministratorRole struct {
	RoleID        string                                     `gorm:"primaryKey;varchar(255)" json:"role_id"`
	RoleName      string                                     `gorm:"unique;not null;varchar(255)" json:"role_name" validate:"required"`
	Policies      []rolePolicyModels.AdministratorRolePolicy `gorm:"many2many:administrator_roles_policies;" json:"policies"`
	CreatedAt     time.Time                                  `json:"CreatedAt"`
	UpdatedAt     time.Time                                  `json:"UpdatedAt"`
	DeletedAt     time.Time                                  `json:"DeletedAt"`
	dsl.DSLFields `gorm:"-" json:"-"`
}

type RoleWithPolicyTitles struct {
	RoleID   string        `json:"role_id"`
	RoleName string        `json:"role_name"`
	Policies []PolicyTitle `json:"policies"`
}

type PolicyTitle struct {
	PolicyID string `json:"policy_id"`
	Title    string `json:"title"`
}

/* search models -----------------------------------------------------*/
type AdministratorUserSearch struct {
	UserID        *string     `form:"user_id"`
	EmployerName  *string     `form:"employer_name"`
	Username      *string     `form:"username"`
	PhoneNumber   *string     `form:"phone_number"`
	UserRole      *string     `form:"user_role"`
	Status        *UserStatus `form:"status"`
	dsl.DSLFields `gorm:"-" json:"-"`
}

type AdministratorRoleSearch struct {
	RoleID        *string `form:"role_id"`
	RoleName      *string `form:"role_name"`
	dsl.DSLFields `gorm:"-" json:"-"`
}

func ValidateAdminRole(adminRole *AdministratorRole) error {
	return validateAdminUsers.Struct(adminRole)
}

func ValidateAdminUser(adminUser *AdministratorUser) error {
	return validateAdminUsers.Struct(adminUser)
}

func phoneValidation(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if !strings.HasPrefix(phone, "+") {
		return false
	}
	for _, char := range phone[1:] {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return len(phone) >= 12
}

func vpnAddressesValidation(fl validator.FieldLevel) bool {
	addresses := strings.Split(fl.Field().String(), ",")
	for _, addr := range addresses {
		if net.ParseIP(strings.TrimSpace(addr)) == nil {
			return false
		}
	}
	return true
}

func userStatusValidation(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	switch UserStatus(status) {
	case UserStatusPending, UserStatusPaused, UserStatusVerified:
		return true
	default:
		return false
	}
}

func init() {
	validateAdminUsers = validator.New()
	validateAdminUsers.RegisterValidation("userStatus", userStatusValidation)
	validateAdminUsers.RegisterValidation("vpnaddr", vpnAddressesValidation)
	validateAdminUsers.RegisterValidation("phone", phoneValidation)

}

func (a *AdministratorUser) BeforeSave(tx *gorm.DB) error {
	return nil
}

func (a *AdministratorUser) BeforeCreate(tx *gorm.DB) error {
	return nil
}
