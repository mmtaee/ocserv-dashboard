package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/gorm"
)

var Migration016 = &gormigrate.Migration{
	ID: "016_create_secondary_servers_table",

	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS secondary_servers (
				id BIGSERIAL PRIMARY KEY,
				title VARCHAR(255) NOT NULL,
				ip VARCHAR(255) NOT NULL,
				port INT NOT NULL,
				token TEXT NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`).Error; err != nil {
			return err
		}

		logger.Info("migration 016 (create secondary servers table) complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			DROP TABLE IF EXISTS secondary_servers;
		`).Error; err != nil {
			return err
		}

		return nil
	},
}
