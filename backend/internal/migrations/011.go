package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration011 = &gormigrate.Migration{
	ID: "011_add_ocserv_agent_port",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE ocserv_agents
				ADD COLUMN port INTEGER NOT NULL DEFAULT 8080
				CHECK (port BETWEEN 1 AND 65535);
		`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE ocserv_agents
				DROP COLUMN IF EXISTS port;
		`).Error
	},
}
