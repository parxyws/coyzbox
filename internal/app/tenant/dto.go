package tenant

import (
	"encoding/json"
	"time"
)

type UpdateOrganizationRequest struct {
	Name            string `json:"name"            validate:"omitempty,min=2,max=200"`
	Email           string `json:"email"           validate:"omitempty,email"`
	Phone           string `json:"phone"           validate:"omitempty,max=30"`
	AddressLine1    string `json:"address_line1"   validate:"omitempty,max=255"`
	AddressLine2    string `json:"address_line2"   validate:"omitempty,max=255"`
	City            string `json:"city"            validate:"omitempty,max=100"`
	State           string `json:"state"           validate:"omitempty,max=100"`
	PostalCode      string `json:"postal_code"     validate:"omitempty,max=20"`
	Country         string `json:"country"         validate:"omitempty,max=100"`
	TaxId           string `json:"tax_id"          validate:"omitempty,max=50"`
	Website         string `json:"website"         validate:"omitempty,max=255"`
	Timezone        string `json:"timezone"        validate:"omitempty,max=50"`
	DefaultCurrency string `json:"default_currency" validate:"omitempty,len=3"`
}

type OrganizationResponse struct {
	Id              string `json:"id"`
	TenantId        string `json:"tenant_id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	AddressLine1    string `json:"address_line1"`
	AddressLine2    string `json:"address_line2"`
	City            string `json:"city"`
	State           string `json:"state"`
	PostalCode      string `json:"postal_code"`
	Country         string `json:"country"`
	TaxId           string `json:"tax_id"`
	LogoS3Key       string `json:"logo_s3_key"`
	Website         string `json:"website"`
	Timezone        string `json:"timezone"`
	DefaultCurrency string `json:"default_currency"`
}

type UpdateTemplateConfigRequest struct {
	Name   *string         `json:"name"     validate:"omitempty,min=2,max=100"`
	Status *string         `json:"status"   validate:"omitempty,oneof=draft published"`
	Config json.RawMessage `json:"config"   validate:"omitempty"`
}

type TemplateConfigResponse struct {
	Id        string          `json:"id"`
	TenantId  string          `json:"tenant_id"`
	BaseType  string          `json:"base_type"`
	Status    string          `json:"status"`
	Name      string          `json:"name"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
