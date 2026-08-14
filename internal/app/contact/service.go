package contact

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/store"
)

type ContactService interface {
	CreateContact(ctx context.Context, tenantID string, req CreateContactRequest) (*ContactResponse, error)
	GetContact(ctx context.Context, id string, tenantID string) (*ContactResponse, error)
	UpdateContact(ctx context.Context, id string, tenantID string, req UpdateContactRequest) (*ContactResponse, error)
	DeleteContact(ctx context.Context, id string, tenantID string) error
	ListContacts(ctx context.Context, tenantID string, search string, role string, page int, limit int) (*ListContactsResponse, error)
}

type contactService struct {
	contactStore store.ContactStore
	orgStore     store.OrganizationStore
}

func NewContactService(contactStore store.ContactStore, orgStore store.OrganizationStore) ContactService {
	return &contactService{
		contactStore: contactStore,
		orgStore:     orgStore,
	}
}

func (s *contactService) CreateContact(ctx context.Context, tenantID string, req CreateContactRequest) (*ContactResponse, error) {
	org, err := s.orgStore.GetOrganizationByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	contact := &domain.Contact{
		Id:             ulid.Make().String(),
		TenantId:       tenantID,
		OrganizationId: org.Id,
		Roles:          req.Roles,
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		AddressLine1:   req.AddressLine1,
		AddressLine2:   req.AddressLine2,
		City:           req.City,
		State:          req.State,
		PostalCode:     req.PostalCode,
		Country:        req.Country,
		TaxId:          req.TaxId,
		Notes:          req.Notes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.contactStore.InsertContact(ctx, contact); err != nil {
		return nil, err
	}

	return mapToResponse(contact), nil
}

func (s *contactService) GetContact(ctx context.Context, id string, tenantID string) (*ContactResponse, error) {
	contact, err := s.contactStore.GetContactByIDAndTenant(ctx, id, tenantID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return mapToResponse(contact), nil
}

func (s *contactService) UpdateContact(ctx context.Context, id string, tenantID string, req UpdateContactRequest) (*ContactResponse, error) {
	contact, err := s.contactStore.GetContactByIDAndTenant(ctx, id, tenantID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if len(req.Roles) > 0 {
		contact.Roles = req.Roles
	}
	if req.Name != "" {
		contact.Name = req.Name
	}
	contact.Email = req.Email
	contact.Phone = req.Phone
	contact.AddressLine1 = req.AddressLine1
	contact.AddressLine2 = req.AddressLine2
	contact.City = req.City
	contact.State = req.State
	contact.PostalCode = req.PostalCode
	contact.Country = req.Country
	contact.TaxId = req.TaxId
	contact.Notes = req.Notes
	contact.UpdatedAt = time.Now()

	if err := s.contactStore.UpdateContact(ctx, contact); err != nil {
		return nil, err
	}

	return mapToResponse(contact), nil
}

func (s *contactService) DeleteContact(ctx context.Context, id string, tenantID string) error {
	err := s.contactStore.SoftDeleteContact(ctx, id, tenantID)
	if err != nil && errors.Is(err, store.ErrNotFound) {
		return domain.ErrNotFound
	}
	return err
}

func (s *contactService) ListContacts(ctx context.Context, tenantID string, search string, role string, page int, limit int) (*ListContactsResponse, error) {
	contacts, total, err := s.contactStore.ListContacts(ctx, tenantID, search, role, page, limit)
	if err != nil {
		return nil, err
	}

	data := make([]ContactResponse, len(contacts))
	for i, c := range contacts {
		data[i] = *mapToResponse(&c)
	}

	return &ListContactsResponse{
		Data:       data,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func mapToResponse(c *domain.Contact) *ContactResponse {
	return &ContactResponse{
		Id:           c.Id,
		Roles:        c.Roles,
		Name:         c.Name,
		Email:        c.Email,
		Phone:        c.Phone,
		AddressLine1: c.AddressLine1,
		AddressLine2: c.AddressLine2,
		City:         c.City,
		State:        c.State,
		PostalCode:   c.PostalCode,
		Country:      c.Country,
		TaxId:        c.TaxId,
		Notes:        c.Notes,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}
