package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration001 = &gormigrate.Migration{
	ID: "001_create_initial_tables",
	Migrate: func(tx *gorm.DB) error {
		// --- Users table ---
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				uid CHAR(26) NOT NULL UNIQUE,
				owner VARCHAR(16) DEFAULT '',
				"group" VARCHAR(16) DEFAULT 'defaults',
				username VARCHAR(16) NOT NULL UNIQUE,
				password VARCHAR(16) NOT NULL,
				is_locked BOOLEAN DEFAULT 0,
				created_at DATETIME,
				updated_at DATETIME,
				expire_at DATE,
				deactivated_at DATE,
				traffic_type VARCHAR(32) NOT NULL DEFAULT '1',
				traffic_size INTEGER NOT NULL,
				rx INTEGER NOT NULL DEFAULT 0,
				tx INTEGER NOT NULL DEFAULT 0,
				description TEXT,
				config TEXT
			);
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_users_uid ON users(uid);`).Error; err != nil {
			return err
		}

		// --- UserTokens table ---
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS user_tokens (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				uid TEXT NOT NULL UNIQUE,
				token TEXT,
				created_at DATETIME,
				expire_at DATETIME,
				FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
			);
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_user_tokens_user_id ON user_tokens(user_id);`).Error; err != nil {
			return err
		}

		// --- System table ---
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS system (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				google_captcha_secret TEXT,
				google_captcha_site_key TEXT
			);
		`).Error; err != nil {
			return err
		}

		// --- OcservGroups table ---
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS ocserv_groups (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name VARCHAR(255) NOT NULL UNIQUE,
				owner VARCHAR(32) DEFAULT '',
				config TEXT
			);
		`).Error; err != nil {
			return err
		}

		// --- OcservUserTrafficStatistics table ---
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS ocserv_user_traffic_statistics (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				oc_user_id INTEGER NOT NULL,
				created_at DATETIME,
				rx INTEGER DEFAULT 0,
				tx INTEGER DEFAULT 0,
				FOREIGN KEY(oc_user_id) REFERENCES users(id) ON DELETE CASCADE
			);
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_traffic_statistics_oc_user_id ON ocserv_user_traffic_statistics(oc_user_id);`).Error; err != nil {
			return err
		}

		return nil
	},
}
