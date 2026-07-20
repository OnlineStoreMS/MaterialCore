package repo

import (
	"materialcore/internal/seed"

	"gorm.io/gorm"
)

type Repos struct {
	db       *gorm.DB
	Category *CategoryRepo
	Material *MaterialRepo
}

func New(db *gorm.DB) *Repos {
	return &Repos{
		db:       db,
		Category: NewCategoryRepo(db),
		Material: NewMaterialRepo(db),
	}
}

func (r *Repos) EnsureDefaultCategories(tenantID uint64) error {
	return seed.EnsureDefaultCategories(r.db, tenantID)
}

func NormalizeTenantID(id uint64) uint64 {
	if id == 0 {
		return 1
	}
	return id
}

func scopeTenant(tenantID uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", NormalizeTenantID(tenantID))
	}
}
