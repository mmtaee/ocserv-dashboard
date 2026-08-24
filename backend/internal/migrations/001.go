package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration001 = &gormigrate.Migration{
	ID: "001_create_dashboard_tables",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			CREATE TABLE users (
				id BIGSERIAL PRIMARY KEY,
				username VARCHAR(16) NOT NULL UNIQUE,
				password VARCHAR(64) NOT NULL,
				is_admin BOOLEAN NOT NULL DEFAULT FALSE,
				salt VARCHAR(8) NOT NULL,
				last_login TIMESTAMPTZ NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
			);

			CREATE INDEX idx_users_last_login ON users(last_login);
			CREATE INDEX idx_users_is_admin ON users(is_admin);
			CREATE INDEX idx_users_created_at ON users(created_at);

			CREATE TABLE user_tokens (
				id BIGSERIAL PRIMARY KEY,
				user_id BIGINT NOT NULL,
				token TEXT,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				expire_at TIMESTAMPTZ NOT NULL,
				CONSTRAINT fk_user_tokens_user
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);

			CREATE INDEX idx_user_tokens_user_id ON user_tokens(user_id);
			CREATE INDEX idx_user_tokens_expire_at ON user_tokens(expire_at);

			CREATE TABLE systems (
				id BIGSERIAL PRIMARY KEY,
				google_captcha_secret_key TEXT,
				google_captcha_site_key TEXT,
				auto_delete_inactive_users BOOLEAN NOT NULL DEFAULT FALSE,
				keep_inactive_user_days INTEGER NOT NULL DEFAULT 30,
				client_profile_server_address VARCHAR(255) NOT NULL DEFAULT '',
				client_profile_server_port INTEGER NOT NULL DEFAULT 443,
				client_profile_connection_name VARCHAR(64) NOT NULL DEFAULT ''
			);
		`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			DROP TABLE IF EXISTS user_tokens;
			DROP TABLE IF EXISTS systems;
			DROP TABLE IF EXISTS users;
		`).Error
	},
}
