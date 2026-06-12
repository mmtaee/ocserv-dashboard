package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/gorm"
)

var Migration007 = &gormigrate.Migration{
	ID: "007_widen_ocserv_users_username_password",

	// Do not widen in 001: shipped migrations must stay immutable; new installs get VARCHAR(16) from 001 then this step.

	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			ALTER TABLE ocserv_users
				ALTER COLUMN username TYPE VARCHAR(255),
				ALTER COLUMN password TYPE VARCHAR(255);
		`).Error; err != nil {
			return err
		}

		logger.Info("migration 007 (widen ocserv_users username/password) complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE ocserv_users
				ALTER COLUMN username TYPE VARCHAR(16),
				ALTER COLUMN password TYPE VARCHAR(16);
		`).Error
	},
}
