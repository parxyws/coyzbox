package models

import (
	"encoding/json"
	"time"
)

// DocumentActivity records lifecycle events for audit trail.
type DocumentActivity struct {
	Id          string              `json:"id" gorm:"column:id;primaryKey"`
	DocumentId  string              `json:"document_id" gorm:"column:document_id"`
	Action      string              `json:"action" gorm:"column:action"`
	FromStatus  *DocumentFlowStatus `json:"from_status" gorm:"column:from_status"`
	ToStatus    *DocumentFlowStatus `json:"to_status" gorm:"column:to_status"`
	PerformedBy *string             `json:"performed_by" gorm:"column:performed_by"`
	Note        string              `json:"note" gorm:"column:note"`
	Metadata    json.RawMessage     `json:"metadata" gorm:"column:metadata;type:jsonb"`
	CreatedAt   time.Time           `json:"created_at" gorm:"column:created_at"`
}

func (da DocumentActivity) TableName() string {
	return "document_activities"
}
