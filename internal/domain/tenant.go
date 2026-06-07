package domain

import (
	"database/sql"
	"time"
)

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusInActive  TenantStatus = "inactive"
)

// Tenant is the top-level isolation boundary for the entire SaaS system.
// Every data access path (documents, contacts, sequences, organizations)
// is scoped by tenant_id, enforced through middleware and GORM scopes.
type Tenant struct {
	Id        string       `json:"id" gorm:"column:id;primaryKey"`
	Name      string       `json:"name" gorm:"column:name"`
	Slug      string       `json:"slug" gorm:"column:slug;uniqueIndex"`
	Status    TenantStatus `json:"status" gorm:"column:status;default:active"`
	OwnerId   string       `json:"owner_id" gorm:"column:owner_id"`
	CreatedAt time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time    `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" gorm:"column:deleted_at"`

	// Relations
	Owner        User           `json:"owner" gorm:"foreignKey:OwnerId;references:Id"`
	Members      []TenantMember `json:"members" gorm:"foreignKey:TenantId;references:Id"`
	Organization *Organization  `json:"organization" gorm:"foreignKey:TenantId;references:Id"`
}

func (t Tenant) TableName() string {
	return "tenants"
}
