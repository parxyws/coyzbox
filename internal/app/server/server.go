package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

type ServerConfig struct {
	Cfg          *config.Config
	Db           *gorm.DB
	MinioClient  *minio.Client
	AuthRedis    *redis.Client
	CacheRedis   *redis.Client
	LimiterRedis *redis.Client
	Logger       *logrus.Logger
	Mail         *gomail.Dialer
}

type Server struct {
	cfg          *config.Config
	app          *gin.Engine
	db           *gorm.DB
	minioClient  *minio.Client
	authRedis    *redis.Client
	cacheRedis   *redis.Client
	limiterRedis *redis.Client
	logger       *logrus.Logger
	mail         *gomail.Dialer
	authService  *auth.Service
}

func Initialize(config *ServerConfig) *Server {
	return &Server{
		cfg:          config.Cfg,
		db:           config.Db,
		minioClient:  config.MinioClient,
		authRedis:    config.AuthRedis,
		cacheRedis:   config.CacheRedis,
		limiterRedis: config.LimiterRedis,
		logger:       config.Logger,
		mail:         config.Mail,
	}
}

func (s *Server) Init() error {

	switch s.cfg.Server.Mode {
	case gin.ReleaseMode:
		gin.SetMode(gin.ReleaseMode)
	case gin.DebugMode:
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.TestMode)
	}

	s.app = gin.New()

	if err := s.Boostrap(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      s.app,
		ReadTimeout:  time.Duration(s.cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.cfg.Server.WriteTimeout) * time.Second,
	}

	serverError := make(chan error, 1)

	if s.cfg.Server.SSL {
		go func() {
			serverError <- srv.ListenAndServeTLS(s.cfg.Server.CertFilePath, s.cfg.Server.KeyFilePath)
		}()
	} else {
		go func() {
			serverError <- srv.ListenAndServe()
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverError:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		log.Printf("received signal %v, shutting down server...", sig)

		ctx, cancel := context.WithTimeout(context.Background(), config.CtxTimeout*time.Second)
		defer cancel()

		// Wait for async operations (e.g., email goroutines) to complete
		if s.authService != nil {
			if err := s.authService.Shutdown(); err != nil {
				log.Printf("auth service shutdown error: %v", err)
			}
		}

		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
		log.Println("server shutdown gracefully")
		return nil
	}
}
