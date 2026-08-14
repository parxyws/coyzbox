package contact

import "time"

type CreateContactRequest struct {
	Roles        []string `json:"roles" validate:"required,min=1,dive,oneof=client supplier"`
	Name         string   `json:"name" validate:"required,min=2,max=255"`
	Email        string   `json:"email" validate:"omitempty,email"`
	Phone        string   `json:"phone"`
	AddressLine1 string   `json:"address_line_1"`
	AddressLine2 string   `json:"address_line_2"`
	City         string   `json:"city"`
	State        string   `json:"state"`
	PostalCode   string   `json:"postal_code"`
	Country      string   `json:"country"`
	TaxId        string   `json:"tax_id"`
	Notes        string   `json:"notes"`
}

type UpdateContactRequest struct {
	Roles        []string `json:"roles" validate:"omitempty,min=1,dive,oneof=client supplier"`
	Name         string   `json:"name" validate:"omitempty,min=2,max=255"`
	Email        string   `json:"email" validate:"omitempty,email"`
	Phone        string   `json:"phone"`
	AddressLine1 string   `json:"address_line_1"`
	AddressLine2 string   `json:"address_line_2"`
	City         string   `json:"city"`
	State        string   `json:"state"`
	PostalCode   string   `json:"postal_code"`
	Country      string   `json:"country"`
	TaxId        string   `json:"tax_id"`
	Notes        string   `json:"notes"`
}

type ContactResponse struct {
	Id           string    `json:"id"`
	Roles        []string  `json:"roles"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	AddressLine1 string    `json:"address_line_1"`
	AddressLine2 string    `json:"address_line_2"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	PostalCode   string    `json:"postal_code"`
	Country      string    `json:"country"`
	TaxId        string    `json:"tax_id"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ListContactsResponse struct {
	Data       []ContactResponse `json:"data"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
}

type ListRequest struct {
	Page   int    `form:"page,default=1" validate:"min=1"`
	Limit  int    `form:"limit,default=20" validate:"min=1,max=100"`
	Search string `form:"search"`
	Role   string `form:"role" validate:"omitempty,oneof=client supplier"`
}
