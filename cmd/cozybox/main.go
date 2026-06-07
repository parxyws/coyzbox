package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/middleware"
	psql2 "github.com/parxyws/cozybox/internal/pkg/database/psql"
	"github.com/parxyws/cozybox/internal/pkg/database/redis"
	"github.com/parxyws/cozybox/internal/pkg/helper"
	"github.com/parxyws/cozybox/internal/pkg/logger"
	"github.com/parxyws/cozybox/internal/pkg/mail"
	"github.com/parxyws/cozybox/internal/pkg/storage"
)

func main() {
	cfg, err := config.InitAppConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logrusLogger := logger.NewLogger(cfg)

	db, err := psql2.InitPostgres(cfg)
	if err != nil {
		logrusLogger.Fatalf("failed to connect to postgres: %v", err)
	}

	authRedis, err := redis.InitAuthRedis(cfg)
	if err != nil {
		logrusLogger.Fatalf("failed to connect to auth redis: %v", err)
	}

	cacheRedis, err := redis.InitCacheRedis(cfg)
	if err != nil {
		logrusLogger.Fatalf("failed to connect to cache redis: %v", err)
	}

	limiterRedis, err := redis.InitLimiterRedis(cfg)
	if err != nil {
		logrusLogger.Fatalf("failed to connect to limiter redis: %v", err)
	}

	minioClient, err := storage.InitMinio(cfg)
	if err != nil {
		logrusLogger.Fatalf("failed to connect to minio: %v", err)
	}

	tokenMaker, err := helper.NewJWTMaker(cfg.Server.JWTSecretKey)
	if err != nil {
		logrusLogger.Fatalf("failed to create jwt maker: %v", err)
	}

	mailDialer := mail.NewGoMailDialer(cfg)
	mailer := mail.NewMailer(mailDialer, cfg)
	s3Service := storage.NewS3Service(minioClient, cfg)

	userRepo := psql2.NewUserRepo(db)
	tenantRepo := psql2.NewTenantRepo(db)
	memberRepo := psql2.NewTenantMemberRepo(db)
	orgRepo := psql2.NewOrgRepo(db)
	sessionStore := redis.NewSessionStore(authRedis.Client)

	addr := cfg.Server.BaseUrl

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

	_ = cacheRedis
	_ = limiterRedis

	r := gin.New()
	r.Use(middleware.CORS())

	api := r.Group("/api")
	authHandler.RegisterRoutes(api)

	protected := api.Group("")
	protected.Use(auth.NewAuthMiddleware(tokenMaker))
	authHandler.RegisterProtectedRoutes(protected)

	logrusLogger.WithFields(map[string]interface{}{
		"host": cfg.Server.Host,
		"port": cfg.Server.Port,
	}).Info("server started")

	addrStr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	if err := r.Run(addrStr); err != nil {
		logrusLogger.Fatalf("failed to start server: %v", err)
	}
}
