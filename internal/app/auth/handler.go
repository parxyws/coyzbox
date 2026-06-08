package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/pkg/helper"
	"github.com/parxyws/cozybox/internal/pkg/validator"

	"github.com/parxyws/cozybox/internal/domain"
)

// AuthService is the consuming-side interface for auth operations.
type AuthService interface {
	Register(ctx context.Context, req *RegisterUserRequest) (*RegisterResponse, error)
	VerifyEmail(ctx context.Context, req *VerifyEmailRequest) (*VerifyEmailResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*UserAuthenticateResponse, error)
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*UserAuthenticateResponse, error)
	Logout(ctx context.Context, sessionID string) error
	CompleteOnboarding(ctx context.Context, userID string, tenantID string, req *OnboardingRequest, image *domain.UploadInput) error
	ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
}

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (a *AuthHandler) RegisterRoutes(route *gin.RouterGroup) {
	authGroup := route.Group("/auth")
	{
		authGroup.POST("/register", a.Register)
		authGroup.POST("/verify-email", a.VerifyEmail)
		authGroup.POST("/login", a.Login)
		authGroup.POST("/refresh-token", a.RefreshToken)
		authGroup.POST("/forgot-password", a.ForgotPassword)
		authGroup.POST("/reset-password", a.ResetPassword)
	}
}

func (a *AuthHandler) RegisterProtectedRoutes(route *gin.RouterGroup) {
	authGroup := route.Group("/auth")
	{
		authGroup.POST("/logout", a.Logout)
	}
}

func (a *AuthHandler) Register(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req RegisterUserRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := a.service.Register(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to register user", err)
		return
	}

	helper.Success(c, http.StatusCreated, "User registered successfully", result)
}

func (a *AuthHandler) VerifyEmail(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req VerifyEmailRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := a.service.VerifyEmail(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to verify email", err)
		return
	}

	helper.Success(c, http.StatusOK, "User email verified successfully", result)
}

func (a *AuthHandler) Login(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req LoginRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := a.service.Login(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusUnauthorized, "Login failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Login successful", result)
}

func (a *AuthHandler) RefreshToken(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req RefreshTokenRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := a.service.RefreshToken(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusUnauthorized, "Token refresh failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Token refreshed successfully", result)
}

func (a *AuthHandler) Logout(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	sessionID, exists := c.Get("session_id")
	if !exists {
		helper.Error(c, http.StatusBadRequest, "Session not found", nil)
		return
	}

	if err := a.service.Logout(ctx, sessionID.(string)); err != nil {
		helper.Error(c, http.StatusInternalServerError, "Logout failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Logged out successfully", nil)
}

func (a *AuthHandler) ForgotPassword(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req ForgotPasswordRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := a.service.ForgotPassword(ctx, &req); err != nil {
		helper.Error(c, http.StatusUnauthorized, "Forgot password failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Request forgot password successfully", nil)
}

func (a *AuthHandler) ResetPassword(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req ResetPasswordRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := a.service.ResetPassword(ctx, &req); err != nil {
		helper.Error(c, http.StatusUnauthorized, "Reset password failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Request reset password successfully", nil)
}
