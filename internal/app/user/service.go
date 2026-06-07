package user

import (
	"context"

	"github.com/parxyws/cozybox/internal/domain"
)

type UserRepository interface {
	Insert(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

type TenantRepository interface {
	Insert(ctx context.Context, tenant *domain.Tenant) error
	Update(ctx context.Context, tenant *domain.Tenant) error
	GetByID(ctx context.Context, id string) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
}

type TenantMemberRepository interface {
	Insert(ctx context.Context, member *domain.TenantMember) error
	GetByUserID(ctx context.Context, userID string) (*domain.TenantMember, error)
}

type OrganizationRepository interface {
	Insert(ctx context.Context, org *domain.Organization) error
	Update(ctx context.Context, org *domain.Organization) error
	GetByID(ctx context.Context, id string) (*domain.Organization, error)
}

type Service struct {
	userRepo   UserRepository
	tenantRepo TenantRepository
	memberRepo TenantMemberRepository
	orgRepo    OrganizationRepository
}
