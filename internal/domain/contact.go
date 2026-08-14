package domain

import (
	"database/sql"
	"time"
)

type ContactRole string

const (
	ContactRoleClient   ContactRole = "client"
	ContactRoleSupplier ContactRole = "supplier"
)

type Contact struct {
	Id             string       `json:"id" gorm:"column:id;primaryKey"`
	TenantId       string       `json:"tenant_id" gorm:"column:tenant_id;index"`
	OrganizationId string       `json:"organization_id" gorm:"column:organization_id"`
	Roles          []string     `json:"roles" gorm:"column:roles;serializer:json"`
	Name           string       `json:"name" gorm:"column:name"`
	Email          string       `json:"email" gorm:"column:email"`
	Phone          string       `json:"phone" gorm:"column:phone"`
	AddressLine1   string       `json:"address_line1" gorm:"column:address_line1"`
	AddressLine2   string       `json:"address_line2" gorm:"column:address_line2"`
	City           string       `json:"city" gorm:"column:city"`
	State          string       `json:"state" gorm:"column:state"`
	PostalCode     string       `json:"postal_code" gorm:"column:postal_code"`
	Country        string       `json:"country" gorm:"column:country"`
	TaxId          string       `json:"tax_id" gorm:"column:tax_id"`
	Notes          string       `json:"notes" gorm:"column:notes"`
	CreatedAt      time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time    `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt      sql.NullTime `json:"deleted_at" gorm:"column:deleted_at"`

	Organization Organization `json:"organization" gorm:"foreignKey:OrganizationId;references:Id"`
}

func (c Contact) TableName() string {
	return "contacts"
}

func (c Contact) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (c Contact) IsClient() bool {
	return c.HasRole(string(ContactRoleClient))
}

func (c Contact) IsSupplier() bool {
	return c.HasRole(string(ContactRoleSupplier))
}
