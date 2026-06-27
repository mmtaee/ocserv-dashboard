package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/gorm"
)

var Migration017 = &gormigrate.Migration{
	ID: "017_remove_expire_at_from_administrator_tokens",

	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			ALTER TABLE administrator_tokens DROP COLUMN IF EXISTS expire_at;
		`).Error; err != nil {
			return err
		}

		logger.Info("migration 017 (remove expire_at from administrator_tokens) complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			ALTER TABLE administrator_tokens ADD COLUMN IF NOT EXISTS expire_at TIMESTAMP NOT NULL DEFAULT NOW();
		`).Error; err != nil {
			return err
		}

		return nil
	},
}
