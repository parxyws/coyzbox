package domain

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// =============================================================================
// Document Type
// =============================================================================

type DocumentType string

const (
	DocumentTypeQuotation     DocumentType = "quotation"
	DocumentTypeInvoice       DocumentType = "invoice"
	DocumentTypeReceipt       DocumentType = "receipt"
	DocumentTypePurchaseOrder DocumentType = "purchase_order"
	DocumentTypeSalesOrder    DocumentType = "sales_order"
	DocumentTypeDebitNote     DocumentType = "debit_note"
)

// =============================================================================
// Document Status — lifecycle state (maps to the 'status' column)
//
// Tracks where the document is in its publication lifecycle:
//   draft → published → cancelled
//   published → expired  (cron: past valid_until)
// =============================================================================

type DocumentStatus string

const (
	DocumentStatusDraft     DocumentStatus = "draft"
	DocumentStatusPublished DocumentStatus = "published"
	DocumentStatusCancelled DocumentStatus = "cancelled"
	DocumentStatusExpired   DocumentStatus = "expired"
)

// IsEditable returns true only when a document's fields may still be changed.
func (s DocumentStatus) IsEditable() bool {
	return s == DocumentStatusDraft
}

// IsTerminalStatus returns true when the document lifecycle has permanently ended.
func (s DocumentStatus) IsTerminalStatus() bool {
	return s == DocumentStatusCancelled
}

// CanPublish returns true if the document is in a state that allows publishing.
func (s DocumentStatus) CanPublish() bool {
	return s == DocumentStatusDraft
}

// =============================================================================
// Document Flow Status — business cycle state (maps to the 'flow_status' column)
//
// Tracks the recipient/payment lifecycle once a document is published:
//   nil (not yet in a business flow)
//   → accepted / rejected
//   → overdue (cron: past due_date without full payment)
//   → paid / partially_paid
// =============================================================================

type DocumentFlowStatus string

const (
	DocumentFlowStatusAccepted      DocumentFlowStatus = "accepted"
	DocumentFlowStatusRejected      DocumentFlowStatus = "rejected"
	DocumentFlowStatusOverdue       DocumentFlowStatus = "overdue"
	DocumentFlowStatusPaid          DocumentFlowStatus = "paid"
	DocumentFlowStatusPartiallyPaid DocumentFlowStatus = "partially_paid"
)

// IsTerminalFlow returns true when the business cycle has permanently ended
// and no further flow transitions are possible.
func (f DocumentFlowStatus) IsTerminalFlow() bool {
	return f == DocumentFlowStatusRejected || f == DocumentFlowStatusPaid
}

// AllowedFlowTransitions returns the valid next flow states from the current one.
// Only manual user transitions are listed here; cron transitions (overdue) are
// applied directly by the worker without going through this guard.
func (f DocumentFlowStatus) AllowedFlowTransitions() []DocumentFlowStatus {
	switch f {
	case DocumentFlowStatusAccepted:
		return []DocumentFlowStatus{
			DocumentFlowStatusPaid,
			DocumentFlowStatusPartiallyPaid,
			DocumentFlowStatusOverdue,
		}
	case DocumentFlowStatusOverdue:
		return []DocumentFlowStatus{
			DocumentFlowStatusPaid,
			DocumentFlowStatusPartiallyPaid,
		}
	case DocumentFlowStatusPartiallyPaid:
		return []DocumentFlowStatus{
			DocumentFlowStatusPaid,
			DocumentFlowStatusOverdue,
		}
	default:
		// rejected, paid, and nil have no further transitions
		return nil
	}
}

// AllowedInitialFlowStatuses returns the flow statuses that can be set directly
// from a published document that has no existing flow status yet (nil).
func AllowedInitialFlowStatuses() []DocumentFlowStatus {
	return []DocumentFlowStatus{
		DocumentFlowStatusAccepted,
		DocumentFlowStatusRejected,
		DocumentFlowStatusPaid,
		DocumentFlowStatusPartiallyPaid,
	}
}

// =============================================================================
// Document Entity
// =============================================================================

type Document struct {
	Id             string  `json:"id" gorm:"column:id;primaryKey"`
	TenantId       string  `json:"tenant_id" gorm:"column:tenant_id;index"`
	OrganizationId string  `json:"organization_id" gorm:"column:organization_id"`
	ContactId      *string `json:"contact_id" gorm:"column:contact_id"`
	ParentId       *string `json:"parent_id" gorm:"column:parent_id"`

	// Classification
	Type        DocumentType        `json:"type" gorm:"column:type"`
	Status      DocumentStatus      `json:"status" gorm:"column:status;default:draft"`
	FlowStatus  *DocumentFlowStatus `json:"flow_status" gorm:"column:flow_status"`
	DocumentRef string              `json:"document_ref" gorm:"column:document_ref"`

	// Dates
	IssueDate  sql.NullTime `json:"issue_date" gorm:"column:issue_date"`
	DueDate    sql.NullTime `json:"due_date" gorm:"column:due_date"`
	ValidUntil sql.NullTime `json:"valid_until" gorm:"column:valid_until"`

	// Financial
	Currency       string          `json:"currency" gorm:"column:currency;default:USD"`
	Subtotal       decimal.Decimal `json:"subtotal" gorm:"column:subtotal;default:0"`
	DiscountAmount decimal.Decimal `json:"discount_amount" gorm:"column:discount_amount;default:0"`
	TaxAmount      decimal.Decimal `json:"tax_amount" gorm:"column:tax_amount;default:0"`
	Total          decimal.Decimal `json:"total" gorm:"column:total;default:0"`
	AmountPaid     decimal.Decimal `json:"amount_paid" gorm:"column:amount_paid;default:0"`

	// Content
	Notes  string `json:"notes" gorm:"column:notes"`
	Terms  string `json:"terms" gorm:"column:terms"`
	Footer string `json:"footer" gorm:"column:footer"`

	Metadata json.RawMessage `json:"metadata" gorm:"column:metadata;type:jsonb;default:'{}'"`

	// File storage
	PdfS3Key       string       `json:"pdf_s3_key" gorm:"column:pdf_s3_key"`
	PdfGeneratedAt sql.NullTime `json:"pdf_generated_at" gorm:"column:pdf_generated_at"`
	PdfSizeBytes   int          `json:"pdf_size_bytes" gorm:"column:pdf_size_bytes"`

	// Audit
	CreatedBy string       `json:"created_by" gorm:"column:created_by"`
	CreatedAt time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time    `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" gorm:"column:deleted_at"`

	// Associations
	Tenant       Tenant             `json:"tenant" gorm:"foreignKey:TenantId;references:Id"`
	Organization Organization       `json:"organization" gorm:"foreignKey:OrganizationId;references:Id"`
	Contact      *Contact           `json:"contact" gorm:"foreignKey:ContactId;references:Id"`
	Parent       *Document          `json:"parent" gorm:"foreignKey:ParentId;references:Id"`
	Children     []Document         `json:"children" gorm:"foreignKey:ParentId;references:Id"`
	Items        []DocumentItem     `json:"items" gorm:"foreignKey:DocumentId;references:Id"`
	Activities   []DocumentActivity `json:"activities" gorm:"foreignKey:DocumentId;references:Id"`
	Creator      User               `json:"creator" gorm:"foreignKey:CreatedBy;references:Id"`
}

func (d Document) TableName() string {
	return "documents"
}

// IsFullySettled returns true when the total amount has been paid.
func (d Document) IsFullySettled() bool {
	return d.AmountPaid.GreaterThanOrEqual(d.Total)
}

// OutstandingAmount returns the remaining unpaid balance.
func (d Document) OutstandingAmount() decimal.Decimal {
	return d.Total.Sub(d.AmountPaid)
}

// IsDeletable returns true when the document can be soft-deleted.
// Only drafts and terminally-closed documents may be removed.
func (d Document) IsDeletable() bool {
	if d.Status == DocumentStatusDraft {
		return true
	}
	if d.Status == DocumentStatusCancelled {
		return true
	}
	if d.FlowStatus != nil && d.FlowStatus.IsTerminalFlow() {
		return true
	}
	return false
}

// CanTransitionFlow returns whether transitioning to the given flow status is
// valid given the document's current status and flow_status.
func (d Document) CanTransitionFlow(next DocumentFlowStatus) bool {
	// Flow transitions only apply to published documents.
	if d.Status != DocumentStatusPublished {
		return false
	}

	// No existing flow status: only the initial allowed set is valid.
	if d.FlowStatus == nil {
		for _, allowed := range AllowedInitialFlowStatuses() {
			if next == allowed {
				return true
			}
		}
		return false
	}

	// Terminal flow states cannot transition further.
	if d.FlowStatus.IsTerminalFlow() {
		return false
	}

	for _, allowed := range d.FlowStatus.AllowedFlowTransitions() {
		if next == allowed {
			return true
		}
	}
	return false
}
