package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/dto"
	"github.com/parxyws/cozybox/internal/pkg/helper"
	"github.com/parxyws/cozybox/internal/pkg/validator"
	"github.com/parxyws/cozybox/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account and tenant. Sends OTP to email for verification.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.RegisterUserRequest true "Registration data"
// @Success      201 {object} dto.SwaggerDocResponse "User registered"
// @Failure      400 {object} dto.SwaggerDocResponse "Invalid request"
// @Failure      500 {object} dto.SwaggerDocResponse "Server error"
// @Router       /auth/register [post]
func (a *AuthHandler) Register(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.RegisterUserRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := a.authService.Register(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to register user", err)
		return
	}

	helper.Success(c, http.StatusCreated, "User registered successfully", result)
}

// VerifyEmail godoc
// @Summary      Verify email with OTP
// @Description  Verifies user email using the OTP token sent during registration.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.VerifyEmailRequest true "OTP verification data"
// @Success      200 {object} dto.SwaggerDocResponse "Email verified"
// @Failure      400 {object} dto.SwaggerDocResponse "Invalid request"
// @Failure      500 {object} dto.SwaggerDocResponse "Server error"
// @Router       /auth/verify-email [post]
func (a *AuthHandler) VerifyEmail(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.VerifyEmailRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := a.authService.VerifyEmail(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to verify email", err)
		return
	}

	helper.Success(c, http.StatusOK, "User email verified successfully", result)
}

// Login godoc
// @Summary      Login
// @Description  Authenticates user with email and password. Returns JWT access and refresh tokens.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.LoginRequest true "Login credentials"
// @Success      200 {object} dto.SwaggerDocResponse "Login successful"
// @Failure      400 {object} dto.SwaggerDocResponse "Invalid request"
// @Failure      401 {object} dto.SwaggerDocResponse "Invalid credentials"
// @Router       /auth/login [post]
func (a *AuthHandler) Login(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.LoginRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := a.authService.Login(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusUnauthorized, "Login failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Login successful", result)
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Exchanges a valid refresh token for a new access/refresh token pair.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.RefreshTokenRequest true "Refresh token"
// @Success      200 {object} dto.SwaggerDocResponse "Tokens refreshed"
// @Failure      400 {object} dto.SwaggerDocResponse "Invalid request"
// @Failure      401 {object} dto.SwaggerDocResponse "Invalid refresh token"
// @Router       /auth/refresh-token [post]
func (a *AuthHandler) RefreshToken(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.RefreshTokenRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := a.authService.RefreshToken(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusUnauthorized, "Token refresh failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Token refreshed successfully", result)
}

// Logout godoc
// @Summary      Logout
// @Description  Invalidates the current session and refresh token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Produce      json
// @Success      200 {object} dto.SwaggerDocResponse "Logged out"
// @Failure      400 {object} dto.SwaggerDocResponse "Session not found"
// @Security     BearerAuth
// @Router       /auth/logout [post]
func (a *AuthHandler) Logout(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	sessionID, exists := c.Get("session_id")
	if !exists {
		helper.Error(c, http.StatusBadRequest, "Session not found", nil)
		return
	}

	if err := a.authService.Logout(ctx, sessionID.(string)); err != nil {
		helper.Error(c, http.StatusInternalServerError, "Logout failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Logged out successfully", nil)
}

// ForgotPassword godoc
// @Summary      Request password reset
// @Description  Sends a password reset OTP to the user's email.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.ForgotPasswordRequest true "Email for reset"
// @Success      200 {object} dto.SwaggerDocResponse "OTP sent"
// @Failure      400 {object} dto.SwaggerDocResponse "Invalid request"
// @Failure      401 {object} dto.SwaggerDocResponse "User not found"
// @Router       /auth/forgot-password [post]
func (a *AuthHandler) ForgotPassword(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := a.authService.ForgotPassword(ctx, &req); err != nil {
		helper.Error(c, http.StatusUnauthorized, "Forgot password failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Request Forgot password successfully", nil)
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Resets password using the OTP token sent via forgot-password.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.ResetPasswordRequest true "Reset password data"
// @Success      200 {object} dto.SwaggerDocResponse "Password reset"
// @Failure      400 {object} dto.SwaggerDocResponse "Invalid request"
// @Failure      401 {object} dto.SwaggerDocResponse "Invalid OTP"
// @Router       /auth/reset-password [post]
func (a *AuthHandler) ResetPassword(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.ResetPasswordRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := a.authService.ResetPassword(ctx, &req); err != nil {
		helper.Error(c, http.StatusUnauthorized, "Reset password failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Request Reset password successfully", nil)
}
