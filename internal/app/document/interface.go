package document

import (
	"context"
	"time"

	"github.com/parxyws/cozybox/internal/domain"
)

type FileStorage interface {
	PutObject(ctx context.Context, input domain.UploadInput) (key string, err error)
}

// Engine defines the deep interface for the Document Two-Phase Commit & Lifecycle Engine.
type Engine interface {
	ReserveSequenceAndDraft(ctx context.Context, tenantID string, userID string, req *CreateDraftRequest) (*DocumentResponse, error)
	PublishArtifact(ctx context.Context, tenantID string, userID string, draftID string, pdfBytes []byte) (*DocumentResponse, error)
	IssueDocument(ctx context.Context, tenantID string, userID string, docID string, issueDate *time.Time) (*DocumentResponse, error)
	TransitionFlowStatus(ctx context.Context, tenantID string, userID string, docID string, newFlow domain.DocumentFlowStatus) (*DocumentResponse, error)
	RecordPayment(ctx context.Context, tenantID string, userID string, docID string, req *RecordPaymentRequest) (*DocumentResponse, error)
	ConvertDocument(ctx context.Context, tenantID string, userID string, sourceDocID string, targetType domain.DocumentType) (*DocumentResponse, error)
	CancelDocument(ctx context.Context, tenantID string, userID string, docID string, reason string) (*DocumentResponse, error)
	GetDocument(ctx context.Context, tenantID string, docID string) (*DocumentResponse, error)
	DeleteDocument(ctx context.Context, tenantID string, docID string) error
	ListDocuments(ctx context.Context, tenantID string, docType string, status string, page int, limit int) ([]DocumentResponse, int64, error)
}
