package middleware

import (
	"sync"

	"github.com/parxyws/cozybox/internal/config"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ManagerMiddleware struct {
	logger          *logrus.Logger
	cfg             *config.Config
	db              *gorm.DB
	mu              sync.RWMutex
	permissionCache map[string]map[string]bool
}

func NewMiddlewareManager(cfg *ConfigMiddleware) *ManagerMiddleware {
	return &ManagerMiddleware{
		cfg:             cfg.Cfg,
		logger:          cfg.Logger,
		db:              cfg.DB,
		permissionCache: make(map[string]map[string]bool),
	}
}

type ConfigMiddleware struct {
	Logger *logrus.Logger
	Cfg    *config.Config
	DB     *gorm.DB
}
