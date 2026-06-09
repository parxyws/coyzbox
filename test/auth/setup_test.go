package auth_test

import (
	"context"
	"errors"
	"time"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/parxyws/cozybox/internal/app/auth"
	authmocks "github.com/parxyws/cozybox/internal/app/auth/mocks"
	"github.com/parxyws/cozybox/internal/domain"
)

type tb interface {
	mock.TestingT
	Cleanup(func())
}

type stubTokenGen struct {
	AccessToken  string
	RefreshToken string
	Err          error
}

func (s *stubTokenGen) CreateAccessToken(_, _, _ string, _ time.Duration) (string, error) {
	if s.Err != nil {
		return "", s.Err
	}
	return s.AccessToken, nil
}

func (s *stubTokenGen) VerifyAccessToken(_ string) (*domain.Claims, error) {
	return nil, errors.New("not implemented")
}
func (s *stubTokenGen) GenerateRefreshToken() (string, error) {
	if s.Err != nil {
		return "", s.Err
	}
	return s.RefreshToken, nil
}

// stubTransactionManager executes the function directly without a real DB transaction.
// This allows unit tests to verify service logic without a database connection.
type stubTransactionManager struct{}

func (s *stubTransactionManager) WithTransaction(_ context.Context, fn func(tx *gorm.DB) error) error {
	// Pass nil — the repoFactory will ignore the tx anyway and return the mocked repos.
	return fn(nil)
}

// stubRepoFactory returns the pre-configured mock repos regardless of the tx parameter.
// This bridges the new transaction-based Register/CompleteOnboarding with the existing mock setup.
type stubRepoFactory struct {
	userRepo   auth.UserRepository
	tenantRepo auth.TenantRepository
	memberRepo auth.TenantMemberRepository
	orgRepo    auth.OrganizationRepository
}

func (f *stubRepoFactory) UserRepo(_ *gorm.DB) auth.UserRepository           { return f.userRepo }
func (f *stubRepoFactory) TenantRepo(_ *gorm.DB) auth.TenantRepository       { return f.tenantRepo }
func (f *stubRepoFactory) MemberRepo(_ *gorm.DB) auth.TenantMemberRepository { return f.memberRepo }
func (f *stubRepoFactory) OrgRepo(_ *gorm.DB) auth.OrganizationRepository    { return f.orgRepo }

type serviceMocks struct {
	userRepo       *authmocks.MockUserRepository
	tenantRepo     *authmocks.MockTenantRepository
	memberRepo     *authmocks.MockTenantMemberRepository
	orgRepo        *authmocks.MockOrganizationRepository
	sessionStore   *authmocks.MockSessionStore
	tokenGenerator *stubTokenGen
	mailer         *authmocks.MockMailer
	fileStorage    *authmocks.MockFileStorage
}

func newServiceWithMocks(t tb) (*auth.Service, *serviceMocks) {
	m := &serviceMocks{
		userRepo:       authmocks.NewMockUserRepository(t),
		tenantRepo:     authmocks.NewMockTenantRepository(t),
		memberRepo:     authmocks.NewMockTenantMemberRepository(t),
		orgRepo:        authmocks.NewMockOrganizationRepository(t),
		sessionStore:   authmocks.NewMockSessionStore(t),
		tokenGenerator: &stubTokenGen{AccessToken: "access-token", RefreshToken: "refresh-token"},
		mailer:         authmocks.NewMockMailer(t),
		fileStorage:    authmocks.NewMockFileStorage(t),
	}

	txManager := &stubTransactionManager{}
	factory := &stubRepoFactory{
		userRepo:   m.userRepo,
		tenantRepo: m.tenantRepo,
		memberRepo: m.memberRepo,
		orgRepo:    m.orgRepo,
	}

	svc := auth.NewAuthService(
		m.userRepo,
		m.tenantRepo,
		m.memberRepo,
		m.orgRepo,
		m.sessionStore,
		m.tokenGenerator,
		m.mailer,
		m.fileStorage,
		txManager,
		factory,
		"http://localhost:8080",
	)
	return svc, m
}
