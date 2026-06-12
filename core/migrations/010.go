package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/gorm"
)

var Migration010 = &gormigrate.Migration{
	ID: "010_add_client_profile_settings",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			ALTER TABLE systems
			ADD COLUMN IF NOT EXISTS client_profile_server_address VARCHAR(255) DEFAULT '',
			ADD COLUMN IF NOT EXISTS client_profile_server_port INTEGER DEFAULT 443,
			ADD COLUMN IF NOT EXISTS client_profile_connection_name VARCHAR(64) DEFAULT '';
		`).Error; err != nil {
			return err
		}

		logger.Info("migration 010 (client profile settings) complete successfully")
		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE systems
			DROP COLUMN IF EXISTS client_profile_server_address,
			DROP COLUMN IF EXISTS client_profile_server_port,
			DROP COLUMN IF EXISTS client_profile_connection_name;
		`).Error
	},
}
