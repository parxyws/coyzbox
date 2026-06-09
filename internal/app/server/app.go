package server

import (
	"time"

	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/parxyws/cozybox/internal/middleware"
	"github.com/parxyws/cozybox/internal/pkg/database/psql"
	"github.com/parxyws/cozybox/internal/pkg/database/redis"
	"github.com/parxyws/cozybox/internal/pkg/database/storage"
	"github.com/parxyws/cozybox/internal/pkg/jwtutil"
	"github.com/parxyws/cozybox/internal/pkg/mail"
	"gorm.io/gorm"
)

// repoFactory implements auth.RepoFactory using the standard psql repositories scoped to a transaction.
// It lives in the server package to avoid circular dependencies between auth and psql.
type repoFactory struct{}

func (f *repoFactory) UserRepo(tx *gorm.DB) auth.UserRepository     { return psql.NewUserRepo(tx) }
func (f *repoFactory) TenantRepo(tx *gorm.DB) auth.TenantRepository { return psql.NewTenantRepo(tx) }
func (f *repoFactory) MemberRepo(tx *gorm.DB) auth.TenantMemberRepository {
	return psql.NewTenantMemberRepo(tx)
}
func (f *repoFactory) OrgRepo(tx *gorm.DB) auth.OrganizationRepository { return psql.NewOrgRepo(tx) }

func (s *Server) Boostrap() error {
	mailer := mail.NewMailer(s.mail, s.cfg)
	s3Service := storage.NewS3Service(s.minioClient, s.cfg)

	// Repositories
	userRepo := psql.NewUserRepo(s.db)
	tenantRepo := psql.NewTenantRepo(s.db)
	memberRepo := psql.NewTenantMemberRepo(s.db)
	orgRepo := psql.NewOrgRepo(s.db)
	sessionStore := redis.NewSessionStore(s.authRedis)

	// Transaction support
	txManager := psql.NewTransactionManager(s.db)
	factory := &repoFactory{}

	tokenMaker, err := jwtutil.NewJWTMaker(s.cfg.Server.JWTSecretKey)
	if err != nil {
		return err
	}

	addr := s.cfg.Server.BaseUrl

	authService := auth.NewAuthService(
		userRepo,
		tenantRepo,
		memberRepo,
		orgRepo,
		sessionStore,
		tokenMaker,
		mailer,
		s3Service,
		txManager,
		factory,
		addr,
	)
	s.authService = authService

	authHandler := auth.NewAuthHandler(authService)

	// Global middleware
	s.app.Use(middleware.RequestID())
	s.app.Use(middleware.CORS())

	api := s.app.Group("/api")

	// Rate-limited public auth routes
	loginLimiter := middleware.NewRateLimiter(s.limiterRedis, middleware.RateLimitConfig{
		MaxRequests: 5,
		Window:      60 * time.Second,
		KeyFunc:     middleware.RateLimitByIP("login"),
	})
	registerLimiter := middleware.NewRateLimiter(s.limiterRedis, middleware.RateLimitConfig{
		MaxRequests: 3,
		Window:      60 * time.Second,
		KeyFunc:     middleware.RateLimitByIP("register"),
	})
	forgotPwLimiter := middleware.NewRateLimiter(s.limiterRedis, middleware.RateLimitConfig{
		MaxRequests: 3,
		Window:      60 * time.Second,
		KeyFunc:     middleware.RateLimitByIP("forgot-password"),
	})

	authHandler.RegisterRoutes(api, loginLimiter, registerLimiter, forgotPwLimiter)

	// Protected routes with session-validated auth middleware
	protected := api.Group("")
	protected.Use(middleware.NewAuthMiddleware(tokenMaker, sessionStore))
	authHandler.RegisterProtectedRoutes(protected)

	return nil
}
