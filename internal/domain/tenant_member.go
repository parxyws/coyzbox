package domain

import "time"

type TenantRole string

const (
	TenantRoleOwner  TenantRole = "owner"
	TenantRoleMember TenantRole = "member"
)

// TenantMember is the junction table mapping users to tenants with roles.
// A user can belong to multiple tenants, and each membership has a role
// that determines their access level within that tenant.
type TenantMember struct {
	Id       string     `json:"id" gorm:"column:id;primaryKey"`
	TenantId string     `json:"tenant_id" gorm:"column:tenant_id;uniqueIndex:idx_tenant_user"`
	UserId   string     `json:"user_id" gorm:"column:user_id;uniqueIndex:idx_tenant_user"`
	Role     TenantRole `json:"role" gorm:"column:role;default:member"`
	JoinedAt time.Time  `json:"joined_at" gorm:"column:joined_at"`

	// Relations
	Tenant Tenant `json:"tenant" gorm:"foreignKey:TenantId;references:Id"`
	User   User   `json:"user" gorm:"foreignKey:UserId;references:Id"`
}

func (tm TenantMember) TableName() string {
	return "tenant_members"
}
