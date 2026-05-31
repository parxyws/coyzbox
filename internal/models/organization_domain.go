package models

import (
	"database/sql"
	"time"
)

type Organization struct {
	Id              string       `json:"id" gorm:"column:id;primaryKey"`
	TenantId        string       `json:"tenant_id" gorm:"column:tenant_id;uniqueIndex"`
	Name            string       `json:"name" gorm:"column:name"`
	Email           string       `json:"email" gorm:"column:email"`
	Phone           string       `json:"phone" gorm:"column:phone"`
	AddressLine1    string       `json:"address_line1" gorm:"column:address_line1"`
	AddressLine2    string       `json:"address_line2" gorm:"column:address_line2"`
	City            string       `json:"city" gorm:"column:city"`
	State           string       `json:"state" gorm:"column:state"`
	PostalCode      string       `json:"postal_code" gorm:"column:postal_code"`
	Country         string       `json:"country" gorm:"column:country"`
	TaxId           string       `json:"tax_id" gorm:"column:tax_id"`
	LogoS3Key       string       `json:"logo_s3_key" gorm:"column:logo_s3_key"`
	Website         string       `json:"website" gorm:"column:website"`
	Timezone        string       `json:"timezone" gorm:"column:timezone"`
	DefaultCurrency string       `json:"default_currency" gorm:"column:default_currency;default:USD"`
	CreatedAt       time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       time.Time    `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt       sql.NullTime `json:"deleted_at" gorm:"column:deleted_at"`

	// Relations
	Tenant   *Tenant   `json:"tenant" gorm:"foreignKey:TenantId;references:Id"`
	Contacts []Contact `json:"contacts" gorm:"foreignKey:OrganizationId;references:Id"`
}

func (o Organization) TableName() string {
	return "organizations"
}
