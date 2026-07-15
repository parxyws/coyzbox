package tenant

import (
	"context"

	"github.com/parxyws/cozybox/internal/domain"
)

type OrganizationRepository interface {
	GetByTenantID(ctx context.Context, tenantID string) (*domain.Organization, error)
	Update(ctx context.Context, org *domain.Organization) error
}

type TemplateConfigRepository interface {
	ListByTenantID(ctx context.Context, tenantID string) ([]domain.TemplateConfig, error)
	GetByID(ctx context.Context, id string) (*domain.TemplateConfig, error)
	Update(ctx context.Context, cfg *domain.TemplateConfig) error
}

type FileStorage interface {
	PutObject(ctx context.Context, input domain.UploadInput) (key string, err error)
}
