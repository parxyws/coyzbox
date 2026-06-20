package server

import (
	"time"

	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/parxyws/cozybox/internal/app/user"
	"github.com/parxyws/cozybox/internal/middleware"
	"github.com/parxyws/cozybox/internal/pkg/database/psql"
	"github.com/parxyws/cozybox/internal/pkg/database/redis"
	"github.com/parxyws/cozybox/internal/pkg/database/storage"
	"github.com/parxyws/cozybox/internal/pkg/jwt"
	"github.com/parxyws/cozybox/internal/pkg/mail"
)

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

	tokenMaker, err := jwt.NewJWTMaker(s.cfg.Server.JWTSecretKey)
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

	userService := user.NewUserService(userRepo)

	authHandler := auth.NewAuthHandler(authService)
	userHandler := user.NewUserHandler(userService)

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
	userHandler.RegisterRoutes(protected)

	return nil
}
