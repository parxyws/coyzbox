package auth

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/pkg/helper"
	"github.com/parxyws/cozybox/internal/pkg/validator"
	"github.com/parxyws/cozybox/internal/store"
)

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (a *AuthHandler) Shutdown() error {
	return a.authService.Shutdown()
}

func (a *AuthHandler) RegisterRoutes(route *gin.RouterGroup, loginLimiter, registerLimiter, forgotPwLimiter gin.HandlerFunc) {
	authGroup := route.Group("/auth")
	{
		authGroup.POST("/register", registerLimiter, a.Register)
		authGroup.POST("/verify-email", a.VerifyEmail)
		authGroup.POST("/login", loginLimiter, a.Login)
		authGroup.POST("/refresh-token", a.RefreshToken)
		authGroup.POST("/forgot-password", forgotPwLimiter, a.ForgotPassword)
		authGroup.POST("/reset-password", a.ResetPassword)
	}
}

func (a *AuthHandler) RegisterProtectedRoutes(route *gin.RouterGroup) {
	authGroup := route.Group("/auth")
	{
		authGroup.POST("/logout", a.Logout)
	}
	route.GET("/workspaces", a.ListWorkspaces)
	route.POST("/workspaces/switch", a.SwitchWorkspace)
	route.POST("/onboarding", a.CompleteOnboarding)
}

func (a *AuthHandler) Register(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	resp, err := a.authService.Register(ctx, req)
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			helper.Error(c, http.StatusConflict, "User or tenant already exists", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to complete registration", err)
		return
	}

	helper.Success(c, http.StatusCreated, "User registered successfully. Please verify your email.", resp)
}

func (a *AuthHandler) VerifyEmail(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	resp, err := a.authService.VerifyEmail(ctx, req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidOTP) {
			helper.Error(c, http.StatusBadRequest, "Invalid or expired OTP", err)
			return
		}
		if errors.Is(err, domain.ErrSessionExpired) {
			helper.Error(c, http.StatusNotFound, "Registration session expired", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to verify email", err)
		return
	}

	helper.Success(c, http.StatusOK, "Email verified successfully", resp)
}

func (a *AuthHandler) Login(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	req.ClientIP = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	resp, err := a.authService.Login(ctx, req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredential) {
			helper.Error(c, http.StatusUnauthorized, "Invalid credentials", err)
			return
		}
		if errors.Is(err, domain.ErrEmailNotVerified) {
			helper.Error(c, http.StatusForbidden, "Email not verified", err)
			return
		}
		if errors.Is(err, store.ErrNotFound) || errors.Is(err, domain.ErrNotFound) {
			helper.Error(c, http.StatusNotFound, "No workspaces found for user", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to authenticate", err)
		return
	}

	helper.Success(c, http.StatusOK, "Login successful", resp)
}

func (a *AuthHandler) RefreshToken(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	resp, err := a.authService.RefreshToken(ctx, req)
	if err != nil {
		if errors.Is(err, domain.ErrSessionExpired) || errors.Is(err, domain.ErrInvalidToken) {
			helper.Error(c, http.StatusUnauthorized, "Session expired or invalid", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to refresh token", err)
		return
	}

	helper.Success(c, http.StatusOK, "Token refreshed successfully", resp)
}

func (a *AuthHandler) ForgotPassword(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	resp, err := a.authService.ForgotPassword(ctx, req)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) || errors.Is(err, domain.ErrNotFound) {
			helper.Error(c, http.StatusNotFound, "User email not found", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to initiate password reset", err)
		return
	}

	helper.Success(c, http.StatusOK, "Password reset email sent", resp)
}

func (a *AuthHandler) ResetPassword(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	if err := a.authService.ResetPassword(ctx, req); err != nil {
		if errors.Is(err, domain.ErrInvalidToken) {
			helper.Error(c, http.StatusBadRequest, "Invalid or expired token", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to reset password", err)
		return
	}

	helper.Success(c, http.StatusOK, "Password reset successfully", nil)
}

func (a *AuthHandler) Logout(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	authHeader := c.GetHeader("Authorization")
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		helper.Error(c, http.StatusUnauthorized, "Missing authorization header", domain.ErrInvalidToken)
		return
	}

	token := fields[1]
	if err := a.authService.Logout(ctx, token); err != nil {
		helper.Error(c, http.StatusUnauthorized, "Failed to logout", err)
		return
	}

	helper.Success(c, http.StatusOK, "Logged out successfully", nil)
}

func (a *AuthHandler) ListWorkspaces(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	userID, ok := helper.UserIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized: user context missing", nil)
		return
	}

	workspaces, err := a.authService.ListWorkspaces(ctx, userID)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to list workspaces", err)
		return
	}

	helper.Success(c, http.StatusOK, "Workspaces retrieved successfully", workspaces)
}

func (a *AuthHandler) SwitchWorkspace(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	userID, ok := helper.UserIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized: user context missing", nil)
		return
	}

	var req SwitchWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	resp, err := a.authService.SwitchWorkspace(ctx, userID, req)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			helper.Error(c, http.StatusForbidden, "User is not a member of the workspace", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to switch workspace", err)
		return
	}

	helper.Success(c, http.StatusOK, "Workspace switched successfully", resp)
}

func (a *AuthHandler) CompleteOnboarding(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	userID, ok := helper.UserIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized: user context missing", nil)
		return
	}
	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized: tenant context missing", nil)
		return
	}

	var req OnboardingRequest
	if err := c.ShouldBind(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid form data", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	var logoBytes []byte
	var logoFilename string
	fileHeader, err := c.FormFile("logo")
	if err == nil && fileHeader != nil {
		f, openErr := fileHeader.Open()
		if openErr == nil {
			defer f.Close()
			bytes, readErr := io.ReadAll(f)
			if readErr == nil {
				logoBytes = bytes
				logoFilename = fileHeader.Filename
			}
		}
	}

	resp, err := a.authService.CompleteOnboarding(ctx, userID, tenantID, req, logoBytes, logoFilename)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to complete onboarding", err)
		return
	}

	helper.Success(c, http.StatusOK, "Onboarding completed successfully", resp)
}
