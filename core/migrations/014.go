package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/gorm"
)

var Migration014 = &gormigrate.Migration{
	ID: "014_add_admin_suspension_fields",

	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			ALTER TABLE administrators
			ADD COLUMN IF NOT EXISTS is_suspended BOOLEAN DEFAULT false,
			ADD COLUMN IF NOT EXISTS suspended_at TIMESTAMP,
			ADD COLUMN IF NOT EXISTS suspended_reason TEXT;
		`).Error; err != nil {
			return err
		}

		logger.Info("migration 014 (add admin suspension fields) complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			ALTER TABLE administrators
			DROP COLUMN IF EXISTS is_suspended,
			DROP COLUMN IF EXISTS suspended_at,
			DROP COLUMN IF EXISTS suspended_reason;
		`).Error; err != nil {
			return err
		}

		return nil
	},
}
