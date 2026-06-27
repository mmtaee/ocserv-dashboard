package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/gorm"
)

var Migration005 = &gormigrate.Migration{
	ID: "005_telegram_settings_card_fields",

	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			ALTER TABLE telegram_settings
				ADD COLUMN IF NOT EXISTS card_number VARCHAR(64) DEFAULT '',
				ADD COLUMN IF NOT EXISTS card_holder VARCHAR(128) DEFAULT '';
		`).Error; err != nil {
			return err
		}
		logger.Info("migration 005 (telegram_settings card fields) complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE telegram_settings
				DROP COLUMN IF EXISTS card_number,
				DROP COLUMN IF EXISTS card_holder;
		`).Error
	},
}
