package domain

import (
	"encoding/json"
	"time"
)

// DocumentActivity records every state change on a document for audit purposes.
//
// The 'action' field identifies what kind of change occurred. Conventions:
//   - "status_changed"     — lifecycle status changed (status column)
//   - "flow_status_changed" — business flow status changed (flow_status column)
//   - "published"          — document published, document_ref assigned
//   - "payment_recorded"   — amount_paid updated
//   - "cancelled"          — document voided
//
// FromStatus and ToStatus track the old and new values as plain strings so that
// both the lifecycle status and the flow status can be recorded using a single
// activity row without requiring separate column pairs.
type DocumentActivity struct {
	Id          string          `json:"id" gorm:"column:id;primaryKey"`
	DocumentId  string          `json:"document_id" gorm:"column:document_id"`
	Action      string          `json:"action" gorm:"column:action"`
	FromStatus  *string         `json:"from_status" gorm:"column:from_status"`
	ToStatus    *string         `json:"to_status" gorm:"column:to_status"`
	PerformedBy *string         `json:"performed_by" gorm:"column:performed_by"`
	Note        string          `json:"note" gorm:"column:note"`
	Metadata    json.RawMessage `json:"metadata" gorm:"column:metadata;type:jsonb"`
	CreatedAt   time.Time       `json:"created_at" gorm:"column:created_at"`
}

func (da DocumentActivity) TableName() string {
	return "document_activities"
}
