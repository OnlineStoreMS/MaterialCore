package seed

import (
	"log"

	"materialcore/internal/model"

	"gorm.io/gorm"
)

var defaultRootCategories = []struct {
	Name string
	Code string
	Sort int
}{
	{Name: "询盘快发", Code: "inquiry", Sort: 10},
	{Name: "预期效果", Code: "effect", Sort: 20},
	{Name: "宣传海报", Code: "poster", Sort: 30},
	{Name: "带货短视频", Code: "short_video", Sort: 40},
	{Name: "未分类", Code: model.CategoryCodeUncategorized, Sort: 999},
}

func normalizeTenantID(id uint64) uint64 {
	if id == 0 {
		return 1
	}
	return id
}

// EnsureDefaultCategories inserts default root categories for a tenant if missing.
func EnsureDefaultCategories(db *gorm.DB, tenantID uint64) error {
	tenantID = normalizeTenantID(tenantID)
	for _, def := range defaultRootCategories {
		var n int64
		q := db.Model(&model.MaterialCategory{}).Where("tenant_id = ?", tenantID)
		if def.Code != "" {
			q = q.Where("code = ?", def.Code)
		} else {
			q = q.Where("parent_id = 0 AND name = ?", def.Name)
		}
		if err := q.Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		row := model.MaterialCategory{
			TenantID: tenantID,
			ParentID: 0,
			Name:     def.Name,
			Code:     def.Code,
			Sort:     def.Sort,
			Enabled:  1,
		}
		if err := db.Create(&row).Error; err != nil {
			return err
		}
		log.Printf("seed category tenant=%d name=%s code=%s", tenantID, def.Name, def.Code)
	}
	return nil
}
