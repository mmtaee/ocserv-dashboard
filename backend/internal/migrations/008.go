package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration008 = &gormigrate.Migration{
	ID: "008_create_ocserv_agents",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			CREATE TABLE ocserv_agents (
				id BIGSERIAL PRIMARY KEY,
				name VARCHAR(128) NOT NULL,
				address_type VARCHAR(16) NOT NULL CHECK (address_type IN ('ip', 'domain')),
				address VARCHAR(255) NOT NULL UNIQUE,
				token VARCHAR(512) NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`DROP TABLE IF EXISTS ocserv_agents;`).Error
	},
}
