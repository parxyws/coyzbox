package psql

import (
	"context"
	"errors"

	"github.com/parxyws/cozybox/internal/domain"
	"gorm.io/gorm"
)

type TenantRepo struct{ DB *gorm.DB }

func NewTenantRepo(db *gorm.DB) *TenantRepo { return &TenantRepo{DB: db} }

func (r *TenantRepo) Insert(ctx context.Context, t *domain.Tenant) error {
	return r.DB.WithContext(ctx).Create(t).Error
}
func (r *TenantRepo) Update(ctx context.Context, t *domain.Tenant) error {
	return r.DB.WithContext(ctx).Save(t).Error
}

func (r *TenantRepo) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepo) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	if err := r.DB.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &tenant, nil
}
