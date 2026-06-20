package server

import (
	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/parxyws/cozybox/internal/pkg/database/psql"
	"gorm.io/gorm"
)

// repoFactory implements auth.RepoFactory using the standard psql repositories scoped to a transaction.
// It lives in the server package to avoid circular dependencies between auth and psql.
type repoFactory struct{}

func (f *repoFactory) UserRepo(tx *gorm.DB) auth.UserRepository     { return psql.NewUserRepo(tx) }
func (f *repoFactory) TenantRepo(tx *gorm.DB) auth.TenantRepository { return psql.NewTenantRepo(tx) }
func (f *repoFactory) MemberRepo(tx *gorm.DB) auth.TenantMemberRepository {
	return psql.NewTenantMemberRepo(tx)
}
func (f *repoFactory) OrgRepo(tx *gorm.DB) auth.OrganizationRepository { return psql.NewOrgRepo(tx) }
