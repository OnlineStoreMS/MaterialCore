package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"materialcore/internal/config"
	"materialcore/internal/model"
	"materialcore/internal/seed"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case "postgres":
		dialector = postgres.Open(cfg.PostgresDSN)
	case "sqlite":
		if err := os.MkdirAll(filepath.Dir(cfg.SQLitePath), 0o755); err != nil {
			return nil, err
		}
		dialector = sqlite.Open(cfg.SQLitePath)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
	return gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.MaterialCategory{},
		&model.Material{},
	); err != nil {
		return err
	}
	if err := seed.EnsureDefaultCategories(db, 1); err != nil {
		log.Printf("seed default categories: %v", err)
	}
	if db.Dialector.Name() == "postgres" {
		return db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_materials_tenant_media ON materials (tenant_id, media_type);
			CREATE INDEX IF NOT EXISTS idx_materials_tenant_product_sn ON materials (tenant_id, product_sn);
		`).Error
	}
	return nil
}
