package psql

import (
	"context"
	"errors"

	"github.com/parxyws/cozybox/internal/domain"
	"gorm.io/gorm"
)

type TenantMemberRepo struct{ DB *gorm.DB }

func NewTenantMemberRepo(db *gorm.DB) *TenantMemberRepo { return &TenantMemberRepo{DB: db} }

func (r *TenantMemberRepo) Insert(ctx context.Context, tm *domain.TenantMember) error {
	return r.DB.WithContext(ctx).Create(tm).Error
}

func (r *TenantMemberRepo) GetByUserID(ctx context.Context, userID string) (*domain.TenantMember, error) {
	var member domain.TenantMember
	if err := r.DB.WithContext(ctx).Where("user_id = ?", userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *TenantMemberRepo) ListByUserID(ctx context.Context, userID string) ([]domain.TenantMember, error) {
	var members []domain.TenantMember
	err := r.DB.WithContext(ctx).
		Preload("Tenant").
		Where("user_id = ?", userID).
		Order("CASE WHEN role = 'owner' THEN 0 ELSE 1 END, joined_at ASC").
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}
