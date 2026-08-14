package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/middleware"
	"github.com/stretchr/testify/assert"
)

type mockUserGetter struct {
	user *domain.User
	err  error
}

func (m *mockUserGetter) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return m.user, m.err
}

func TestRequireOnboarding_Completed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userGetter := &mockUserGetter{
		user: &domain.User{Id: "user-1", OnboardingCompleted: true},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(config.UserID), "user-1")
		c.Next()
	})
	router.Use(middleware.RequireOnboarding(userGetter))
	router.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestRequireOnboarding_NotCompleted_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userGetter := &mockUserGetter{
		user: &domain.User{Id: "user-1", OnboardingCompleted: false},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(config.UserID), "user-1")
		c.Next()
	})
	router.Use(middleware.RequireOnboarding(userGetter))
	router.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
