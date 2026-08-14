package server

import (
	"time"

	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/parxyws/cozybox/internal/app/contact"
	"github.com/parxyws/cozybox/internal/app/document"
	"github.com/parxyws/cozybox/internal/app/tenant"
	"github.com/parxyws/cozybox/internal/app/user"
	"github.com/parxyws/cozybox/internal/middleware"
	"github.com/parxyws/cozybox/internal/pkg/database/redis"
	"github.com/parxyws/cozybox/internal/pkg/database/storage"
	"github.com/parxyws/cozybox/internal/pkg/jwt"
	"github.com/parxyws/cozybox/internal/pkg/mail"
	"github.com/parxyws/cozybox/internal/store"
)

func (s *Server) Boostrap() error {
	mailer := mail.NewMailer(s.mail, s.cfg)
	s3Service := storage.NewS3Service(s.minioClient, s.cfg)

	// Centralized Store & Redis Session Store
	appStore := store.NewStore(s.db)
	sessionStore := redis.NewSessionStore(s.authRedis)

	tokenMaker, err := jwt.NewJWTMaker(s.cfg.Server.JWTSecretKey)
	if err != nil {
		return err
	}

	addr := s.cfg.Server.BaseUrl

	// Domain Services
	authService := auth.NewAuthService(appStore, sessionStore, tokenMaker, mailer, s3Service, addr)
	userService := user.NewUserService(appStore.User())
	contactService := contact.NewContactService(appStore.Contact(), appStore.Organization())
	tenantService := tenant.NewTenantService(appStore.Tenant(), appStore.Organization(), s3Service)
	docEngine := document.NewService(appStore, s3Service)

	// Handlers
	authHandler := auth.NewAuthHandler(authService)
	s.authHandler = authHandler
	userHandler := user.NewUserHandler(userService)
	contactHandler := contact.NewContactHandler(contactService)
	tenantHandler := tenant.NewHandler(tenantService)
	docHandler := document.NewHandler(docEngine)

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
	contactHandler.RegisterRoutes(protected)
	tenantHandler.RegisterRoutes(protected)

	onboarded := protected.Group("")
	onboarded.Use(middleware.RequireOnboarding(appStore.User()))
	docHandler.RegisterRoutes(onboarded)

	return nil
}
