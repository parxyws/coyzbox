package auth_test

import (
	"errors"
	"time"

	"github.com/stretchr/testify/mock"

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
	svc := auth.NewAuthService(
		m.userRepo,
		m.tenantRepo,
		m.memberRepo,
		m.orgRepo,
		m.sessionStore,
		m.tokenGenerator,
		m.mailer,
		m.fileStorage,
		"http://localhost:8080",
	)
	return svc, m
}
