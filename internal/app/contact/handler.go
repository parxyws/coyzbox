package contact

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/pkg/helper"
	"github.com/parxyws/cozybox/internal/pkg/validator"
)

type ContactHandler struct {
	contactService ContactService
}

func NewContactHandler(contactService ContactService) *ContactHandler {
	return &ContactHandler{contactService: contactService}
}

func (h *ContactHandler) RegisterRoutes(route *gin.RouterGroup) {
	contactGroup := route.Group("/contacts")
	{
		contactGroup.GET("", h.ListContacts)
		contactGroup.POST("", h.CreateContact)
		contactGroup.GET("/:id", h.GetContact)
		contactGroup.PUT("/:id", h.UpdateContact)
		contactGroup.DELETE("/:id", h.DeleteContact)
	}
}

func (h *ContactHandler) CreateContact(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}

	var req CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	resp, err := h.contactService.CreateContact(ctx, tenantID, req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to create contact", err)
		return
	}

	helper.Success(c, http.StatusCreated, "Contact created successfully", resp)
}

func (h *ContactHandler) GetContact(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}

	contactID := c.Param("id")
	resp, err := h.contactService.GetContact(ctx, contactID, tenantID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			helper.Error(c, http.StatusNotFound, "Contact not found", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to get contact", err)
		return
	}

	helper.Success(c, http.StatusOK, "Contact retrieved successfully", resp)
}

func (h *ContactHandler) UpdateContact(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}

	contactID := c.Param("id")

	var req UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	resp, err := h.contactService.UpdateContact(ctx, contactID, tenantID, req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			helper.Error(c, http.StatusNotFound, "Contact not found", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to update contact", err)
		return
	}

	helper.Success(c, http.StatusOK, "Contact updated successfully", resp)
}

func (h *ContactHandler) DeleteContact(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}

	contactID := c.Param("id")
	if err := h.contactService.DeleteContact(ctx, contactID, tenantID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			helper.Error(c, http.StatusNotFound, "Contact not found", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, "Failed to delete contact", err)
		return
	}

	helper.Success(c, http.StatusOK, "Contact deleted successfully", nil)
}

func (h *ContactHandler) ListContacts(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}

	search := c.Query("search")
	role := c.Query("role")
	if role == "" {
		role = c.Query("type")
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	resp, err := h.contactService.ListContacts(ctx, tenantID, search, role, page, limit)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to list contacts", err)
		return
	}

	helper.Success(c, http.StatusOK, "Contacts retrieved successfully", resp)
}
