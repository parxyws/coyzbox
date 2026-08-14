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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestForgotPassword_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, m := newHandlerWithMocks(t)

	reqDto := auth.ForgotPasswordRequest{Email: "test@example.com"}

	m.userStore.On("GetUserByEmail", mock.Anything, reqDto.Email).Return(&domain.User{
		Id:    "user-1",
		Email: reqDto.Email,
		Name:  "Test User",
	}, nil)

	m.sessionStore.On("SetMultiple", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.mailer.On("SendResetPassword", reqDto.Email, mock.Anything).Return(nil).Maybe()

	body, _ := json.Marshal(reqDto)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/auth/forgot-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ForgotPassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
