package auth

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/domain"
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

func NewAuthMiddleware(tokenGen domain.TokenGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		claims, err := tokenGen.VerifyAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(string(config.UserID), claims.UserID)
		c.Set(string(config.TenantID), claims.TenantID)
		c.Set(string(config.SessionID), claims.SessionID)

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, config.UserID, claims.UserID)
		ctx = context.WithValue(ctx, config.TenantID, claims.TenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
