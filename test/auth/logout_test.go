package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, m := newHandlerWithMocks(t)

	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer mock-access-token")
	ctx := context.WithValue(req.Context(), config.SessionID, "session-123")
	c.Request = req.WithContext(ctx)
	c.Set(string(config.SessionID), "session-123")

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
