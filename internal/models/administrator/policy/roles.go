package policy

import (
	"time"

	"github.com/denizumutdereli/stream-admin/internal/dsl"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type PolicyTargeting string
type PolicyStatus string
type PolicyEditing string

const (
	RoleStatusActive PolicyStatus = "active"
	RoleStatusPaused PolicyStatus = "paused"

	RolesPolicy PolicyTargeting = "roles"

	PolicyReadonly PolicyEditing = "true"
	PolicyEditable PolicyEditing = "false"
)

var validateAdminRolePolicy *validator.Validate

type AdministratorRolePolicy struct {
	PolicyID       string                  `gorm:"primaryKey;size:255" json:"policy_id"`
	Readonly       PolicyEditing           `gorm:"type:varchar(255);default:'false'" json:"readonly" validate:"policyEditing"`
	Target         PolicyTargeting         `gorm:"type:varchar(255);default:'roles'" json:"target" validate:"policyTarget"`
	Title          string                  `gorm:"unique;not null;size:255" json:"title" validate:"required"`
	SubPolicies    string                  `gorm:"type:json" json:"sub_policies" validate:"required"`
	SubPolicyRules []types.SubRolePolicies `gorm:"-" json:"sub_policy_rules"`
	Status         PolicyStatus            `gorm:"type:varchar(255);default:'active'" json:"status" validate:"policyStatus"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	DeletedAt      gorm.DeletedAt          `gorm:"index" json:"deleted_at"`
}

type AdministratorRolePolicyResponse struct {
	PolicyID       string                  `json:"policy_id"`
	Readonly       PolicyEditing           `json:"readonly"`
	Target         PolicyTargeting         `json:"target"`
	Title          string                  `json:"title"`
	SubPolicyRules []types.SubRolePolicies `json:"sub_policy_rules"`
	Status         PolicyStatus            `json:"status"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	DeletedAt      *time.Time              `json:"deleted_at"`
}

type AdministratorRolePolicySearch struct {
	PolicyID      *string          `form:"policy_id"`
	Readonly      *PolicyEditing   `form:"readonly"`
	PolicyTarget  *PolicyTargeting `form:"target"`
	Title         *string          `form:"title"`
	Status        *PolicyStatus    `form:"status"`
	CreatedAt     *time.Time       `form:"created_at"`
	UpdatedAt     *time.Time       `form:"updated_at"`
	DeletedAt     *gorm.DeletedAt  `form:"deleted_at"`
	dsl.DSLFields `gorm:"-" json:"-"`
}

func ValidateAdminRolePolicy(adminRolePolicy *AdministratorRolePolicy) error {
	return validateAdminRolePolicy.Struct(adminRolePolicy)
}

func policyStatusValidation(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	switch PolicyStatus(status) {
	case RoleStatusActive, RoleStatusPaused:
		return true
	default:
		return false
	}
}

func policyTargetingValidation(fl validator.FieldLevel) bool {
	target := fl.Field().String()
	switch PolicyTargeting(target) {
	case RolesPolicy:
		return true
	default:
		return false
	}
}

func policyEditingValidation(fl validator.FieldLevel) bool {
	readonlyOption := fl.Field().String()
	switch PolicyEditing(readonlyOption) {
	case PolicyReadonly, PolicyEditable:
		return true
	default:
		return false
	}
}

func init() {
	validateAdminRolePolicy = validator.New()
	validateAdminRolePolicy.RegisterValidation("policyStatus", policyStatusValidation)
	validateAdminRolePolicy.RegisterValidation("policyTarget", policyTargetingValidation)
	validateAdminRolePolicy.RegisterValidation("policyEditing", policyEditingValidation)
}
