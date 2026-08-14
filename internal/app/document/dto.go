package document

import (
	"time"

	"github.com/parxyws/cozybox/internal/domain"
	"github.com/shopspring/decimal"
)

type CreateItemRequest struct {
	SortOrder   int             `json:"sort_order"`
	Description string          `json:"description" binding:"required"`
	Quantity    decimal.Decimal `json:"quantity" binding:"required"`
	Unit        string          `json:"unit"`
	UnitPrice   decimal.Decimal `json:"unit_price" binding:"required"`
	DiscountPct decimal.Decimal `json:"discount_pct"`
	TaxPct      decimal.Decimal `json:"tax_pct"`
}

type CreateDraftRequest struct {
	Type             domain.DocumentType `json:"type" binding:"required"`
	TemplateConfigID *string             `json:"template_config_id"`
	ContactID        *string             `json:"contact_id"`
	IssueDate        *time.Time          `json:"issue_date"`
	DueDate          *time.Time          `json:"due_date"`
	ValidUntil       *time.Time          `json:"valid_until"`
	Currency         string              `json:"currency"`
	Notes            string              `json:"notes"`
	Terms            string              `json:"terms"`
	Footer           string              `json:"footer"`
	DiscountAmount   decimal.Decimal     `json:"discount_amount"`
	Items            []CreateItemRequest `json:"items" binding:"required,gt=0"`
}

type PublishArtifactRequest struct {
	PdfBase64 string `json:"pdf_base64" binding:"required"`
}

type TransitionFlowStatusRequest struct {
	FlowStatus domain.DocumentFlowStatus `json:"flow_status" binding:"required"`
}

type CancelDocumentRequest struct {
	Reason string `json:"reason"`
}

type RecordPaymentRequest struct {
	Amount        decimal.Decimal `json:"amount" binding:"required"`
	PaymentMethod string          `json:"payment_method" binding:"required"`
	PaymentDate   *time.Time      `json:"payment_date"`
	ReferenceNo   string          `json:"reference_no"`
	Notes         string          `json:"notes"`
}

type ConvertDocumentRequest struct {
	TargetType domain.DocumentType `json:"target_type" binding:"required"`
}

type DocumentItemResponse struct {
	ID          string          `json:"id"`
	SortOrder   int             `json:"sort_order"`
	Description string          `json:"description"`
	Quantity    decimal.Decimal `json:"quantity"`
	Unit        string          `json:"unit"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	DiscountPct decimal.Decimal `json:"discount_pct"`
	TaxPct      decimal.Decimal `json:"tax_pct"`
	Amount      decimal.Decimal `json:"amount"`
}

type DocumentActivityResponse struct {
	ID          string    `json:"id"`
	Action      string    `json:"action"`
	FromStatus  *string   `json:"from_status,omitempty"`
	ToStatus    *string   `json:"to_status,omitempty"`
	PerformedBy *string   `json:"performed_by,omitempty"`
	Note        string    `json:"note"`
	CreatedAt   time.Time `json:"created_at"`
}

type DocumentPaymentResponse struct {
	ID            string          `json:"id"`
	Amount        decimal.Decimal `json:"amount"`
	PaymentMethod string          `json:"payment_method"`
	PaymentDate   time.Time       `json:"payment_date"`
	ReferenceNo   string          `json:"reference_no"`
	Notes         string          `json:"notes"`
	CreatedBy     string          `json:"created_by"`
	CreatedAt     time.Time       `json:"created_at"`
}

type DocumentResponse struct {
	ID             string                     `json:"id"`
	TenantID       string                     `json:"tenant_id"`
	OrganizationID string                     `json:"organization_id"`
	ContactID      *string                    `json:"contact_id,omitempty"`
	ParentID       *string                    `json:"parent_id,omitempty"`
	Type           domain.DocumentType        `json:"type"`
	Status         domain.DocumentStatus      `json:"status"`
	FlowStatus     *domain.DocumentFlowStatus `json:"flow_status,omitempty"`
	DocumentRef    string                     `json:"document_ref"`
	IssueDate      *time.Time                 `json:"issue_date,omitempty"`
	DueDate        *time.Time                 `json:"due_date,omitempty"`
	ValidUntil     *time.Time                 `json:"valid_until,omitempty"`
	Currency       string                     `json:"currency"`
	Subtotal       decimal.Decimal            `json:"subtotal"`
	DiscountAmount decimal.Decimal            `json:"discount_amount"`
	TaxAmount      decimal.Decimal            `json:"tax_amount"`
	Total          decimal.Decimal            `json:"total"`
	AmountPaid     decimal.Decimal            `json:"amount_paid"`
	Notes          string                     `json:"notes"`
	Terms          string                     `json:"terms"`
	Footer         string                     `json:"footer"`
	PdfS3Key       string                     `json:"pdf_s3_key,omitempty"`
	PdfGeneratedAt *time.Time                 `json:"pdf_generated_at,omitempty"`
	Items          []DocumentItemResponse     `json:"items,omitempty"`
	Activities     []DocumentActivityResponse `json:"activities,omitempty"`
	Payments       []DocumentPaymentResponse  `json:"payments,omitempty"`
	CreatedAt      time.Time                  `json:"created_at"`
	UpdatedAt      time.Time                  `json:"updated_at"`
}
