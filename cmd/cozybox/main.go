package main

import (
	"fmt"
	"os"

	"github.com/parxyws/cozybox/internal/app/server"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/pkg/database/psql"
	"github.com/parxyws/cozybox/internal/pkg/database/redis"
	"github.com/parxyws/cozybox/internal/pkg/database/storage"
	"github.com/parxyws/cozybox/internal/pkg/logger"
	"github.com/parxyws/cozybox/internal/pkg/mail"
)

func main() {
	cfg, err := config.InitAppConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logrusLogger, logWriter := logger.NewLogger(cfg)
	defer func() {
		if err := logWriter.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close log writer: %v\n", err)
		}
	}()

	db, err := psql.InitPostgres(cfg)
	if err != nil {
		logrusLogger.Fatalf("failed to connect to postgres: %v", err)
	}
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			logrusLogger.Errorf("failed to get underlying sql.DB: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			logrusLogger.Errorf("failed to close database connection: %v", err)
		}
	}()

	authRedis, err := redis.InitAuthRedis(cfg)
	if err != nil {
		logrusLogger.Fatalf("failed to connect to auth redis: %v", err)
	}
	defer func() {
		if err := authRedis.Close(); err != nil {
			logrusLogger.Errorf("failed to close auth redis: %v", err)
		}
	}()

	cacheRedis, err := redis.InitCacheRedis(cfg)
	if err != nil {
		logrusLogger.Fatalf("failed to connect to cache redis: %v", err)
	}
	defer func() {
		if err := cacheRedis.Close(); err != nil {
			logrusLogger.Errorf("failed to close cache redis: %v", err)
		}
	}()

	limiterRedis, err := redis.InitLimiterRedis(cfg)
	if err != nil {
		logrusLogger.Fatalf("failed to connect to limiter redis: %v", err)
	}
	defer func() {
		if err := limiterRedis.Close(); err != nil {
			logrusLogger.Errorf("failed to close limiter redis: %v", err)
		}
	}()

	minioClient, err := storage.InitMinio(cfg)
	if err != nil {
		logrusLogger.Fatalf("failed to connect to minio: %v", err)
	}

	//tokenMaker, err := jwt.NewJWTMaker(cfg.Server.JWTSecretKey)
	//if err != nil {
	//	logrusLogger.Fatalf("failed to create jwt maker: %v", err)
	//}

	gomail := mail.NewGoMailDialer(cfg)

	srv := server.Initialize(&server.ServerConfig{
		Cfg:          cfg,
		Db:           db,
		MinioClient:  minioClient,
		AuthRedis:    authRedis.Client,
		CacheRedis:   cacheRedis.Client,
		LimiterRedis: limiterRedis.Client,
		Logger:       logrusLogger,
		Mail:         gomail,
	})

	if err := srv.Init(); err != nil {
		logrusLogger.Fatalf("failed to start server: %v", err)
	}
}
