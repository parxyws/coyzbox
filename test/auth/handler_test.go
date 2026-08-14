package auth_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuthHandlerTest(t *testing.T) (*gin.Engine, *serviceMocks) {
	gin.SetMode(gin.TestMode)
	handler, m := newHandlerWithMocks(t)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, config.UserID, "user-1")
		ctx = context.WithValue(ctx, config.TenantID, "tenant-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	api := router.Group("/api")
	handler.RegisterProtectedRoutes(api)

	return router, m
}

func TestHandler_CompleteOnboarding(t *testing.T) {
	router, m := setupAuthHandlerTest(t)

	m.tenantStore.On("GetTenantByID", mock.Anything, "tenant-1").Return(&domain.Tenant{Id: "tenant-1", Name: "Test Tenant"}, nil)
	m.tenantStore.On("UpdateTenant", mock.Anything, mock.Anything).Return(nil)
	m.userStore.On("GetUserByID", mock.Anything, "user-1").Return(&domain.User{Id: "user-1"}, nil)
	m.orgStore.On("GetOrganizationByTenantID", mock.Anything, "tenant-1").Return(&domain.Organization{Id: "org-1", TenantId: "tenant-1"}, nil)
	m.orgStore.On("UpdateOrganization", mock.Anything, mock.Anything).Return(nil)
	m.userStore.On("UpdateUser", mock.Anything, mock.Anything).Return(nil)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("tenant_name", "Test Tenant")
	_ = writer.WriteField("email", "test@tenant.com")
	_ = writer.WriteField("address_line1", "123 Main St")
	_ = writer.WriteField("country", "US")
	_ = writer.WriteField("tax_id", "TAX123")
	_ = writer.WriteField("default_currency", "USD")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/onboarding", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
