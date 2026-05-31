package psql

import "gorm.io/gorm"

func TenantDetail(userId string, tenantId string, organizationDetail bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if organizationDetail {
			return db.Preload("Organizations").Where("tenant_id = ? AND user_ = ?", tenantId, userId)
		}

		return db.Where("tenant_id = ? AND user_id = ?", tenantId, userId)
	}
}
