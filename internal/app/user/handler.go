package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/pkg/helper"
	"github.com/parxyws/cozybox/internal/pkg/validator"
)

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) RegisterRoutes(route *gin.RouterGroup) {
	userGroup := route.Group("/user")
	{
		userGroup.GET("/profile", h.GetProfile)
		userGroup.PUT("/profile", h.UpdateProfile)
		userGroup.PUT("/password", h.UpdatePassword)
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	userID, ok := helper.UserIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized: user context missing", nil)
		return
	}

	profile, err := h.userService.GetProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			helper.Error(c, http.StatusNotFound, "User not found", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to get profile", err)
		return
	}

	helper.Success(c, http.StatusOK, "Profile retrieved successfully", profile)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	userID, ok := helper.UserIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized: user context missing", nil)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	profile, err := h.userService.UpdateProfile(ctx, userID, req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			helper.Error(c, http.StatusNotFound, "User not found", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to update profile", err)
		return
	}

	helper.Success(c, http.StatusOK, "Profile updated successfully", profile)
}

func (h *UserHandler) UpdatePassword(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	userID, ok := helper.UserIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized: user context missing", nil)
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	if err := h.userService.UpdatePassword(ctx, userID, req); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			helper.Error(c, http.StatusNotFound, "User not found", err)
			return
		}
		if errors.Is(err, domain.ErrInvalidCredential) {
			helper.Error(c, http.StatusUnauthorized, "Current password is incorrect", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to update password", err)
		return
	}

	helper.Success(c, http.StatusOK, "Password updated successfully", nil)
}
