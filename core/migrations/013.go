package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/gorm"
)

var Migration013 = &gormigrate.Migration{
	ID: "013_add_client_profile_fields_to_systems",

	Migrate: func(tx *gorm.DB) error {
		// Add missing columns
		if err := tx.Exec(`
			ALTER TABLE systems
			ADD COLUMN IF NOT EXISTS client_profile_server_address VARCHAR(255) DEFAULT '',
			ADD COLUMN IF NOT EXISTS client_profile_server_port INTEGER DEFAULT 443,
			ADD COLUMN IF NOT EXISTS client_profile_connection_name VARCHAR(64) DEFAULT '';
		`).Error; err != nil {
			return err
		}

		// Check if there are any rows and insert if none exist
		var count int64
		if err := tx.Raw(`SELECT COUNT(*) FROM systems;`).Scan(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			if err := tx.Exec(`
				INSERT INTO systems (id, google_captcha_secret_key, google_captcha_site_key, auto_delete_inactive_users, keep_inactive_user_days, client_profile_server_address, client_profile_server_port, client_profile_connection_name)
				VALUES (1, '', '', false, 30, '', 443, '');
			`).Error; err != nil {
				return err
			}
		}

		logger.Info("migration 013 (add client profile fields to systems) complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			ALTER TABLE systems
			DROP COLUMN IF EXISTS client_profile_server_address,
			DROP COLUMN IF EXISTS client_profile_server_port,
			DROP COLUMN IF EXISTS client_profile_connection_name;
		`).Error; err != nil {
			return err
		}

		return nil
	},
}
