package auth_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, m := newHandlerWithMocks(t)

	reqDto := auth.RegisterUserRequest{
		Name:       "Test User",
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "securepass123",
		TenantName: "Test Corp",
	}

	m.tenantStore.On("RegisterTenantOwner", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.sessionStore.On("SetMultiple", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.mailer.On("SendOTP", reqDto.Email, mock.Anything).Return(nil).Maybe()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(reqDto)
	c.Request = httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Register(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	m.tenantStore.AssertExpectations(t)
	m.sessionStore.AssertExpectations(t)
}

func TestRegister_UserInsertFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, m := newHandlerWithMocks(t)

	reqDto := auth.RegisterUserRequest{
		Name:       "Test User",
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "securepass123",
		TenantName: "Test Corp",
	}

	m.tenantStore.On("RegisterTenantOwner", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(reqDto)
	c.Request = httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Register(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	m.tenantStore.AssertExpectations(t)
}
