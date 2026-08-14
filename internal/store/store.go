package store

import (
	"gorm.io/gorm"
)

type Store interface {
	User() UserStore
	Tenant() TenantStore
	Organization() OrganizationStore
	Contact() ContactStore
	Document() DocumentStore
	Sequence() SequenceStore
	Activity() ActivityStore
}

type psqlStore struct {
	db       *gorm.DB
	user     UserStore
	tenant   TenantStore
	org      OrganizationStore
	contact  ContactStore
	doc      DocumentStore
	sequence SequenceStore
	activity ActivityStore
}

func NewStore(db *gorm.DB) Store {
	return &psqlStore{
		db:       db,
		user:     NewUserStore(db),
		tenant:   NewTenantStore(db),
		org:      NewOrganizationStore(db),
		contact:  NewContactStore(db),
		doc:      NewDocumentStore(db),
		sequence: NewSequenceStore(db),
		activity: NewActivityStore(db),
	}
}

func (s *psqlStore) User() UserStore                 { return s.user }
func (s *psqlStore) Tenant() TenantStore             { return s.tenant }
func (s *psqlStore) Organization() OrganizationStore { return s.org }
func (s *psqlStore) Contact() ContactStore           { return s.contact }
func (s *psqlStore) Document() DocumentStore         { return s.doc }
func (s *psqlStore) Sequence() SequenceStore         { return s.sequence }
func (s *psqlStore) Activity() ActivityStore         { return s.activity }
