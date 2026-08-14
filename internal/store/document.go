package store

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/parxyws/cozybox/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DocumentStore interface {
	InsertDocument(ctx context.Context, doc *domain.Document) error
	UpdateDocument(ctx context.Context, doc *domain.Document) error
	GetDocumentByIDAndTenant(ctx context.Context, id string, tenantID string) (*domain.Document, error)
	SoftDeleteDocument(ctx context.Context, id string, tenantID string) error
	ListDocumentsByTenant(ctx context.Context, tenantID string, docType string, status string, page int, limit int) ([]domain.Document, int64, error)

	// Payment operations
	InsertPayment(ctx context.Context, payment *domain.DocumentPayment) error
	ListPaymentsByDocumentID(ctx context.Context, tenantID string, documentID string) ([]domain.DocumentPayment, error)
	RecordPaymentAndUpdateStatus(ctx context.Context, payment *domain.DocumentPayment, updatedDoc *domain.Document, activity *domain.DocumentActivity) error
}

type SequenceStore interface {
	GetSequenceSetting(ctx context.Context, tenantID string, docType domain.DocumentType) (*domain.TenantSequenceSetting, error)
	ReserveNextSequence(ctx context.Context, tenantID string, docType domain.DocumentType) (string, error)
}

type ActivityStore interface {
	InsertActivity(ctx context.Context, activity *domain.DocumentActivity) error
	ListActivitiesByDocumentID(ctx context.Context, documentID string) ([]domain.DocumentActivity, error)
}

type psqlDocumentStore struct {
	db *gorm.DB
}

func NewDocumentStore(db *gorm.DB) DocumentStore {
	return &psqlDocumentStore{db: db}
}

func (s *psqlDocumentStore) InsertDocument(ctx context.Context, doc *domain.Document) error {
	return s.db.WithContext(ctx).Create(doc).Error
}

func (s *psqlDocumentStore) UpdateDocument(ctx context.Context, doc *domain.Document) error {
	return s.db.WithContext(ctx).Save(doc).Error
}

func (s *psqlDocumentStore) GetDocumentByIDAndTenant(ctx context.Context, id string, tenantID string) (*domain.Document, error) {
	var doc domain.Document
	err := s.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).
		Preload("Activities", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).
		Preload("Payments", func(db *gorm.DB) *gorm.DB { return db.Order("payment_date DESC") }).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
		First(&doc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &doc, nil
}

func (s *psqlDocumentStore) SoftDeleteDocument(ctx context.Context, id string, tenantID string) error {
	return s.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Update("deleted_at", time.Now()).Error
}

func (s *psqlDocumentStore) ListDocumentsByTenant(ctx context.Context, tenantID string, docType string, status string, page int, limit int) ([]domain.Document, int64, error) {
	var docs []domain.Document
	var total int64

	db := s.db.WithContext(ctx).Model(&domain.Document{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	if docType != "" {
		db = db.Where("type = ?", docType)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := db.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Preload("Items").
		Find(&docs).Error
	if err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

func (s *psqlDocumentStore) InsertPayment(ctx context.Context, payment *domain.DocumentPayment) error {
	return s.db.WithContext(ctx).Create(payment).Error
}

func (s *psqlDocumentStore) ListPaymentsByDocumentID(ctx context.Context, tenantID string, documentID string) ([]domain.DocumentPayment, error) {
	var payments []domain.DocumentPayment
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ? AND deleted_at IS NULL", tenantID, documentID).
		Order("payment_date DESC").
		Find(&payments).Error
	return payments, err
}

func (s *psqlDocumentStore) RecordPaymentAndUpdateStatus(ctx context.Context, payment *domain.DocumentPayment, updatedDoc *domain.Document, activity *domain.DocumentActivity) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if payment != nil {
			if err := tx.Create(payment).Error; err != nil {
				return fmt.Errorf("failed to insert payment record: %w", err)
			}
		}
		if updatedDoc != nil {
			if err := tx.Save(updatedDoc).Error; err != nil {
				return fmt.Errorf("failed to update document status: %w", err)
			}
		}
		if activity != nil {
			if err := tx.Create(activity).Error; err != nil {
				return fmt.Errorf("failed to log activity: %w", err)
			}
		}
		return nil
	})
}

type psqlSequenceStore struct {
	db *gorm.DB
}

func NewSequenceStore(db *gorm.DB) SequenceStore {
	return &psqlSequenceStore{db: db}
}

func (s *psqlSequenceStore) GetSequenceSetting(ctx context.Context, tenantID string, docType domain.DocumentType) (*domain.TenantSequenceSetting, error) {
	var setting domain.TenantSequenceSetting
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND document_type = ?", tenantID, docType).
		First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &domain.TenantSequenceSetting{
				Id:              ulid.Make().String(),
				TenantId:        tenantID,
				DocumentType:    docType,
				PatternTemplate: "{{TYPE}}/{{YYYY}}/{{MM}}/{{SEQ:4}}",
				ResetCycle:      domain.SequenceResetYearly,
			}, nil
		}
		return nil, err
	}
	return &setting, nil
}

func (s *psqlSequenceStore) ReserveNextSequence(ctx context.Context, tenantID string, docType domain.DocumentType) (string, error) {
	setting, err := s.GetSequenceSetting(ctx, tenantID, docType)
	if err != nil {
		return "", err
	}

	now := time.Now()
	var period string
	switch setting.ResetCycle {
	case domain.SequenceResetMonthly:
		period = now.Format("2006-01")
	case domain.SequenceResetYearly:
		period = now.Format("2006")
	default:
		period = "ALL"
	}

	var state domain.TenantSequenceState
	err = s.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tenant_id = ? AND document_type = ? AND current_period = ?", tenantID, docType, period).
		First(&state).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			state = domain.TenantSequenceState{
				TenantId:      tenantID,
				DocumentType:  docType,
				CurrentPeriod: period,
				LastValue:     0,
				UpdatedAt:     now,
			}
		} else {
			return "", err
		}
	}

	state.LastValue++
	state.UpdatedAt = now

	if err := s.db.WithContext(ctx).Save(&state).Error; err != nil {
		return "", err
	}

	return formatSequencePattern(setting.PatternTemplate, docType, now, state.LastValue), nil
}

type psqlActivityStore struct {
	db *gorm.DB
}

func NewActivityStore(db *gorm.DB) ActivityStore {
	return &psqlActivityStore{db: db}
}

func (s *psqlActivityStore) InsertActivity(ctx context.Context, activity *domain.DocumentActivity) error {
	return s.db.WithContext(ctx).Create(activity).Error
}

func (s *psqlActivityStore) ListActivitiesByDocumentID(ctx context.Context, documentID string) ([]domain.DocumentActivity, error) {
	var activities []domain.DocumentActivity
	err := s.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Order("created_at ASC").
		Find(&activities).Error
	return activities, err
}

var seqPatternRegex = regexp.MustCompile(`\{\{SEQ:(\d+)\}\}`)

func formatSequencePattern(pattern string, docType domain.DocumentType, now time.Time, seqVal int) string {
	res := pattern
	res = strings.ReplaceAll(res, "{{TYPE}}", strings.ToUpper(string(docType)))
	res = strings.ReplaceAll(res, "{{YYYY}}", now.Format("2006"))
	res = strings.ReplaceAll(res, "{{MM}}", now.Format("01"))
	res = strings.ReplaceAll(res, "{{DD}}", now.Format("02"))

	res = seqPatternRegex.ReplaceAllStringFunc(res, func(m string) string {
		match := seqPatternRegex.FindStringSubmatch(m)
		if len(match) > 1 {
			width, _ := strconv.Atoi(match[1])
			return fmt.Sprintf("%0*d", width, seqVal)
		}
		return fmt.Sprintf("%04d", seqVal)
	})

	return res
}
