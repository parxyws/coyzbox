package tenant

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/pkg/helper"
	"github.com/parxyws/cozybox/internal/pkg/validator"
)

// TenantService is the consuming-side interface for tenant operations.
type TenantService interface {
	GetOrganization(ctx context.Context, tenantID string) (*OrganizationResponse, error)
	UpdateOrganization(ctx context.Context, tenantID string, req *UpdateOrganizationRequest, logo *domain.UploadInput) (*OrganizationResponse, error)
	ListTemplateConfigs(ctx context.Context, tenantID string) ([]TemplateConfigResponse, error)
	UpdateTemplateConfig(ctx context.Context, tenantID, configID string, req *UpdateTemplateConfigRequest) (*TemplateConfigResponse, error)
}

type Handler struct {
	service TenantService
}

func NewHandler(service TenantService) *Handler {
	return &Handler{service: service}
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

	tenantID, exists := c.Get(string(config.TenantID))
	if !exists {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	result, err := h.service.GetOrganization(ctx, tenantID.(string))
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to get organization", err)
		return
	}

	helper.Success(c, http.StatusOK, "Organization retrieved successfully", result)
}

func (h *Handler) UpdateOrganization(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, exists := c.Get(string(config.TenantID))
	if !exists {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req UpdateOrganizationRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	var logo *domain.UploadInput
	file, header, err := c.Request.FormFile("logo")
	if err == nil {
		defer file.Close()
		if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
			helper.Error(c, http.StatusBadRequest, "Logo must be an image file", nil)
			return
		}
		logo = &domain.UploadInput{
			Object:      file,
			ObjectName:  header.Filename,
			ObjectSize:  header.Size,
			ContentType: header.Header.Get("Content-Type"),
		}
	}

	result, err := h.service.UpdateOrganization(ctx, tenantID.(string), &req, logo)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to update organization", err)
		return
	}

	helper.Success(c, http.StatusOK, "Organization updated successfully", result)
}

func (h *Handler) ListTemplateConfigs(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, exists := c.Get(string(config.TenantID))
	if !exists {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	result, err := h.service.ListTemplateConfigs(ctx, tenantID.(string))
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to list template configs", err)
		return
	}

	helper.Success(c, http.StatusOK, "Template configs retrieved successfully", result)
}

func (h *Handler) UpdateTemplateConfig(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, exists := c.Get(string(config.TenantID))
	if !exists {
		helper.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
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
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := h.service.UpdateTemplateConfig(ctx, tenantID.(string), configID, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to update template config", err)
		return
	}

	helper.Success(c, http.StatusOK, "Template config updated successfully", result)
}
