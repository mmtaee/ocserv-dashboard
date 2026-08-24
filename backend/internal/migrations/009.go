package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration009 = &gormigrate.Migration{
	ID: "009_create_agent_tokens",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			CREATE TABLE agent_tokens (
				id BIGSERIAL PRIMARY KEY,
				token VARCHAR(512) NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE UNIQUE INDEX idx_agent_tokens_singleton ON agent_tokens ((TRUE));
		`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`DROP TABLE IF EXISTS agent_tokens;`).Error
	},
}
