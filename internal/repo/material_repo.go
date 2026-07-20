package repo

import (
	"strings"
	"time"

	"materialcore/internal/model"

	"gorm.io/gorm"
)

type MaterialListFilter struct {
	CategoryIDs []uint64
	MediaType   string
	Keyword     string
	ProductSn   string
	HasProduct  string // yes | no | ""
	Tag         string
	Page        int
	PageSize    int
}

type MaterialRepo struct {
	db *gorm.DB
}

func NewMaterialRepo(db *gorm.DB) *MaterialRepo {
	return &MaterialRepo{db: db}
}

func (r *MaterialRepo) Get(tenantID, id uint64) (*model.Material, error) {
	var row model.Material
	err := r.db.Scopes(scopeTenant(tenantID)).
		Where("status = ?", model.MaterialStatusActive).
		First(&row, id).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *MaterialRepo) GetByIDs(tenantID uint64, ids []uint64) ([]model.Material, error) {
	var list []model.Material
	if len(ids) == 0 {
		return list, nil
	}
	err := r.db.Scopes(scopeTenant(tenantID)).
		Where("status = ? AND id IN ?", model.MaterialStatusActive, ids).
		Find(&list).Error
	return list, err
}

func (r *MaterialRepo) Create(row *model.Material) error {
	return r.db.Create(row).Error
}

func (r *MaterialRepo) Update(row *model.Material) error {
	return r.db.Save(row).Error
}

func (r *MaterialRepo) SoftDelete(tenantID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&model.Material{}).
		Scopes(scopeTenant(tenantID)).
		Where("id IN ?", ids).
		Updates(map[string]any{
			"status":     model.MaterialStatusDeleted,
			"updated_at": time.Now(),
		}).Error
}

func (r *MaterialRepo) CountByCategory(tenantID, categoryID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.Material{}).
		Scopes(scopeTenant(tenantID)).
		Where("category_id = ? AND status = ?", categoryID, model.MaterialStatusActive).
		Count(&n).Error
	return n, err
}

func (r *MaterialRepo) List(tenantID uint64, f MaterialListFilter) ([]model.Material, int64, error) {
	q := r.db.Model(&model.Material{}).
		Scopes(scopeTenant(tenantID)).
		Where("status = ?", model.MaterialStatusActive)

	if len(f.CategoryIDs) > 0 {
		q = q.Where("category_id IN ?", f.CategoryIDs)
	}
	if f.MediaType != "" {
		q = q.Where("media_type = ?", f.MediaType)
	}
	if kw := strings.TrimSpace(f.Keyword); kw != "" {
		like := "%" + strings.ToLower(kw) + "%"
		q = q.Where("LOWER(title) LIKE ? OR LOWER(remark) LIKE ? OR LOWER(product_sn) LIKE ? OR LOWER(file_name) LIKE ?", like, like, like, like)
	}
	if sn := strings.TrimSpace(f.ProductSn); sn != "" {
		q = q.Where("LOWER(product_sn) LIKE ?", "%"+strings.ToLower(sn)+"%")
	}
	switch f.HasProduct {
	case "yes":
		q = q.Where("(product_id IS NOT NULL AND product_id > 0) OR (product_sn <> '' AND product_sn IS NOT NULL)")
	case "no":
		q = q.Where("(product_id IS NULL OR product_id = 0) AND (product_sn = '' OR product_sn IS NULL)")
	}
	if tag := strings.TrimSpace(f.Tag); tag != "" {
		q = q.Where("tags_json LIKE ?", "%\""+tag+"\"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, pageSize := f.Page, f.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 24
	}
	if pageSize > 200 {
		pageSize = 200
	}

	var list []model.Material
	err := q.Order("sort ASC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *MaterialRepo) Stats(tenantID uint64) (total, images, videos, weekNew int64, err error) {
	base := r.db.Model(&model.Material{}).Scopes(scopeTenant(tenantID)).Where("status = ?", model.MaterialStatusActive)
	if err = base.Count(&total).Error; err != nil {
		return
	}
	if err = r.db.Model(&model.Material{}).Scopes(scopeTenant(tenantID)).
		Where("status = ? AND media_type = ?", model.MaterialStatusActive, model.MediaTypeImage).
		Count(&images).Error; err != nil {
		return
	}
	if err = r.db.Model(&model.Material{}).Scopes(scopeTenant(tenantID)).
		Where("status = ? AND media_type = ?", model.MaterialStatusActive, model.MediaTypeVideo).
		Count(&videos).Error; err != nil {
		return
	}
	weekAgo := time.Now().AddDate(0, 0, -7)
	err = r.db.Model(&model.Material{}).Scopes(scopeTenant(tenantID)).
		Where("status = ? AND created_at >= ?", model.MaterialStatusActive, weekAgo).
		Count(&weekNew).Error
	return
}

type categoryCountRow struct {
	CategoryID uint64
	Count      int64
}

func (r *MaterialRepo) CountGroupByCategory(tenantID uint64) ([]categoryCountRow, error) {
	var rows []categoryCountRow
	err := r.db.Model(&model.Material{}).
		Select("category_id as category_id, count(*) as count").
		Scopes(scopeTenant(tenantID)).
		Where("status = ?", model.MaterialStatusActive).
		Group("category_id").
		Scan(&rows).Error
	return rows, err
}
