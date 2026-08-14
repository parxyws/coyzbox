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
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVerifyEmail_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, m := newHandlerWithMocks(t)

	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"
	userData := auth.UserResponse{
		Id:    "user-123",
		Name:  "Test User",
		Email: "test@example.com",
	}
	data, _ := json.Marshal(userData)

	m.sessionStore.On("Get", mock.Anything, "otp-"+refID).Return("123456", nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)
	m.sessionStore.On("Get", mock.Anything, "ref-"+refID).Return(string(data), nil)
	m.userStore.On("GetUserByID", mock.Anything, "user-123").Return(&domain.User{
		Id:         "user-123",
		Email:      "test@example.com",
		IsVerified: false,
	}, nil)
	m.userStore.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.IsVerified == true
	})).Return(nil)

	reqDto := auth.VerifyEmailRequest{ReferenceId: refID, Otp: "123456"}
	body, _ := json.Marshal(reqDto)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VerifyEmail(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVerifyEmail_OTPNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, m := newHandlerWithMocks(t)

	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"
	m.sessionStore.On("Get", mock.Anything, "otp-"+refID).Return("", errors.New("not found"))

	reqDto := auth.VerifyEmailRequest{ReferenceId: refID, Otp: "123456"}
	body, _ := json.Marshal(reqDto)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VerifyEmail(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyEmail_OTPMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, m := newHandlerWithMocks(t)

	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"
	m.sessionStore.On("Get", mock.Anything, "otp-"+refID).Return("999999", nil)

	reqDto := auth.VerifyEmailRequest{ReferenceId: refID, Otp: "123456"}
	body, _ := json.Marshal(reqDto)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.VerifyEmail(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
