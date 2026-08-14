package contact_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/app/contact"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockContactStore struct {
	mock.Mock
}

func (m *mockContactStore) InsertContact(ctx context.Context, contact *domain.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}
func (m *mockContactStore) UpdateContact(ctx context.Context, contact *domain.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}
func (m *mockContactStore) GetContactByIDAndTenant(ctx context.Context, id string, tenantID string) (*domain.Contact, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Contact), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockContactStore) SoftDeleteContact(ctx context.Context, id string, tenantID string) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}
func (m *mockContactStore) ListContacts(ctx context.Context, tenantID string, search string, role string, page int, limit int) ([]domain.Contact, int64, error) {
	args := m.Called(ctx, tenantID, search, role, page, limit)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Contact), int64(args.Int(1)), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

type mockOrgStore struct {
	mock.Mock
}

func (m *mockOrgStore) GetOrganizationByTenantID(ctx context.Context, tenantID string) (*domain.Organization, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Organization), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockOrgStore) InsertOrganization(ctx context.Context, org *domain.Organization) error {
	return nil
}
func (m *mockOrgStore) UpdateOrganization(ctx context.Context, org *domain.Organization) error {
	return nil
}

func setupHandlerTest(t *testing.T) (*gin.Engine, *mockContactStore, *mockOrgStore) {
	gin.SetMode(gin.TestMode)
	mockContactStore := &mockContactStore{}
	mockOrgStore := &mockOrgStore{}
	svc := contact.NewContactService(mockContactStore, mockOrgStore)
	handler := contact.NewContactHandler(svc)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, config.TenantID, "tenant-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	api := router.Group("/api")
	handler.RegisterRoutes(api)

	return router, mockContactStore, mockOrgStore
}

func TestHandler_CreateContact(t *testing.T) {
	router, mockContactStore, mockOrgStore := setupHandlerTest(t)

	reqDto := contact.CreateContactRequest{
		Roles: []string{"client"},
		Name:  "Test Client",
		Email: "test@client.com",
	}

	mockOrgStore.On("GetOrganizationByTenantID", mock.Anything, "tenant-1").Return(&domain.Organization{Id: "org-1"}, nil)
	mockContactStore.On("InsertContact", mock.Anything, mock.MatchedBy(func(c *domain.Contact) bool {
		return c.Name == reqDto.Name && c.Email == reqDto.Email && c.TenantId == "tenant-1"
	})).Return(nil)

	body, _ := json.Marshal(reqDto)
	req, _ := http.NewRequest("POST", "/api/contacts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockContactStore.AssertExpectations(t)
	mockOrgStore.AssertExpectations(t)
}

func TestHandler_GetContact(t *testing.T) {
	router, mockContactStore, _ := setupHandlerTest(t)

	mockContact := &domain.Contact{
		Id:       "contact-1",
		TenantId: "tenant-1",
		Name:     "Test Client",
		Email:    "test@client.com",
	}

	mockContactStore.On("GetContactByIDAndTenant", mock.Anything, "contact-1", "tenant-1").Return(mockContact, nil)

	req, _ := http.NewRequest("GET", "/api/contacts/contact-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockContactStore.AssertExpectations(t)
}

func TestHandler_DeleteContact(t *testing.T) {
	router, mockContactStore, _ := setupHandlerTest(t)

	mockContactStore.On("SoftDeleteContact", mock.Anything, "contact-1", "tenant-1").Return(nil)

	req, _ := http.NewRequest("DELETE", "/api/contacts/contact-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockContactStore.AssertExpectations(t)
}
