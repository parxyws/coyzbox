package document_test

import (
	"context"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/parxyws/cozybox/internal/app/document"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/store"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memoryUserStore struct{}

func (s *memoryUserStore) InsertUser(ctx context.Context, u *domain.User) error { return nil }
func (s *memoryUserStore) UpdateUser(ctx context.Context, u *domain.User) error { return nil }
func (s *memoryUserStore) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, nil
}
func (s *memoryUserStore) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, nil
}
func (s *memoryUserStore) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (s *memoryUserStore) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return nil, nil
}

type memoryTenantStore struct {
	tmpl *domain.TemplateConfig
}

func (s *memoryTenantStore) InsertTenant(ctx context.Context, t *domain.Tenant) error { return nil }
func (s *memoryTenantStore) UpdateTenant(ctx context.Context, t *domain.Tenant) error { return nil }
func (s *memoryTenantStore) GetTenantByID(ctx context.Context, id string) (*domain.Tenant, error) {
	return nil, nil
}
func (s *memoryTenantStore) GetTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	return nil, nil
}
func (s *memoryTenantStore) RegisterTenantOwner(ctx context.Context, user *domain.User, tenant *domain.Tenant, member *domain.TenantMember, org *domain.Organization) error {
	return nil
}
func (s *memoryTenantStore) InsertTenantMember(ctx context.Context, tm *domain.TenantMember) error {
	return nil
}
func (s *memoryTenantStore) GetTenantMemberByUserID(ctx context.Context, userID string) (*domain.TenantMember, error) {
	return nil, nil
}
func (s *memoryTenantStore) ListTenantMembersByUserID(ctx context.Context, userID string) ([]domain.TenantMember, error) {
	return nil, nil
}
func (s *memoryTenantStore) ListTemplateConfigsByTenantID(ctx context.Context, tenantID string) ([]domain.TemplateConfig, error) {
	if s.tmpl != nil {
		return []domain.TemplateConfig{*s.tmpl}, nil
	}
	return nil, nil
}
func (s *memoryTenantStore) GetTemplateConfigByID(ctx context.Context, id string) (*domain.TemplateConfig, error) {
	if s.tmpl != nil && s.tmpl.Id == id {
		return s.tmpl, nil
	}
	return nil, store.ErrNotFound
}
func (s *memoryTenantStore) UpdateTemplateConfig(ctx context.Context, cfg *domain.TemplateConfig) error {
	s.tmpl = cfg
	return nil
}

type memoryDocStore struct {
	docs     map[string]*domain.Document
	payments map[string][]domain.DocumentPayment
}

func newMemoryDocStore() *memoryDocStore {
	return &memoryDocStore{
		docs:     make(map[string]*domain.Document),
		payments: make(map[string][]domain.DocumentPayment),
	}
}

func (r *memoryDocStore) InsertDocument(ctx context.Context, doc *domain.Document) error {
	r.docs[doc.Id] = doc
	return nil
}

func (r *memoryDocStore) UpdateDocument(ctx context.Context, doc *domain.Document) error {
	r.docs[doc.Id] = doc
	return nil
}

func (r *memoryDocStore) GetDocumentByIDAndTenant(ctx context.Context, id string, tenantID string) (*domain.Document, error) {
	doc, exists := r.docs[id]
	if !exists || doc.TenantId != tenantID {
		return nil, domain.ErrNotFound
	}
	return doc, nil
}

func (r *memoryDocStore) SoftDeleteDocument(ctx context.Context, id string, tenantID string) error {
	if doc, exists := r.docs[id]; exists && doc.TenantId == tenantID {
		now := time.Now()
		doc.DeletedAt.Time = now
		doc.DeletedAt.Valid = true
	}
	return nil
}

func (r *memoryDocStore) ListDocumentsByTenant(ctx context.Context, tenantID string, docType string, status string, page int, limit int) ([]domain.Document, int64, error) {
	var result []domain.Document
	for _, doc := range r.docs {
		if doc.TenantId == tenantID && !doc.DeletedAt.Valid {
			result = append(result, *doc)
		}
	}
	return result, int64(len(result)), nil
}

func (r *memoryDocStore) InsertPayment(ctx context.Context, payment *domain.DocumentPayment) error {
	r.payments[payment.DocumentId] = append(r.payments[payment.DocumentId], *payment)
	return nil
}

func (r *memoryDocStore) ListPaymentsByDocumentID(ctx context.Context, tenantID string, documentID string) ([]domain.DocumentPayment, error) {
	return r.payments[documentID], nil
}

func (r *memoryDocStore) RecordPaymentAndUpdateStatus(ctx context.Context, payment *domain.DocumentPayment, updatedDoc *domain.Document, activity *domain.DocumentActivity) error {
	if payment != nil {
		r.payments[payment.DocumentId] = append(r.payments[payment.DocumentId], *payment)
	}
	if updatedDoc != nil {
		r.docs[updatedDoc.Id] = updatedDoc
	}
	return nil
}

type memorySeqStore struct {
	counter int
}

func (r *memorySeqStore) GetSequenceSetting(ctx context.Context, tenantID string, docType domain.DocumentType) (*domain.TenantSequenceSetting, error) {
	return &domain.TenantSequenceSetting{
		PatternTemplate: "{{TYPE}}/2026/08/{{SEQ:4}}",
	}, nil
}

func (r *memorySeqStore) ReserveNextSequence(ctx context.Context, tenantID string, docType domain.DocumentType) (string, error) {
	r.counter++
	return "INV/2026/08/0001", nil
}

type memoryActStore struct {
	activities []domain.DocumentActivity
}

func (r *memoryActStore) InsertActivity(ctx context.Context, activity *domain.DocumentActivity) error {
	r.activities = append(r.activities, *activity)
	return nil
}

func (r *memoryActStore) ListActivitiesByDocumentID(ctx context.Context, documentID string) ([]domain.DocumentActivity, error) {
	return r.activities, nil
}

type memoryContactStore struct {
	contacts map[string]*domain.Contact
}

func (r *memoryContactStore) GetContactByIDAndTenant(ctx context.Context, id string, tenantID string) (*domain.Contact, error) {
	if c, ok := r.contacts[id]; ok {
		return c, nil
	}
	return nil, domain.ErrNotFound
}

func (r *memoryContactStore) InsertContact(ctx context.Context, contact *domain.Contact) error {
	return nil
}
func (r *memoryContactStore) UpdateContact(ctx context.Context, contact *domain.Contact) error {
	return nil
}
func (r *memoryContactStore) SoftDeleteContact(ctx context.Context, id string, tenantID string) error {
	return nil
}
func (r *memoryContactStore) ListContacts(ctx context.Context, tenantID string, search string, role string, page int, limit int) ([]domain.Contact, int64, error) {
	return nil, 0, nil
}

type memoryOrgStore struct {
	org *domain.Organization
}

func (r *memoryOrgStore) InsertOrganization(ctx context.Context, org *domain.Organization) error {
	r.org = org
	return nil
}

func (r *memoryOrgStore) UpdateOrganization(ctx context.Context, org *domain.Organization) error {
	r.org = org
	return nil
}

func (r *memoryOrgStore) GetOrganizationByTenantID(ctx context.Context, tenantID string) (*domain.Organization, error) {
	return r.org, nil
}

type memoryStorage struct{}

func (s *memoryStorage) PutObject(ctx context.Context, input domain.UploadInput) (string, error) {
	return input.ObjectName, nil
}

type memoryStore struct {
	user     store.UserStore
	tenant   store.TenantStore
	org      store.OrganizationStore
	contact  store.ContactStore
	doc      store.DocumentStore
	sequence store.SequenceStore
	activity store.ActivityStore
}

func (s *memoryStore) User() store.UserStore                 { return s.user }
func (s *memoryStore) Tenant() store.TenantStore             { return s.tenant }
func (s *memoryStore) Organization() store.OrganizationStore { return s.org }
func (s *memoryStore) Contact() store.ContactStore           { return s.contact }
func (s *memoryStore) Document() store.DocumentStore         { return s.doc }
func (s *memoryStore) Sequence() store.SequenceStore         { return s.sequence }
func (s *memoryStore) Activity() store.ActivityStore         { return s.activity }

func TestDocumentEngine_TwoPhaseCommitAndPaymentFlow(t *testing.T) {
	ctx := context.Background()
	tenantID := ulid.Make().String()
	userID := ulid.Make().String()
	contactID := ulid.Make().String()

	org := &domain.Organization{
		Id:              ulid.Make().String(),
		TenantId:        tenantID,
		Name:            "Cozybox Test Corp",
		DefaultCurrency: "USD",
	}

	contact := &domain.Contact{
		Id:       contactID,
		TenantId: tenantID,
		Name:     "Acme Client",
		Roles:    []string{"client"},
	}

	docStore := newMemoryDocStore()
	seqStore := &memorySeqStore{}
	actStore := &memoryActStore{}
	contactStore := &memoryContactStore{contacts: map[string]*domain.Contact{contactID: contact}}
	orgStore := &memoryOrgStore{org: org}
	storage := &memoryStorage{}

	appStore := &memoryStore{
		user:     &memoryUserStore{},
		tenant:   &memoryTenantStore{},
		doc:      docStore,
		sequence: seqStore,
		activity: actStore,
		contact:  contactStore,
		org:      orgStore,
	}

	engine := document.NewService(appStore, storage)

	// Phase 1: Draft
	req := &document.CreateDraftRequest{
		Type:           domain.DocumentTypeInvoice,
		ContactID:      &contactID,
		Currency:       "USD",
		DiscountAmount: decimal.NewFromInt(10),
		Items: []document.CreateItemRequest{
			{
				SortOrder:   1,
				Description: "Web Services",
				Quantity:    decimal.NewFromInt(10),
				UnitPrice:   decimal.NewFromInt(100),
				DiscountPct: decimal.Zero,
				TaxPct:      decimal.NewFromInt(10),
			},
		},
	}

	draft, err := engine.ReserveSequenceAndDraft(ctx, tenantID, userID, req)
	require.NoError(t, err)
	assert.Equal(t, domain.DocumentStatusDraft, draft.Status)

	// Phase 2: Publish
	fakePdf := []byte("%PDF-1.4 Fake PDF Content")
	pub, err := engine.PublishArtifact(ctx, tenantID, userID, draft.ID, fakePdf)
	require.NoError(t, err)
	assert.Equal(t, domain.DocumentStatusPublished, pub.Status)

	// Issue Document
	issued, err := engine.IssueDocument(ctx, tenantID, userID, draft.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, domain.DocumentFlowStatusIssued, *issued.FlowStatus)

	// Record Partial Payment
	payResp1, err := engine.RecordPayment(ctx, tenantID, userID, draft.ID, &document.RecordPaymentRequest{
		Amount:        decimal.NewFromInt(500),
		PaymentMethod: "bank_transfer",
		ReferenceNo:   "REF-001",
	})
	require.NoError(t, err)
	assert.Equal(t, domain.DocumentFlowStatusPartiallyPaid, *payResp1.FlowStatus)
	assert.True(t, payResp1.AmountPaid.Equal(decimal.NewFromInt(500)))

	// Record Full Payment
	payResp2, err := engine.RecordPayment(ctx, tenantID, userID, draft.ID, &document.RecordPaymentRequest{
		Amount:        decimal.NewFromInt(590),
		PaymentMethod: "bank_transfer",
		ReferenceNo:   "REF-002",
	})
	require.NoError(t, err)
	assert.Equal(t, domain.DocumentFlowStatusPaid, *payResp2.FlowStatus)
	assert.True(t, payResp2.AmountPaid.Equal(decimal.NewFromInt(1090)))
}

func TestDocumentEngine_ConversionLineage(t *testing.T) {
	ctx := context.Background()
	tenantID := ulid.Make().String()
	userID := ulid.Make().String()
	contactID := ulid.Make().String()

	org := &domain.Organization{
		Id:              ulid.Make().String(),
		TenantId:        tenantID,
		Name:            "Cozybox Test Corp",
		DefaultCurrency: "USD",
	}

	contact := &domain.Contact{
		Id:       contactID,
		TenantId: tenantID,
		Name:     "Acme Client",
		Roles:    []string{"client"},
	}

	docStore := newMemoryDocStore()
	seqStore := &memorySeqStore{}
	actStore := &memoryActStore{}
	contactStore := &memoryContactStore{contacts: map[string]*domain.Contact{contactID: contact}}
	orgStore := &memoryOrgStore{org: org}
	storage := &memoryStorage{}

	appStore := &memoryStore{
		user:     &memoryUserStore{},
		tenant:   &memoryTenantStore{},
		doc:      docStore,
		sequence: seqStore,
		activity: actStore,
		contact:  contactStore,
		org:      orgStore,
	}

	engine := document.NewService(appStore, storage)

	// Create & Publish Quotation
	quotationReq := &document.CreateDraftRequest{
		Type:      domain.DocumentTypeQuotation,
		ContactID: &contactID,
		Items: []document.CreateItemRequest{
			{Description: "Design Consulting", Quantity: decimal.NewFromInt(1), UnitPrice: decimal.NewFromInt(1000)},
		},
	}
	qDraft, err := engine.ReserveSequenceAndDraft(ctx, tenantID, userID, quotationReq)
	require.NoError(t, err)
	qPub, err := engine.PublishArtifact(ctx, tenantID, userID, qDraft.ID, []byte("pdf-bytes"))
	require.NoError(t, err)

	// Convert Quotation -> Invoice
	invDraft, err := engine.ConvertDocument(ctx, tenantID, userID, qPub.ID, domain.DocumentTypeInvoice)
	require.NoError(t, err)
	assert.Equal(t, domain.DocumentTypeInvoice, invDraft.Type)
}
