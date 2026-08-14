package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(pwd string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	return string(h)
}

var correctHash = hashPassword("correctpass")

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, m := newHandlerWithMocks(t)

	reqDto := auth.LoginRequest{Email: "test@example.com", Password: "correctpass"}

	m.userStore.On("GetUserByEmail", mock.Anything, reqDto.Email).Return(&domain.User{
		Id:         "user-123",
		Email:      reqDto.Email,
		Password:   correctHash,
		IsVerified: true,
	}, nil)

	m.tenantStore.On("ListTenantMembersByUserID", mock.Anything, "user-123").Return([]domain.TenantMember{{
		Id:       "member-1",
		TenantId: "tenant-1",
		UserId:   "user-123",
		Role:     domain.TenantRoleOwner,
		Tenant:   domain.Tenant{Id: "tenant-1", Name: "Test Corp", Slug: "test-corp"},
	}}, nil)

	m.sessionStore.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.userStore.On("UpdateUser", mock.Anything, mock.Anything).Return(nil)

	body, _ := json.Marshal(reqDto)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogin_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, m := newHandlerWithMocks(t)

	reqDto := auth.LoginRequest{Email: "notfound@example.com", Password: "correctpass"}

	m.userStore.On("GetUserByEmail", mock.Anything, reqDto.Email).Return(nil, store.ErrNotFound)

	body, _ := json.Marshal(reqDto)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
