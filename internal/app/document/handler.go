package document

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/pkg/helper"
)

type Handler struct {
	engine Engine
}

func NewHandler(engine Engine) *Handler {
	return &Handler{engine: engine}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	docs := rg.Group("/documents")
	{
		docs.POST("/draft", h.ReserveDraft)
		docs.POST("/:id/publish", h.PublishArtifact)
		docs.POST("/:id/issue", h.IssueDocument)
		docs.PATCH("/:id/flow-status", h.TransitionFlowStatus)
		docs.POST("/:id/payments", h.RecordPayment)
		docs.POST("/:id/convert", h.ConvertDocument)
		docs.POST("/:id/cancel", h.CancelDocument)
		docs.GET("/:id", h.GetDocument)
		docs.DELETE("/:id", h.DeleteDocument)
		docs.GET("", h.ListDocuments)
	}
}

func (h *Handler) ReserveDraft(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}
	userID, _ := c.Get("user_id")

	var req CreateDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request payload: "+err.Error(), err)
		return
	}

	userIDStr, _ := userID.(string)
	resp, err := h.engine.ReserveSequenceAndDraft(ctx, tenantID, userIDStr, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	helper.Success(c, http.StatusCreated, "Sequence reserved and draft created successfully", resp)
}

func (h *Handler) PublishArtifact(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}
	userID, _ := c.Get("user_id")
	docID := c.Param("id")

	var req PublishArtifactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request payload: "+err.Error(), err)
		return
	}

	pdfBytes, err := base64.StdEncoding.DecodeString(req.PdfBase64)
	if err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid base64 PDF payload: "+err.Error(), err)
		return
	}

	userIDStr, _ := userID.(string)
	resp, err := h.engine.PublishArtifact(ctx, tenantID, userIDStr, docID, pdfBytes)
	if err != nil {
		helper.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	helper.Success(c, http.StatusOK, "Document PDF published successfully", resp)
}

func (h *Handler) IssueDocument(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}
	userID, _ := c.Get("user_id")
	docID := c.Param("id")

	var req struct {
		IssueDate *time.Time `json:"issue_date"`
	}
	_ = c.ShouldBindJSON(&req)

	userIDStr, _ := userID.(string)
	resp, err := h.engine.IssueDocument(ctx, tenantID, userIDStr, docID, req.IssueDate)
	if err != nil {
		helper.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	helper.Success(c, http.StatusOK, "Document issued successfully", resp)
}

func (h *Handler) TransitionFlowStatus(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}
	userID, _ := c.Get("user_id")
	docID := c.Param("id")

	var req TransitionFlowStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request payload: "+err.Error(), err)
		return
	}

	userIDStr, _ := userID.(string)
	resp, err := h.engine.TransitionFlowStatus(ctx, tenantID, userIDStr, docID, req.FlowStatus)
	if err != nil {
		helper.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	helper.Success(c, http.StatusOK, "Document flow status updated successfully", resp)
}

func (h *Handler) RecordPayment(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}
	userID, _ := c.Get("user_id")
	docID := c.Param("id")

	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request payload: "+err.Error(), err)
		return
	}

	userIDStr, _ := userID.(string)
	resp, err := h.engine.RecordPayment(ctx, tenantID, userIDStr, docID, &req)
	if err != nil {
		helper.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	helper.Success(c, http.StatusOK, "Payment recorded successfully", resp)
}

func (h *Handler) ConvertDocument(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}
	userID, _ := c.Get("user_id")
	docID := c.Param("id")

	var req ConvertDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request payload: "+err.Error(), err)
		return
	}

	userIDStr, _ := userID.(string)
	resp, err := h.engine.ConvertDocument(ctx, tenantID, userIDStr, docID, req.TargetType)
	if err != nil {
		helper.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	helper.Success(c, http.StatusCreated, "Document converted successfully", resp)
}

func (h *Handler) CancelDocument(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}
	userID, _ := c.Get("user_id")
	docID := c.Param("id")

	var req CancelDocumentRequest
	_ = c.ShouldBindJSON(&req)

	userIDStr, _ := userID.(string)
	resp, err := h.engine.CancelDocument(ctx, tenantID, userIDStr, docID, req.Reason)
	if err != nil {
		helper.Error(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	helper.Success(c, http.StatusOK, "Document cancelled successfully", resp)
}

func (h *Handler) GetDocument(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}
	docID := c.Param("id")

	resp, err := h.engine.GetDocument(ctx, tenantID, docID)
	if err != nil {
		if err == domain.ErrNotFound {
			helper.Error(c, http.StatusNotFound, "Document not found", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	helper.Success(c, http.StatusOK, "Document retrieved successfully", resp)
}

func (h *Handler) ListDocuments(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}

	docType := c.Query("type")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	resp, total, err := h.engine.ListDocuments(ctx, tenantID, docType, status, page, limit)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	helper.Success(c, http.StatusOK, "Documents retrieved successfully", gin.H{
		"documents": resp,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

func (h *Handler) DeleteDocument(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	tenantID, ok := helper.TenantIDFromContext(ctx)
	if !ok {
		helper.Error(c, http.StatusUnauthorized, "Tenant ID not found in context", nil)
		return
	}
	docID := c.Param("id")

	err := h.engine.DeleteDocument(ctx, tenantID, docID)
	if err != nil {
		if errors.Is(err, domain.ErrDocumentImmutable) {
			helper.Error(c, http.StatusUnprocessableEntity, "Published documents are permanent compliance records and cannot be deleted", err)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			helper.Error(c, http.StatusNotFound, "Document not found", err)
			return
		}
		helper.Error(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	helper.Success(c, http.StatusOK, "Document deleted successfully", nil)
}
