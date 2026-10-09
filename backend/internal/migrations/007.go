package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration007 = &gormigrate.Migration{
	ID: "007_replace_jwt_with_database_sessions",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			DELETE FROM user_tokens;
			ALTER TABLE user_tokens
				ALTER COLUMN token TYPE VARCHAR(64),
				ALTER COLUMN token SET NOT NULL,
				ADD COLUMN user_agent VARCHAR(512) NOT NULL DEFAULT '';
			CREATE UNIQUE INDEX idx_user_tokens_token ON user_tokens(token);
		`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			DROP INDEX IF EXISTS idx_user_tokens_token;
			ALTER TABLE user_tokens
				DROP COLUMN IF EXISTS user_agent,
				ALTER COLUMN token TYPE TEXT;
		`).Error
	},
}
