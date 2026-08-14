package tenant

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/store"
)

type TenantService interface {
	GetOrganization(ctx context.Context, tenantID string) (*OrganizationResponse, error)
	UpdateOrganization(ctx context.Context, tenantID string, req UpdateOrganizationRequest, logoFile []byte, logoFilename string, contentType string) (*OrganizationResponse, error)
	ListTemplateConfigs(ctx context.Context, tenantID string) ([]TemplateConfigResponse, error)
	UpdateTemplateConfig(ctx context.Context, tenantID string, configID string, req UpdateTemplateConfigRequest) (*TemplateConfigResponse, error)
}

type tenantService struct {
	tenantStore store.TenantStore
	orgStore    store.OrganizationStore
	fileStorage FileStorage
}

func NewTenantService(tenantStore store.TenantStore, orgStore store.OrganizationStore, fileStorage FileStorage) TenantService {
	return &tenantService{
		tenantStore: tenantStore,
		orgStore:    orgStore,
		fileStorage: fileStorage,
	}
}

func (s *tenantService) GetOrganization(ctx context.Context, tenantID string) (*OrganizationResponse, error) {
	org, err := s.orgStore.GetOrganizationByTenantID(ctx, tenantID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return mapOrgToResponse(org), nil
}

func (s *tenantService) UpdateOrganization(ctx context.Context, tenantID string, req UpdateOrganizationRequest, logoFile []byte, logoFilename string, contentType string) (*OrganizationResponse, error) {
	org, err := s.orgStore.GetOrganizationByTenantID(ctx, tenantID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if len(logoFile) > 0 {
		logoInput := domain.UploadInput{
			Object:      bytes.NewReader(logoFile),
			ObjectName:  fmt.Sprintf("tenants/%s/logos/%s", tenantID, logoFilename),
			ObjectSize:  int64(len(logoFile)),
			ContentType: contentType,
		}
		key, err := s.fileStorage.PutObject(ctx, logoInput)
		if err != nil {
			return nil, fmt.Errorf("failed to upload logo: %w", err)
		}
		org.LogoS3Key = key
	}

	applyOrgField(req.Name, &org.Name)
	applyOrgField(req.Email, &org.Email)
	applyOrgField(req.Phone, &org.Phone)
	applyOrgField(req.AddressLine1, &org.AddressLine1)
	applyOrgField(req.AddressLine2, &org.AddressLine2)
	applyOrgField(req.City, &org.City)
	applyOrgField(req.State, &org.State)
	applyOrgField(req.PostalCode, &org.PostalCode)
	applyOrgField(req.Country, &org.Country)
	applyOrgField(req.TaxId, &org.TaxId)
	applyOrgField(req.Website, &org.Website)
	applyOrgField(req.Timezone, &org.Timezone)
	applyOrgField(req.DefaultCurrency, &org.DefaultCurrency)

	org.UpdatedAt = time.Now()
	if err := s.orgStore.UpdateOrganization(ctx, org); err != nil {
		return nil, err
	}

	return mapOrgToResponse(org), nil
}

func (s *tenantService) ListTemplateConfigs(ctx context.Context, tenantID string) ([]TemplateConfigResponse, error) {
	configs, err := s.tenantStore.ListTemplateConfigsByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]TemplateConfigResponse, len(configs))
	for i, cfg := range configs {
		result[i] = TemplateConfigResponse{
			Id:        cfg.Id,
			TenantId:  cfg.TenantId,
			BaseType:  string(cfg.BaseType),
			Status:    string(cfg.Status),
			Name:      cfg.Name,
			Config:    cfg.Config,
			CreatedAt: cfg.CreatedAt,
			UpdatedAt: cfg.UpdatedAt,
		}
	}

	return result, nil
}

func (s *tenantService) UpdateTemplateConfig(ctx context.Context, tenantID string, configID string, req UpdateTemplateConfigRequest) (*TemplateConfigResponse, error) {
	cfg, err := s.tenantStore.GetTemplateConfigByID(ctx, configID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if cfg.TenantId != tenantID {
		return nil, domain.ErrUnauthorized
	}

	if req.Name != nil {
		cfg.Name = *req.Name
	}
	if req.Status != nil {
		cfg.Status = domain.TemplateStatus(*req.Status)
	}
	if req.Config != nil {
		cfg.Config = req.Config
	}
	cfg.UpdatedAt = time.Now()

	if err := s.tenantStore.UpdateTemplateConfig(ctx, cfg); err != nil {
		return nil, err
	}

	return &TemplateConfigResponse{
		Id:        cfg.Id,
		TenantId:  cfg.TenantId,
		BaseType:  string(cfg.BaseType),
		Status:    string(cfg.Status),
		Name:      cfg.Name,
		Config:    cfg.Config,
		CreatedAt: cfg.CreatedAt,
		UpdatedAt: cfg.UpdatedAt,
	}, nil
}

func mapOrgToResponse(org *domain.Organization) *OrganizationResponse {
	return &OrganizationResponse{
		Id:              org.Id,
		TenantId:        org.TenantId,
		Name:            org.Name,
		Email:           org.Email,
		Phone:           org.Phone,
		AddressLine1:    org.AddressLine1,
		AddressLine2:    org.AddressLine2,
		City:            org.City,
		State:           org.State,
		PostalCode:      org.PostalCode,
		Country:         org.Country,
		TaxId:           org.TaxId,
		LogoS3Key:       org.LogoS3Key,
		Website:         org.Website,
		Timezone:        org.Timezone,
		DefaultCurrency: org.DefaultCurrency,
	}
}

func applyOrgField(val string, target *string) {
	if val != "" {
		*target = val
	}
}
