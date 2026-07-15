package tenant

import (
	"context"
	"time"

	"github.com/parxyws/cozybox/internal/domain"
)

type Service struct {
	orgRepo      OrganizationRepository
	templateRepo TemplateConfigRepository
	fileStorage  FileStorage
}

func NewService(orgRepo OrganizationRepository, templateRepo TemplateConfigRepository, fileStorage FileStorage) *Service {
	return &Service{
		orgRepo:      orgRepo,
		templateRepo: templateRepo,
		fileStorage:  fileStorage,
	}
}

func (s *Service) GetOrganization(ctx context.Context, tenantID string) (*OrganizationResponse, error) {
	org, err := s.orgRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

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
	}, nil
}

func (s *Service) UpdateOrganization(ctx context.Context, tenantID string, req *UpdateOrganizationRequest, logo *domain.UploadInput) (*OrganizationResponse, error) {
	org, err := s.orgRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if logo != nil {
		key, err := s.fileStorage.PutObject(ctx, *logo)
		if err != nil {
			return nil, err
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
	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, err
	}

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
	}, nil
}

func (s *Service) ListTemplateConfigs(ctx context.Context, tenantID string) ([]TemplateConfigResponse, error) {
	cfgs, err := s.templateRepo.ListByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	out := make([]TemplateConfigResponse, 0, len(cfgs))
	for _, cfg := range cfgs {
		out = append(out, TemplateConfigResponse{
			Id:       cfg.Id,
			TenantId: cfg.TenantId,
			BaseType: string(cfg.BaseType),
			Status:   string(cfg.Status),
			Name:     cfg.Name,
			Config:   cfg.Config,
		})
	}
	return out, nil
}

func (s *Service) UpdateTemplateConfig(ctx context.Context, tenantID, configID string, req *UpdateTemplateConfigRequest) (*TemplateConfigResponse, error) {
	cfg, err := s.templateRepo.GetByID(ctx, configID)
	if err != nil {
		return nil, err
	}
	if cfg.TenantId != tenantID {
		return nil, domain.ErrForbidden
	}

	if req.Status != nil {
		cfg.Status = domain.TemplateStatus(*req.Status)
	}
	if req.Config != nil {
		cfg.Config = req.Config
	}
	cfg.UpdatedAt = time.Now()

	if err := s.templateRepo.Update(ctx, cfg); err != nil {
		return nil, err
	}

	return &TemplateConfigResponse{
		Id:       cfg.Id,
		TenantId: cfg.TenantId,
		BaseType: string(cfg.BaseType),
		Status:   string(cfg.Status),
		Name:     cfg.Name,
		Config:   cfg.Config,
	}, nil
}

func applyOrgField(val string, target *string) {
	if val != "" {
		*target = val
	}
}
