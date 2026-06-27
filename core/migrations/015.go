package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/gorm"
)

var Migration015 = &gormigrate.Migration{
	ID: "015_create_administrator_tokens_table",

	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS administrator_tokens (
				id BIGSERIAL PRIMARY KEY,
				administrator_id BIGINT NOT NULL,
				token TEXT NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				expire_at TIMESTAMP NOT NULL,
				CONSTRAINT fk_administrator_tokens_administrator FOREIGN KEY (administrator_id) REFERENCES administrators(id) ON DELETE CASCADE ON UPDATE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_administrator_tokens_administrator_id ON administrator_tokens(administrator_id);
			CREATE INDEX IF NOT EXISTS idx_administrator_tokens_token ON administrator_tokens(token);
		`).Error; err != nil {
			return err
		}

		logger.Info("migration 015 (create administrator tokens table) complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			DROP INDEX IF EXISTS idx_administrator_tokens_administrator_id;
			DROP INDEX IF EXISTS idx_administrator_tokens_token;
			DROP TABLE IF EXISTS administrator_tokens;
		`).Error; err != nil {
			return err
		}

		return nil
	},
}
