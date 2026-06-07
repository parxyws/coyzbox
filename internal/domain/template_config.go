package domain

import (
	"database/sql"
	"encoding/json"
	"time"
)

type TemplateConfig struct {
	Id       string          `json:"id" gorm:"column:id;primaryKey"`
	TenantId string          `json:"tenant_id" gorm:"column:tenant_id;index"`
	BaseType DocumentType    `json:"base_type" gorm:"column:base_type"`
	Status   string          `json:"status" gorm:"column:status;default:draft"`
	Name     string          `json:"name" gorm:"column:name"`
	Config   json.RawMessage `json:"config" gorm:"column:config;type:jsonb;default:'{}'"`

	CreatedAt time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time    `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" gorm:"column:deleted_at"`

	Tenant Tenant `json:"tenant" gorm:"foreignKey:TenantId;references:Id"`
}

func (tc TemplateConfig) TableName() string {
	return "template_configs"
}
