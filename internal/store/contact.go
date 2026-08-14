package store

import (
	"context"
	"errors"

	"github.com/parxyws/cozybox/internal/domain"
	"gorm.io/gorm"
)

type ContactStore interface {
	InsertContact(ctx context.Context, contact *domain.Contact) error
	UpdateContact(ctx context.Context, contact *domain.Contact) error
	GetContactByIDAndTenant(ctx context.Context, id string, tenantID string) (*domain.Contact, error)
	SoftDeleteContact(ctx context.Context, id string, tenantID string) error
	ListContacts(ctx context.Context, tenantID string, search string, role string, page int, limit int) ([]domain.Contact, int64, error)
}

type psqlContactStore struct {
	db *gorm.DB
}

func NewContactStore(db *gorm.DB) ContactStore {
	return &psqlContactStore{db: db}
}

func (s *psqlContactStore) InsertContact(ctx context.Context, contact *domain.Contact) error {
	return s.db.WithContext(ctx).Create(contact).Error
}

func (s *psqlContactStore) UpdateContact(ctx context.Context, contact *domain.Contact) error {
	return s.db.WithContext(ctx).Save(contact).Error
}

func (s *psqlContactStore) GetContactByIDAndTenant(ctx context.Context, id string, tenantID string) (*domain.Contact, error) {
	var contact domain.Contact
	err := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&contact).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &contact, nil
}

func (s *psqlContactStore) SoftDeleteContact(ctx context.Context, id string, tenantID string) error {
	res := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&domain.Contact{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *psqlContactStore) ListContacts(ctx context.Context, tenantID string, search string, role string, page int, limit int) ([]domain.Contact, int64, error) {
	var contacts []domain.Contact
	var total int64

	query := s.db.WithContext(ctx).Model(&domain.Contact{}).Where("tenant_id = ?", tenantID)

	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if role != "" {
		query = query.Where("roles LIKE ?", "%"+role+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&contacts).Error; err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}
