package models

import "time"

// DocumentSequence manages auto-incrementing document reference numbers
// per organization and document type.
type DocumentSequence struct {
	Id             string       `json:"id" gorm:"column:id;primaryKey"`
	TenantId       string       `json:"tenant_id" gorm:"column:tenant_id;index"`
	OrganizationId string       `json:"organization_id" gorm:"column:organization_id"`
	Type           DocumentType `json:"type" gorm:"column:type"`
	Prefix         string       `json:"prefix" gorm:"column:prefix"`
	NextNumber     int          `json:"next_number" gorm:"column:next_number;default:1"`
	Format         string       `json:"format" gorm:"column:format;default:{PREFIX}-{YEAR}-{SEQ:4}"`
	LastResetAt    *time.Time   `json:"last_reset_at" gorm:"column:last_reset_at"`
	CreatedAt      time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time    `json:"updated_at" gorm:"column:updated_at"`
}

func (ds DocumentSequence) TableName() string {
	return "document_sequences"
}
