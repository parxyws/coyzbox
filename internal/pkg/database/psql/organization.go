package psql

import (
	"context"

	"github.com/parxyws/cozybox/internal/domain"
	"gorm.io/gorm"
)

type OrgRepo struct{ DB *gorm.DB }

func NewOrgRepo(db *gorm.DB) *OrgRepo { return &OrgRepo{DB: db} }

func (r *OrgRepo) Insert(ctx context.Context, o *domain.Organization) error {
	return r.DB.WithContext(ctx).Create(o).Error
}
