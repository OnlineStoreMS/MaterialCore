package repo

import (
	"materialcore/internal/model"

	"gorm.io/gorm"
)

type CategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) ListAll(tenantID uint64) ([]model.MaterialCategory, error) {
	var list []model.MaterialCategory
	err := r.db.Scopes(scopeTenant(tenantID)).
		Order("sort ASC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *CategoryRepo) Get(tenantID, id uint64) (*model.MaterialCategory, error) {
	var row model.MaterialCategory
	err := r.db.Scopes(scopeTenant(tenantID)).First(&row, id).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CategoryRepo) GetByCode(tenantID uint64, code string) (*model.MaterialCategory, error) {
	var row model.MaterialCategory
	err := r.db.Scopes(scopeTenant(tenantID)).Where("code = ?", code).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CategoryRepo) Create(row *model.MaterialCategory) error {
	return r.db.Create(row).Error
}

func (r *CategoryRepo) Update(row *model.MaterialCategory) error {
	return r.db.Save(row).Error
}

func (r *CategoryRepo) Delete(tenantID, id uint64) error {
	return r.db.Scopes(scopeTenant(tenantID)).Delete(&model.MaterialCategory{}, id).Error
}

func (r *CategoryRepo) CountChildren(tenantID, parentID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.MaterialCategory{}).
		Scopes(scopeTenant(tenantID)).
		Where("parent_id = ?", parentID).
		Count(&n).Error
	return n, err
}

func (r *CategoryRepo) CountByTenant(tenantID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.MaterialCategory{}).Scopes(scopeTenant(tenantID)).Count(&n).Error
	return n, err
}

func (r *CategoryRepo) ExistsName(tenantID, parentID uint64, name string, excludeID uint64) (bool, error) {
	q := r.db.Model(&model.MaterialCategory{}).
		Scopes(scopeTenant(tenantID)).
		Where("parent_id = ? AND name = ?", parentID, name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *CategoryRepo) ListDescendantIDs(tenantID, rootID uint64) ([]uint64, error) {
	all, err := r.ListAll(tenantID)
	if err != nil {
		return nil, err
	}
	children := map[uint64][]uint64{}
	for _, c := range all {
		children[c.ParentID] = append(children[c.ParentID], c.ID)
	}
	var out []uint64
	var walk func(uint64)
	walk = func(id uint64) {
		out = append(out, id)
		for _, cid := range children[id] {
			walk(cid)
		}
	}
	walk(rootID)
	return out, nil
}
