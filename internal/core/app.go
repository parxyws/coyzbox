package core

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
	"github.com/parxyws/cozybox/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

const (
	ctxTimeout = 5
	certFile   = "./certs/server.crt"
	keyFile    = "./certs/server.key"
)

type AppConfig struct {
	Cfg          *config.Config
	Db           *gorm.DB
	Minio        *minio.Client
	AuthRedis    *redis.Client
	CacheRedis   *redis.Client
	LimiterRedis *redis.Client
	Logger       *logrus.Logger
	Mail         *gomail.Dialer
}

type App struct {
	app          *gin.Engine
	cfg          *config.Config
	db           *gorm.DB
	minio        *minio.Client
	authRedis    *redis.Client
	cacheRedis   *redis.Client
	limiterRedis *redis.Client
	logger       *logrus.Logger
	mail         *gomail.Dialer
}

func NewApplication(opt *AppConfig) *App {
	return &App{
		cfg:          opt.Cfg,
		db:           opt.Db,
		minio:        opt.Minio,
		authRedis:    opt.AuthRedis,
		cacheRedis:   opt.CacheRedis,
		limiterRedis: opt.LimiterRedis,
		logger:       opt.Logger,
		mail:         opt.Mail,
	}
}

func (a *App) Initialize() error {
	switch a.cfg.Server.Mode {
	case gin.ReleaseMode:
		gin.SetMode(gin.ReleaseMode)
	case gin.DebugMode:
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.TestMode)
	}

	a.app = gin.New()

	if err := a.Bootstrap(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      a.app,
		ReadTimeout:  time.Duration(a.cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(a.cfg.Server.WriteTimeout) * time.Second,
	}

	if a.cfg.Server.SSL {
		serverError := make(chan error)

		go func() {
			serverError <- srv.ListenAndServeTLS(certFile, keyFile)
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		select {
		case err := <-serverError:
			log.Fatalf("server error: %s", err)
		case <-quit:
			ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				log.Fatalf("server shutdown error: %s", err)
			}
			log.Println("server shutdown gracefully")
			return nil
		}

	} else {
		serverError := make(chan error)

		go func() {
			serverError <- srv.ListenAndServe()
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		select {
		case err := <-serverError:
			log.Fatalf("Failed to start TLS server: %v", err)
		case <-quit:
			ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				log.Fatalf("Error gracefully shutting down server: %v", err)
			}
			log.Println("Server exited properly")
			return nil
		}
	}

	return nil
}
