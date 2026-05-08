package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	"gorm.io/gorm"
)

var Migration004 = &gormigrate.Migration{
	ID: "004_create_telegram_tables",

	Migrate: func(tx *gorm.DB) error {
		// =========================
		// TELEGRAM SETTINGS (singleton)
		// =========================
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS telegram_settings (
				id BIGSERIAL PRIMARY KEY,
				enabled BOOLEAN DEFAULT FALSE,
				bot_token VARCHAR(255) DEFAULT '',
				bot_username VARCHAR(64) DEFAULT '',
				admin_chat_id BIGINT DEFAULT 0,
				low_quota_threshold_mb INTEGER DEFAULT 200,
				default_language VARCHAR(8) DEFAULT 'en',
				ocserv_host VARCHAR(255) DEFAULT '',
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			INSERT INTO telegram_settings (id, enabled, default_language, low_quota_threshold_mb)
			VALUES (1, FALSE, 'en', 200)
			ON CONFLICT (id) DO NOTHING;
		`).Error; err != nil {
			return err
		}

		// =========================
		// TELEGRAM ACCOUNTS
		// =========================
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS telegram_accounts (
				id BIGSERIAL PRIMARY KEY,
				chat_id BIGINT NOT NULL,
				telegram_username VARCHAR(64) DEFAULT '',
				language VARCHAR(8) DEFAULT 'en',
				ocserv_user_id BIGINT NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				last_low_quota_notified_at TIMESTAMP NULL,
				CONSTRAINT fk_telegram_accounts_ocserv_user
					FOREIGN KEY(ocserv_user_id)
					REFERENCES ocserv_users(id)
					ON DELETE CASCADE
			);
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_telegram_accounts_chat_id
			ON telegram_accounts(chat_id);
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS uniq_telegram_accounts_chat_user
			ON telegram_accounts(chat_id, ocserv_user_id);
		`).Error; err != nil {
			return err
		}

		// =========================
		// TELEGRAM PACKAGES
		// =========================
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS telegram_packages (
				id BIGSERIAL PRIMARY KEY,
				title VARCHAR(128) NOT NULL,
				days INTEGER NOT NULL,
				traffic_size_gb INTEGER NOT NULL,
				traffic_type VARCHAR(32) NOT NULL DEFAULT 'TotallyTransmit',
				price_text VARCHAR(64) DEFAULT '',
				is_active BOOLEAN DEFAULT TRUE,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_telegram_packages_active
			ON telegram_packages(is_active);
		`).Error; err != nil {
			return err
		}

		// =========================
		// TELEGRAM REQUESTS
		// =========================
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS telegram_requests (
				id BIGSERIAL PRIMARY KEY,
				chat_id BIGINT NOT NULL,
				telegram_username VARCHAR(64) DEFAULT '',
				type VARCHAR(16) NOT NULL,
				package_id BIGINT NULL,
				target_ocserv_id BIGINT NULL,
				desired_username VARCHAR(64) DEFAULT '',
				status VARCHAR(32) NOT NULL DEFAULT 'pending',
				receipt_file_path VARCHAR(255) DEFAULT '',
				user_message TEXT,
				admin_note TEXT,
				delivered_at TIMESTAMP NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_telegram_requests_status
			ON telegram_requests(status);
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_telegram_requests_chat_id
			ON telegram_requests(chat_id);
		`).Error; err != nil {
			return err
		}

		logger.Info("migration 004 (create telegram tables) complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		statements := []string{
			`DROP TABLE IF EXISTS telegram_requests;`,
			`DROP TABLE IF EXISTS telegram_packages;`,
			`DROP TABLE IF EXISTS telegram_accounts;`,
			`DROP TABLE IF EXISTS telegram_settings;`,
		}
		for _, stmt := range statements {
			if err := tx.Exec(stmt).Error; err != nil {
				return err
			}
		}
		return nil
	},
}
