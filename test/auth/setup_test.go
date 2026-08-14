package auth_test

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/parxyws/cozybox/internal/app/auth"
	authmocks "github.com/parxyws/cozybox/internal/app/auth/mocks"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/store"
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

func (s *stubTokenGen) CreateAccessToken(userID, tenantID, sessionID string, duration time.Duration) (string, error) {
	if s.Err != nil {
		return "", s.Err
	}
	return s.AccessToken, nil
}

func (s *stubTokenGen) VerifyAccessToken(token string) (*domain.Claims, error) {
	return &domain.Claims{SessionID: "sess-1", UserID: "user-1", TenantID: "tenant-1"}, nil
}

func (s *stubTokenGen) GenerateRefreshToken() (string, error) {
	if s.Err != nil {
		return "", s.Err
	}
	return s.RefreshToken, nil
}

type mockUserStore struct {
	mock.Mock
}

func (m *mockUserStore) InsertUser(ctx context.Context, u *domain.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserStore) UpdateUser(ctx context.Context, u *domain.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserStore) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUserStore) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return m.GetUserByID(ctx, id)
}
func (m *mockUserStore) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockUserStore) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

type mockTenantStore struct {
	mock.Mock
}

func (m *mockTenantStore) InsertTenant(ctx context.Context, t *domain.Tenant) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockTenantStore) UpdateTenant(ctx context.Context, t *domain.Tenant) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockTenantStore) GetTenantByID(ctx context.Context, id string) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Tenant), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockTenantStore) GetTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Tenant), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockTenantStore) RegisterTenantOwner(ctx context.Context, user *domain.User, tenant *domain.Tenant, member *domain.TenantMember, org *domain.Organization) error {
	return m.Called(ctx, user, tenant, member, org).Error(0)
}
func (m *mockTenantStore) InsertTenantMember(ctx context.Context, tm *domain.TenantMember) error {
	return m.Called(ctx, tm).Error(0)
}
func (m *mockTenantStore) GetTenantMemberByUserID(ctx context.Context, userID string) (*domain.TenantMember, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.TenantMember), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockTenantStore) ListTenantMembersByUserID(ctx context.Context, userID string) ([]domain.TenantMember, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.TenantMember), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockTenantStore) ListTemplateConfigsByTenantID(ctx context.Context, tenantID string) ([]domain.TemplateConfig, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.TemplateConfig), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockTenantStore) GetTemplateConfigByID(ctx context.Context, id string) (*domain.TemplateConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.TemplateConfig), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockTenantStore) UpdateTemplateConfig(ctx context.Context, cfg *domain.TemplateConfig) error {
	return m.Called(ctx, cfg).Error(0)
}

type mockOrgStore struct {
	mock.Mock
}

func (m *mockOrgStore) InsertOrganization(ctx context.Context, o *domain.Organization) error {
	return m.Called(ctx, o).Error(0)
}
func (m *mockOrgStore) GetOrganizationByTenantID(ctx context.Context, tenantID string) (*domain.Organization, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Organization), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockOrgStore) UpdateOrganization(ctx context.Context, o *domain.Organization) error {
	return m.Called(ctx, o).Error(0)
}

type mockAppStore struct {
	user   *mockUserStore
	tenant *mockTenantStore
	org    *mockOrgStore
}

func (s *mockAppStore) User() store.UserStore                 { return s.user }
func (s *mockAppStore) Tenant() store.TenantStore             { return s.tenant }
func (s *mockAppStore) Organization() store.OrganizationStore { return s.org }
func (s *mockAppStore) Contact() store.ContactStore           { return nil }
func (s *mockAppStore) Document() store.DocumentStore         { return nil }
func (s *mockAppStore) Sequence() store.SequenceStore         { return nil }
func (s *mockAppStore) Activity() store.ActivityStore         { return nil }

type serviceMocks struct {
	userStore      *mockUserStore
	tenantStore    *mockTenantStore
	orgStore       *mockOrgStore
	sessionStore   *authmocks.MockSessionStore
	tokenGenerator *stubTokenGen
	mailer         *authmocks.MockMailer
	fileStorage    *authmocks.MockFileStorage
}

func newHandlerWithMocks(t tb) (*auth.AuthHandler, *serviceMocks) {
	m := &serviceMocks{
		userStore:      &mockUserStore{},
		tenantStore:    &mockTenantStore{},
		orgStore:       &mockOrgStore{},
		sessionStore:   authmocks.NewMockSessionStore(t),
		tokenGenerator: &stubTokenGen{AccessToken: "access-token", RefreshToken: "refresh-token"},
		mailer:         authmocks.NewMockMailer(t),
		fileStorage:    authmocks.NewMockFileStorage(t),
	}

	appStore := &mockAppStore{
		user:   m.userStore,
		tenant: m.tenantStore,
		org:    m.orgStore,
	}

	svc := auth.NewAuthService(
		appStore,
		m.sessionStore,
		m.tokenGenerator,
		m.mailer,
		m.fileStorage,
		"http://localhost:8080",
	)

	handler := auth.NewAuthHandler(svc)
	return handler, m
}
