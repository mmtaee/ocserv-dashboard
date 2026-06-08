package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	"gorm.io/gorm"
)

var Migration011 = &gormigrate.Migration{
	ID: "011_add_ocserv_user_usage_reset_at",

	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			ALTER TABLE ocserv_users
			ADD COLUMN IF NOT EXISTS usage_reset_at TIMESTAMPTZ NULL;
		`).Error; err != nil {
			return err
		}

		logger.Info("migration 011 (ocserv_users.usage_reset_at) complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE ocserv_users
			DROP COLUMN IF EXISTS usage_reset_at;
		`).Error
	},
}
