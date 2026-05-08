package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	"gorm.io/gorm"
)

var Migration003 = &gormigrate.Migration{
	ID: "003_widen_ocserv_users_username_password",

	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			ALTER TABLE ocserv_users
				ALTER COLUMN username TYPE VARCHAR(255),
				ALTER COLUMN password TYPE VARCHAR(255);
		`).Error; err != nil {
			return err
		}

		logger.Info("migration 003 (widen ocserv_users username/password) complete successfully")
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
