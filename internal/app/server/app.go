package server

import (
	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/parxyws/cozybox/internal/middleware"
	psql2 "github.com/parxyws/cozybox/internal/pkg/database/psql"
	"github.com/parxyws/cozybox/internal/pkg/database/redis"
	"github.com/parxyws/cozybox/internal/pkg/database/storage"
	"github.com/parxyws/cozybox/internal/pkg/jwtutil"
	"github.com/parxyws/cozybox/internal/pkg/mail"
)

func (s *Server) Boostrap() error {
	mailer := mail.NewMailer(s.mail, s.cfg)
	s3Service := storage.NewS3Service(s.minioClient, s.cfg)

	userRepo := psql2.NewUserRepo(s.db)
	tenantRepo := psql2.NewTenantRepo(s.db)
	memberRepo := psql2.NewTenantMemberRepo(s.db)
	orgRepo := psql2.NewOrgRepo(s.db)
	sessionStore := redis.NewSessionStore(s.authRedis)

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
		addr,
	)

	authHandler := auth.NewAuthHandler(authService)

	s.app.Use(middleware.CORS())

	api := s.app.Group("/api")
	authHandler.RegisterRoutes(api)

	protected := api.Group("")
	protected.Use(middleware.NewAuthMiddleware(tokenMaker))
	authHandler.RegisterProtectedRoutes(protected)

	return nil
}
