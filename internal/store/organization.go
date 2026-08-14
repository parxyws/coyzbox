package store

import (
	"context"
	"errors"

	"github.com/parxyws/cozybox/internal/domain"
	"gorm.io/gorm"
)

type OrganizationStore interface {
	InsertOrganization(ctx context.Context, o *domain.Organization) error
	GetOrganizationByTenantID(ctx context.Context, tenantID string) (*domain.Organization, error)
	UpdateOrganization(ctx context.Context, o *domain.Organization) error
}

type psqlOrgStore struct {
	db *gorm.DB
}

func NewOrganizationStore(db *gorm.DB) OrganizationStore {
	return &psqlOrgStore{db: db}
}

func (s *psqlOrgStore) InsertOrganization(ctx context.Context, o *domain.Organization) error {
	return s.db.WithContext(ctx).Create(o).Error
}

func (s *psqlOrgStore) GetOrganizationByTenantID(ctx context.Context, tenantID string) (*domain.Organization, error) {
	var org domain.Organization
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &org, nil
}

func (s *psqlOrgStore) UpdateOrganization(ctx context.Context, o *domain.Organization) error {
	return s.db.WithContext(ctx).Save(o).Error
}
