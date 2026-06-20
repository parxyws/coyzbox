package user

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/pkg/helper"
	"github.com/parxyws/cozybox/internal/pkg/validator"
)

// UserService is the consuming-side interface for user operations.
type UserService interface {
	GetProfile(ctx context.Context) (*UserProfileResponse, error)
	UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*UserProfileResponse, error)
	UpdatePassword(ctx context.Context, req *UpdatePasswordRequest) error
}

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
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

	result, err := h.service.GetProfile(ctx)
	if err != nil {
		helper.Error(c, http.StatusUnauthorized, "Failed to get profile", err)
		return
	}

	helper.Success(c, http.StatusOK, "Profile retrieved successfully", result)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req UpdateProfileRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := h.service.UpdateProfile(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to update profile", err)
		return
	}

	helper.Success(c, http.StatusOK, "Profile updated successfully", result)
}

func (h *UserHandler) UpdatePassword(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req UpdatePasswordRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.UpdatePassword(ctx, &req); err != nil {
		helper.Error(c, http.StatusUnauthorized, "Failed to update password", err)
		return
	}

	helper.Success(c, http.StatusOK, "Password updated successfully", nil)
}
