package domain

import "time"

type SequenceResetCycle string

const (
	SequenceResetMonthly SequenceResetCycle = "monthly"
	SequenceResetYearly  SequenceResetCycle = "yearly"
	SequenceResetNever   SequenceResetCycle = "never"
)

type TenantSequenceSetting struct {
	Id              string             `json:"id" gorm:"column:id;primaryKey"`
	TenantId        string             `json:"tenant_id" gorm:"column:tenant_id;uniqueIndex:idx_tenant_type"`
	DocumentType    DocumentType       `json:"document_type" gorm:"column:document_type;uniqueIndex:idx_tenant_type"`
	PatternTemplate string             `json:"pattern_template" gorm:"column:pattern_template"`
	ResetCycle      SequenceResetCycle `json:"reset_cycle" gorm:"column:reset_cycle"`
	CreatedAt       time.Time          `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       time.Time          `json:"updated_at" gorm:"column:updated_at"`
}

func (ts TenantSequenceSetting) TableName() string {
	return "tenant_sequence_settings"
}

type TenantSequenceState struct {
	TenantId      string       `json:"tenant_id" gorm:"column:tenant_id;primaryKey"`
	DocumentType  DocumentType `json:"document_type" gorm:"column:document_type;primaryKey"`
	CurrentPeriod string       `json:"current_period" gorm:"column:current_period;primaryKey"`
	LastValue     int          `json:"last_value" gorm:"column:last_value;default:0"`
	UpdatedAt     time.Time    `json:"updated_at" gorm:"column:updated_at"`
}

func (ts TenantSequenceState) TableName() string {
	return "tenant_sequence_state"
}
