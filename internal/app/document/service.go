package document

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/store"
	"github.com/shopspring/decimal"
)

type Service struct {
	appStore    store.Store
	fileStorage FileStorage
}

func NewService(
	appStore store.Store,
	fileStorage FileStorage,
) *Service {
	return &Service{
		appStore:    appStore,
		fileStorage: fileStorage,
	}
}

func (s *Service) ReserveSequenceAndDraft(ctx context.Context, tenantID string, userID string, req *CreateDraftRequest) (*DocumentResponse, error) {
	var contact *domain.Contact
	if req.ContactID != nil && *req.ContactID != "" {
		c, err := s.appStore.Contact().GetContactByIDAndTenant(ctx, *req.ContactID, tenantID)
		if err != nil {
			return nil, fmt.Errorf("invalid contact: %w", err)
		}
		if err := validateContactRole(req.Type, c); err != nil {
			return nil, err
		}
		contact = c
	}

	if req.TemplateConfigID != nil && *req.TemplateConfigID != "" {
		tmpl, err := s.appStore.Tenant().GetTemplateConfigByID(ctx, *req.TemplateConfigID)
		if err != nil {
			return nil, fmt.Errorf("invalid template config: %w", err)
		}
		if tmpl.TenantId != tenantID {
			return nil, fmt.Errorf("template config belongs to another tenant")
		}
		if tmpl.Status != domain.TemplateStatusPublished {
			return nil, fmt.Errorf("template config must be published before creating documents")
		}
	}

	org, err := s.appStore.Organization().GetOrganizationByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve organization profile: %w", err)
	}

	now := time.Now()
	docID := ulid.Make().String()

	var subtotal decimal.Decimal
	var taxAmount decimal.Decimal
	items := make([]domain.DocumentItem, len(req.Items))

	for i, itemReq := range req.Items {
		item := domain.DocumentItem{
			Id:          ulid.Make().String(),
			DocumentId:  docID,
			SortOrder:   itemReq.SortOrder,
			Description: itemReq.Description,
			Quantity:    itemReq.Quantity,
			Unit:        itemReq.Unit,
			UnitPrice:   itemReq.UnitPrice,
			DiscountPct: itemReq.DiscountPct,
			TaxPct:      itemReq.TaxPct,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		item.Amount = item.CalculateAmount()
		subtotal = subtotal.Add(item.Quantity.Mul(item.UnitPrice))

		gross := item.Quantity.Mul(item.UnitPrice)
		disc := gross.Mul(item.DiscountPct).Div(decimal.NewFromInt(100))
		afterDisc := gross.Sub(disc)
		tax := afterDisc.Mul(item.TaxPct).Div(decimal.NewFromInt(100))
		taxAmount = taxAmount.Add(tax)

		items[i] = item
	}

	total := subtotal.Sub(req.DiscountAmount).Add(taxAmount)
	currency := req.Currency
	if currency == "" {
		currency = org.DefaultCurrency
	}
	if currency == "" {
		currency = "USD"
	}

	snapshotMap := map[string]any{
		"organization": org,
		"contact":      contact,
		"request":      req,
		"frozen_at":    now,
	}
	snapshotBytes, _ := json.Marshal(snapshotMap)

	docRef, err := s.appStore.Sequence().ReserveNextSequence(ctx, tenantID, req.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve sequence: %w", err)
	}

	doc := &domain.Document{
		Id:             docID,
		TenantId:       tenantID,
		OrganizationId: org.Id,
		ContactId:      req.ContactID,
		Type:           req.Type,
		Status:         domain.DocumentStatusDraft,
		DocumentRef:    docRef,
		Currency:       currency,
		Subtotal:       subtotal,
		DiscountAmount: req.DiscountAmount,
		TaxAmount:      taxAmount,
		Total:          total,
		AmountPaid:     decimal.Zero,
		Notes:          req.Notes,
		Terms:          req.Terms,
		Footer:         req.Footer,
		Metadata:       snapshotBytes,
		CreatedBy:      userID,
		CreatedAt:      now,
		UpdatedAt:      now,
		Items:          items,
	}

	if req.IssueDate != nil {
		doc.IssueDate = sql.NullTime{Time: *req.IssueDate, Valid: true}
	}
	if req.DueDate != nil {
		doc.DueDate = sql.NullTime{Time: *req.DueDate, Valid: true}
	}
	if req.ValidUntil != nil {
		doc.ValidUntil = sql.NullTime{Time: *req.ValidUntil, Valid: true}
	}

	if err := s.appStore.Document().InsertDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to insert document draft: %w", err)
	}

	toStatus := string(domain.DocumentStatusDraft)
	act := &domain.DocumentActivity{
		Id:          ulid.Make().String(),
		DocumentId:  doc.Id,
		Action:      "created",
		ToStatus:    &toStatus,
		PerformedBy: &userID,
		Note:        "Reserved sequence and created draft snapshot",
		CreatedAt:   now,
	}
	if err := s.appStore.Activity().InsertActivity(ctx, act); err != nil {
		return nil, fmt.Errorf("failed to log document activity: %w", err)
	}

	return toDocumentResponse(doc), nil
}

func (s *Service) PublishArtifact(ctx context.Context, tenantID string, userID string, draftID string, pdfBytes []byte) (*DocumentResponse, error) {
	doc, err := s.appStore.Document().GetDocumentByIDAndTenant(ctx, draftID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	if doc.Status != domain.DocumentStatusDraft {
		return nil, fmt.Errorf("only draft documents may be published, current status: %s", doc.Status)
	}

	s3Key := fmt.Sprintf("tenants/%s/documents/%s/%s.pdf", tenantID, doc.Type, doc.Id)
	uploadInput := domain.UploadInput{
		Object:      bytes.NewReader(pdfBytes),
		ObjectName:  s3Key,
		ObjectSize:  int64(len(pdfBytes)),
		ContentType: "application/pdf",
	}

	key, err := s.fileStorage.PutObject(ctx, uploadInput)
	if err != nil {
		return nil, fmt.Errorf("failed to upload PDF artifact: %w", err)
	}

	now := time.Now()
	fromStatus := string(domain.DocumentStatusDraft)
	toStatus := string(domain.DocumentStatusPublished)

	doc.Status = domain.DocumentStatusPublished
	doc.PdfS3Key = key
	doc.PdfGeneratedAt = sql.NullTime{Time: now, Valid: true}
	doc.PdfSizeBytes = len(pdfBytes)
	doc.UpdatedAt = now

	if err := s.appStore.Document().UpdateDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to publish document: %w", err)
	}

	act := &domain.DocumentActivity{
		Id:          ulid.Make().String(),
		DocumentId:  doc.Id,
		Action:      "published",
		FromStatus:  &fromStatus,
		ToStatus:    &toStatus,
		PerformedBy: &userID,
		Note:        "Compiled PDF artifact and published document",
		CreatedAt:   now,
	}
	if err := s.appStore.Activity().InsertActivity(ctx, act); err != nil {
		return nil, fmt.Errorf("failed to log document activity: %w", err)
	}

	return toDocumentResponse(doc), nil
}

func (s *Service) IssueDocument(ctx context.Context, tenantID string, userID string, docID string, issueDate *time.Time) (*DocumentResponse, error) {
	doc, err := s.appStore.Document().GetDocumentByIDAndTenant(ctx, docID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	if doc.Status != domain.DocumentStatusPublished {
		return nil, fmt.Errorf("only published documents can be issued")
	}

	if !doc.CanTransitionFlow(domain.DocumentFlowStatusIssued) {
		return nil, fmt.Errorf("cannot transition flow status to issued from current flow_status")
	}

	now := time.Now()
	if issueDate != nil {
		doc.IssueDate = sql.NullTime{Time: *issueDate, Valid: true}
	} else if !doc.IssueDate.Valid {
		doc.IssueDate = sql.NullTime{Time: now, Valid: true}
	}

	issuedFlow := domain.DocumentFlowStatusIssued
	doc.FlowStatus = &issuedFlow
	doc.UpdatedAt = now

	if err := s.appStore.Document().UpdateDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to issue document: %w", err)
	}

	toFlow := string(domain.DocumentFlowStatusIssued)
	act := &domain.DocumentActivity{
		Id:          ulid.Make().String(),
		DocumentId:  doc.Id,
		Action:      "issued",
		ToStatus:    &toFlow,
		PerformedBy: &userID,
		Note:        "Issued document to client",
		CreatedAt:   now,
	}
	_ = s.appStore.Activity().InsertActivity(ctx, act)

	return toDocumentResponse(doc), nil
}

func (s *Service) TransitionFlowStatus(ctx context.Context, tenantID string, userID string, docID string, newFlow domain.DocumentFlowStatus) (*DocumentResponse, error) {
	now := time.Now()

	doc, err := s.appStore.Document().GetDocumentByIDAndTenant(ctx, docID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	if !doc.CanTransitionFlow(newFlow) {
		return nil, fmt.Errorf("cannot transition flow status to %s", newFlow)
	}

	var fromFlowStr *string
	if doc.FlowStatus != nil {
		sStr := string(*doc.FlowStatus)
		fromFlowStr = &sStr
	}
	toFlowStr := string(newFlow)

	doc.FlowStatus = &newFlow
	doc.UpdatedAt = now

	if err := s.appStore.Document().UpdateDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to update flow status: %w", err)
	}

	act := &domain.DocumentActivity{
		Id:          ulid.Make().String(),
		DocumentId:  doc.Id,
		Action:      "flow_status_changed",
		FromStatus:  fromFlowStr,
		ToStatus:    &toFlowStr,
		PerformedBy: &userID,
		Note:        fmt.Sprintf("Transitioned flow status to %s", newFlow),
		CreatedAt:   now,
	}

	if err := s.appStore.Activity().InsertActivity(ctx, act); err != nil {
		return nil, fmt.Errorf("failed to log activity: %w", err)
	}

	return toDocumentResponse(doc), nil
}

func (s *Service) RecordPayment(ctx context.Context, tenantID string, userID string, docID string, req *RecordPaymentRequest) (*DocumentResponse, error) {
	doc, err := s.appStore.Document().GetDocumentByIDAndTenant(ctx, docID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	if doc.Status != domain.DocumentStatusPublished {
		return nil, fmt.Errorf("payments can only be recorded for published documents")
	}

	now := time.Now()
	payDate := now
	if req.PaymentDate != nil {
		payDate = *req.PaymentDate
	}

	payment := &domain.DocumentPayment{
		Id:            ulid.Make().String(),
		TenantId:      tenantID,
		DocumentId:    docID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		PaymentDate:   payDate,
		ReferenceNo:   req.ReferenceNo,
		Notes:         req.Notes,
		CreatedBy:     userID,
		CreatedAt:     now,
	}

	doc.AmountPaid = doc.AmountPaid.Add(req.Amount)
	doc.UpdatedAt = now

	var newFlow domain.DocumentFlowStatus
	if doc.AmountPaid.GreaterThanOrEqual(doc.Total) {
		newFlow = domain.DocumentFlowStatusPaid
	} else {
		newFlow = domain.DocumentFlowStatusPartiallyPaid
	}
	doc.FlowStatus = &newFlow

	toFlow := string(newFlow)
	act := &domain.DocumentActivity{
		Id:          ulid.Make().String(),
		DocumentId:  docID,
		Action:      "payment_received",
		ToStatus:    &toFlow,
		PerformedBy: &userID,
		Note:        fmt.Sprintf("Recorded payment of %s via %s (Ref: %s)", req.Amount.String(), req.PaymentMethod, req.ReferenceNo),
		CreatedAt:   now,
	}

	if err := s.appStore.Document().RecordPaymentAndUpdateStatus(ctx, payment, doc, act); err != nil {
		return nil, fmt.Errorf("failed to record payment: %w", err)
	}

	return toDocumentResponse(doc), nil
}

func (s *Service) ConvertDocument(ctx context.Context, tenantID string, userID string, sourceDocID string, targetType domain.DocumentType) (*DocumentResponse, error) {
	sourceDoc, err := s.appStore.Document().GetDocumentByIDAndTenant(ctx, sourceDocID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("source document not found: %w", err)
	}

	if sourceDoc.Status != domain.DocumentStatusPublished {
		return nil, fmt.Errorf("only published documents can be converted")
	}

	itemsReq := make([]CreateItemRequest, len(sourceDoc.Items))
	for i, item := range sourceDoc.Items {
		itemsReq[i] = CreateItemRequest{
			SortOrder:   item.SortOrder,
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			DiscountPct: item.DiscountPct,
			TaxPct:      item.TaxPct,
		}
	}

	createReq := &CreateDraftRequest{
		Type:           targetType,
		ContactID:      sourceDoc.ContactId,
		Currency:       sourceDoc.Currency,
		Notes:          sourceDoc.Notes,
		Terms:          sourceDoc.Terms,
		Footer:         sourceDoc.Footer,
		DiscountAmount: sourceDoc.DiscountAmount,
		Items:          itemsReq,
	}

	newDoc, err := s.ReserveSequenceAndDraft(ctx, tenantID, userID, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create converted draft: %w", err)
	}

	// Update newDoc parent_id to sourceDocID
	newDocEntity, err := s.appStore.Document().GetDocumentByIDAndTenant(ctx, newDoc.ID, tenantID)
	if err == nil {
		newDocEntity.ParentId = &sourceDocID
		_ = s.appStore.Document().UpdateDocument(ctx, newDocEntity)
	}

	// Update source doc flow status to accepted if applicable
	acceptedFlow := domain.DocumentFlowStatusAccepted
	sourceDoc.FlowStatus = &acceptedFlow
	_ = s.appStore.Document().UpdateDocument(ctx, sourceDoc)

	toFlow := string(domain.DocumentFlowStatusAccepted)
	act := &domain.DocumentActivity{
		Id:          ulid.Make().String(),
		DocumentId:  sourceDocID,
		Action:      "converted",
		ToStatus:    &toFlow,
		PerformedBy: &userID,
		Note:        fmt.Sprintf("Converted to %s (ID: %s)", targetType, newDoc.ID),
		CreatedAt:   time.Now(),
	}
	_ = s.appStore.Activity().InsertActivity(ctx, act)

	return newDoc, nil
}

func (s *Service) CancelDocument(ctx context.Context, tenantID string, userID string, docID string, reason string) (*DocumentResponse, error) {
	now := time.Now()

	doc, err := s.appStore.Document().GetDocumentByIDAndTenant(ctx, docID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	if doc.Status == domain.DocumentStatusCancelled {
		return nil, fmt.Errorf("document is already cancelled")
	}

	fromStatus := string(doc.Status)
	toStatus := string(domain.DocumentStatusCancelled)

	doc.Status = domain.DocumentStatusCancelled
	doc.UpdatedAt = now

	if err := s.appStore.Document().UpdateDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to cancel document: %w", err)
	}

	act := &domain.DocumentActivity{
		Id:          ulid.Make().String(),
		DocumentId:  doc.Id,
		Action:      "cancelled",
		FromStatus:  &fromStatus,
		ToStatus:    &toStatus,
		PerformedBy: &userID,
		Note:        reason,
		CreatedAt:   now,
	}

	if err := s.appStore.Activity().InsertActivity(ctx, act); err != nil {
		return nil, fmt.Errorf("failed to log activity: %w", err)
	}

	return toDocumentResponse(doc), nil
}

func (s *Service) GetDocument(ctx context.Context, tenantID string, docID string) (*DocumentResponse, error) {
	doc, err := s.appStore.Document().GetDocumentByIDAndTenant(ctx, docID, tenantID)
	if err != nil {
		return nil, err
	}
	acts, err := s.appStore.Activity().ListActivitiesByDocumentID(ctx, docID)
	if err == nil && len(acts) > 0 {
		doc.Activities = acts
	}
	payments, err := s.appStore.Document().ListPaymentsByDocumentID(ctx, tenantID, docID)
	if err == nil && len(payments) > 0 {
		doc.Payments = payments
	}
	return toDocumentResponse(doc), nil
}

func (s *Service) DeleteDocument(ctx context.Context, tenantID string, docID string) error {
	doc, err := s.appStore.Document().GetDocumentByIDAndTenant(ctx, docID, tenantID)
	if err != nil {
		return err
	}

	if !doc.IsDeletable() {
		return domain.ErrDocumentImmutable
	}

	return s.appStore.Document().SoftDeleteDocument(ctx, docID, tenantID)
}

func (s *Service) ListDocuments(ctx context.Context, tenantID string, docType string, status string, page int, limit int) ([]DocumentResponse, int64, error) {
	docs, total, err := s.appStore.Document().ListDocumentsByTenant(ctx, tenantID, docType, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]DocumentResponse, len(docs))
	for i, doc := range docs {
		resp[i] = *toDocumentResponse(&doc)
	}
	return resp, total, nil
}

func validateContactRole(docType domain.DocumentType, contact *domain.Contact) error {
	if contact == nil {
		return nil
	}
	switch docType {
	case domain.DocumentTypeQuotation, domain.DocumentTypeInvoice, domain.DocumentTypeReceipt, domain.DocumentTypeSalesOrder:
		if !contact.IsClient() {
			return fmt.Errorf("document type %s requires contact with 'client' role", docType)
		}
	case domain.DocumentTypePurchaseOrder, domain.DocumentTypeDebitNote:
		if !contact.IsSupplier() {
			return fmt.Errorf("document type %s requires contact with 'supplier' role", docType)
		}
	}
	return nil
}

func toDocumentResponse(d *domain.Document) *DocumentResponse {
	var issueDate, dueDate, validUntil, pdfGenAt *time.Time
	if d.IssueDate.Valid {
		issueDate = &d.IssueDate.Time
	}
	if d.DueDate.Valid {
		dueDate = &d.DueDate.Time
	}
	if d.ValidUntil.Valid {
		validUntil = &d.ValidUntil.Time
	}
	if d.PdfGeneratedAt.Valid {
		pdfGenAt = &d.PdfGeneratedAt.Time
	}

	items := make([]DocumentItemResponse, len(d.Items))
	for i, item := range d.Items {
		items[i] = DocumentItemResponse{
			ID:          item.Id,
			SortOrder:   item.SortOrder,
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			DiscountPct: item.DiscountPct,
			TaxPct:      item.TaxPct,
			Amount:      item.Amount,
		}
	}

	activities := make([]DocumentActivityResponse, len(d.Activities))
	for i, act := range d.Activities {
		activities[i] = DocumentActivityResponse{
			ID:          act.Id,
			Action:      act.Action,
			FromStatus:  act.FromStatus,
			ToStatus:    act.ToStatus,
			PerformedBy: act.PerformedBy,
			Note:        act.Note,
			CreatedAt:   act.CreatedAt,
		}
	}

	payments := make([]DocumentPaymentResponse, len(d.Payments))
	for i, p := range d.Payments {
		payments[i] = DocumentPaymentResponse{
			ID:            p.Id,
			Amount:        p.Amount,
			PaymentMethod: p.PaymentMethod,
			PaymentDate:   p.PaymentDate,
			ReferenceNo:   p.ReferenceNo,
			Notes:         p.Notes,
			CreatedBy:     p.CreatedBy,
			CreatedAt:     p.CreatedAt,
		}
	}

	return &DocumentResponse{
		ID:             d.Id,
		TenantID:       d.TenantId,
		OrganizationID: d.OrganizationId,
		ContactID:      d.ContactId,
		ParentID:       d.ParentId,
		Type:           d.Type,
		Status:         d.Status,
		FlowStatus:     d.FlowStatus,
		DocumentRef:    d.DocumentRef,
		IssueDate:      issueDate,
		DueDate:        dueDate,
		ValidUntil:     validUntil,
		Currency:       d.Currency,
		Subtotal:       d.Subtotal,
		DiscountAmount: d.DiscountAmount,
		TaxAmount:      d.TaxAmount,
		Total:          d.Total,
		AmountPaid:     d.AmountPaid,
		Notes:          d.Notes,
		Terms:          d.Terms,
		Footer:         d.Footer,
		PdfS3Key:       d.PdfS3Key,
		PdfGeneratedAt: pdfGenAt,
		Items:          items,
		Activities:     activities,
		Payments:       payments,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}
