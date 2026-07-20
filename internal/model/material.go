package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	MediaTypeImage = "image"
	MediaTypeVideo = "video"

	CategoryCodeUncategorized = "uncategorized"

	MaterialStatusActive  = 1
	MaterialStatusDeleted = 0
)

// MaterialCategory 素材分类（树形，可自定义）
type MaterialCategory struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	TenantID  uint64         `gorm:"not null;uniqueIndex:uk_cat_tenant_parent_name,priority:1;index" json:"tenantId"`
	ParentID  uint64         `gorm:"not null;default:0;uniqueIndex:uk_cat_tenant_parent_name,priority:2" json:"parentId"`
	Name      string         `gorm:"size:128;not null;uniqueIndex:uk_cat_tenant_parent_name,priority:3" json:"name"`
	Code      string         `gorm:"size:64;index" json:"code"` // 系统保留如 uncategorized
	Sort      int            `gorm:"not null;default:0" json:"sort"`
	Enabled   int            `gorm:"not null;default:1" json:"enabled"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MaterialCategory) TableName() string { return "material_categories" }

// Material 素材条目
type Material struct {
	ID         uint64         `gorm:"primaryKey" json:"id"`
	TenantID   uint64         `gorm:"not null;index:idx_mat_tenant_cat,priority:1;index:idx_mat_tenant_created,priority:1" json:"tenantId"`
	CategoryID uint64         `gorm:"not null;default:0;index:idx_mat_tenant_cat,priority:2" json:"categoryId"`
	Title      string         `gorm:"size:256;not null" json:"title"`
	MediaType  string         `gorm:"size:16;not null;index" json:"mediaType"` // image | video
	URL        string         `gorm:"size:1024;not null" json:"url"`
	CoverURL   string         `gorm:"size:1024" json:"coverUrl"`
	FileName   string         `gorm:"size:256" json:"fileName"`
	Mime       string         `gorm:"size:128" json:"mime"`
	SizeBytes  int64          `gorm:"not null;default:0" json:"sizeBytes"`
	DurationMs int64          `gorm:"not null;default:0" json:"durationMs"`
	Width      int            `gorm:"not null;default:0" json:"width"`
	Height     int            `gorm:"not null;default:0" json:"height"`
	TagsJSON   string         `gorm:"type:text" json:"tagsJson"` // JSON string[]
	Remark     string         `gorm:"size:1024" json:"remark"`
	ProductID  *uint64        `json:"productId"`
	ProductSn  string         `gorm:"size:128;index" json:"productSn"`
	Sort       int            `gorm:"not null;default:0" json:"sort"`
	Status     int            `gorm:"not null;default:1;index" json:"status"`
	CreatedBy  uint64         `gorm:"not null;default:0" json:"createdBy"`
	CreatedAt  time.Time      `gorm:"index:idx_mat_tenant_created,priority:2" json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Material) TableName() string { return "materials" }
