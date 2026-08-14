package tenant

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/pkg/helper"
	"github.com/parxyws/cozybox/internal/pkg/validator"
)

type Handler struct {
	tenantService TenantService
}

func NewHandler(tenantService TenantService) *Handler {
	return &Handler{
		tenantService: tenantService,
	}
}

func (h *Handler) RegisterRoutes(route *gin.RouterGroup) {
	tenantGroup := route.Group("/tenant")
	{
		tenantGroup.GET("/organization", h.GetOrganization)
		tenantGroup.PUT("/organization", h.UpdateOrganization)
		tenantGroup.GET("/template-configs", h.ListTemplateConfigs)
		tenantGroup.PUT("/template-configs/:id", h.UpdateTemplateConfig)
	}
}

func (h *Handler) GetOrganization(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized: tenant context missing", nil)
		return
	}

	resp, err := h.tenantService.GetOrganization(ctx, tenantID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			helper.Error(c, http.StatusNotFound, "Organization profile not found", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to get organization", err)
		return
	}

	helper.Success(c, http.StatusOK, "Organization retrieved successfully", resp)
}

func (h *Handler) UpdateOrganization(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized: tenant context missing", nil)
		return
	}

	var req UpdateOrganizationRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	var logoBytes []byte
	var logoFilename string
	var contentType string
	file, header, err := c.Request.FormFile("logo")
	if err == nil {
		defer file.Close()
		contentType = header.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			helper.Error(c, http.StatusBadRequest, "Logo must be an image file", nil)
			return
		}
		b, readErr := io.ReadAll(file)
		if readErr == nil {
			logoBytes = b
			logoFilename = header.Filename
		}
	}

	resp, err := h.tenantService.UpdateOrganization(ctx, tenantID, req, logoBytes, logoFilename, contentType)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			helper.Error(c, http.StatusNotFound, "Organization profile not found", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to update organization", err)
		return
	}

	helper.Success(c, http.StatusOK, "Organization updated successfully", resp)
}

func (h *Handler) ListTemplateConfigs(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized: tenant context missing", nil)
		return
	}

	result, err := h.tenantService.ListTemplateConfigs(ctx, tenantID)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to list template configs", err)
		return
	}

	helper.Success(c, http.StatusOK, "Template configs retrieved successfully", result)
}

func (h *Handler) UpdateTemplateConfig(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized: tenant context missing", nil)
		return
	}

	configID := c.Param("id")
	if configID == "" {
		helper.Error(c, http.StatusBadRequest, "Template config ID is required", nil)
		return
	}

	var req UpdateTemplateConfigRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	resp, err := h.tenantService.UpdateTemplateConfig(ctx, tenantID, configID, req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			helper.Error(c, http.StatusNotFound, "Template config not found", err)
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			helper.Error(c, http.StatusForbidden, "Forbidden: template config belongs to another tenant", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to update template config", err)
		return
	}

	helper.Success(c, http.StatusOK, "Template config updated successfully", resp)
}
