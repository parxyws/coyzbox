package auth

import (
	"context"
	"time"

	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/pkg/mail"
	"gorm.io/gorm"
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
	ListByUserID(ctx context.Context, userID string) ([]domain.TenantMember, error)
}

type OrganizationRepository interface {
	Insert(ctx context.Context, org *domain.Organization) error
}

type SessionStore interface {
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	SetMultiple(ctx context.Context, pairs map[string]string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keys ...string) error
}

type Mailer interface {
	SendOTP(to string, data mail.OTPData) error
	SendResetPassword(to string, data mail.ResetPasswordData) error
}

type FileStorage interface {
	PutObject(ctx context.Context, input domain.UploadInput) (key string, err error)
}

// TransactionManager provides the ability to execute operations within a database transaction.
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

// RepoFactory creates new repository instances scoped to a given transaction.
// This allows the service to swap repos to transactional versions without
// embedding WithTx in the repository interfaces.
type RepoFactory interface {
	UserRepo(tx *gorm.DB) UserRepository
	TenantRepo(tx *gorm.DB) TenantRepository
	MemberRepo(tx *gorm.DB) TenantMemberRepository
	OrgRepo(tx *gorm.DB) OrganizationRepository
}
