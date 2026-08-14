package store

import (
	"context"
	"errors"

	"github.com/parxyws/cozybox/internal/domain"
	"gorm.io/gorm"
)

type TenantStore interface {
	InsertTenant(ctx context.Context, t *domain.Tenant) error
	UpdateTenant(ctx context.Context, t *domain.Tenant) error
	GetTenantByID(ctx context.Context, id string) (*domain.Tenant, error)
	GetTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error)

	// Task-Oriented Direct Transaction Method
	RegisterTenantOwner(ctx context.Context, user *domain.User, tenant *domain.Tenant, member *domain.TenantMember, org *domain.Organization) error

	// Member operations
	InsertTenantMember(ctx context.Context, tm *domain.TenantMember) error
	GetTenantMemberByUserID(ctx context.Context, userID string) (*domain.TenantMember, error)
	ListTenantMembersByUserID(ctx context.Context, userID string) ([]domain.TenantMember, error)

	// Template Config operations
	ListTemplateConfigsByTenantID(ctx context.Context, tenantID string) ([]domain.TemplateConfig, error)
	GetTemplateConfigByID(ctx context.Context, id string) (*domain.TemplateConfig, error)
	UpdateTemplateConfig(ctx context.Context, cfg *domain.TemplateConfig) error
}

type psqlTenantStore struct {
	db *gorm.DB
}

func NewTenantStore(db *gorm.DB) TenantStore {
	return &psqlTenantStore{db: db}
}

func (s *psqlTenantStore) InsertTenant(ctx context.Context, t *domain.Tenant) error {
	return s.db.WithContext(ctx).Create(t).Error
}

func (s *psqlTenantStore) UpdateTenant(ctx context.Context, t *domain.Tenant) error {
	return s.db.WithContext(ctx).Save(t).Error
}

func (s *psqlTenantStore) GetTenantByID(ctx context.Context, id string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

func (s *psqlTenantStore) GetTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	if err := s.db.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

// RegisterTenantOwner executes atomic multi-table persistence inside a DB transaction internally.
func (s *psqlTenantStore) RegisterTenantOwner(ctx context.Context, user *domain.User, tenant *domain.Tenant, member *domain.TenantMember, org *domain.Organization) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if user != nil {
			if err := tx.Create(user).Error; err != nil {
				return err
			}
		}
		if tenant != nil {
			if err := tx.Create(tenant).Error; err != nil {
				return err
			}
		}
		if member != nil {
			if err := tx.Create(member).Error; err != nil {
				return err
			}
		}
		if org != nil {
			if err := tx.Create(org).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *psqlTenantStore) InsertTenantMember(ctx context.Context, tm *domain.TenantMember) error {
	return s.db.WithContext(ctx).Create(tm).Error
}

func (s *psqlTenantStore) GetTenantMemberByUserID(ctx context.Context, userID string) (*domain.TenantMember, error) {
	var member domain.TenantMember
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (s *psqlTenantStore) ListTenantMembersByUserID(ctx context.Context, userID string) ([]domain.TenantMember, error) {
	var members []domain.TenantMember
	err := s.db.WithContext(ctx).
		Preload("Tenant").
		Where("user_id = ?", userID).
		Order("CASE WHEN role = 'owner' THEN 0 ELSE 1 END, joined_at ASC").
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (s *psqlTenantStore) ListTemplateConfigsByTenantID(ctx context.Context, tenantID string) ([]domain.TemplateConfig, error) {
	var configs []domain.TemplateConfig
	err := s.db.WithContext(ctx).Where("tenant_id = ? AND deleted_at IS NULL", tenantID).Find(&configs).Error
	return configs, err
}

func (s *psqlTenantStore) GetTemplateConfigByID(ctx context.Context, id string) (*domain.TemplateConfig, error) {
	var cfg domain.TemplateConfig
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&cfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &cfg, nil
}

func (s *psqlTenantStore) UpdateTemplateConfig(ctx context.Context, cfg *domain.TemplateConfig) error {
	return s.db.WithContext(ctx).Save(cfg).Error
}
