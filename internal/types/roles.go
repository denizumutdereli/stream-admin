package types

type DefaultUserRoles string

const (
	SuperAdmin  DefaultUserRoles = "superAdmin"
	RegularUser DefaultUserRoles = "regular"
)

type SubRolePolicies struct {
	Source    string `json:"source"`
	Action    string `json:"action"`
	Allowance string `json:"allowance"`
}
