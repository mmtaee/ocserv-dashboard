package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/gorm"
)

var Migration006 = &gormigrate.Migration{
	ID: "006_telegram_settings_support_username",

	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			ALTER TABLE telegram_settings
				ADD COLUMN IF NOT EXISTS support_username VARCHAR(64) DEFAULT '';
		`).Error; err != nil {
			return err
		}
		logger.Info("migration 006 (telegram_settings support_username) complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE telegram_settings
				DROP COLUMN IF EXISTS support_username;
		`).Error
	},
}
