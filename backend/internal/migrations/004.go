package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration004 = &gormigrate.Migration{
	ID: "004_create_telegram_tables",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			CREATE TABLE telegram_settings (
				id BIGINT PRIMARY KEY,
				enabled BOOLEAN NOT NULL DEFAULT FALSE,
				bot_token VARCHAR(255) NOT NULL DEFAULT '',
				bot_username VARCHAR(64) NOT NULL DEFAULT '',
				admin_chat_id BIGINT NOT NULL DEFAULT 0,
				low_quota_threshold_mb INTEGER NOT NULL DEFAULT 200,
				default_language VARCHAR(8) NOT NULL DEFAULT 'en',
				ocserv_host VARCHAR(255) NOT NULL DEFAULT '',
				card_number VARCHAR(64) NOT NULL DEFAULT '',
				card_holder VARCHAR(128) NOT NULL DEFAULT '',
				support_username VARCHAR(64) NOT NULL DEFAULT '',
				updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
			);

			INSERT INTO telegram_settings (id) VALUES (1);

			CREATE TABLE telegram_accounts (
				id BIGSERIAL PRIMARY KEY,
				chat_id BIGINT NOT NULL,
				telegram_username VARCHAR(64) NOT NULL DEFAULT '',
				language VARCHAR(8) NOT NULL DEFAULT 'en',
				ocserv_user_id BIGINT NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				last_low_quota_notified_at TIMESTAMPTZ NULL,
				CONSTRAINT fk_telegram_accounts_ocserv_user
					FOREIGN KEY (ocserv_user_id) REFERENCES ocserv_users(id) ON DELETE CASCADE
			);

			CREATE INDEX idx_telegram_accounts_chat_id ON telegram_accounts(chat_id);
			CREATE INDEX idx_telegram_accounts_ocserv_user_id ON telegram_accounts(ocserv_user_id);
			CREATE UNIQUE INDEX uniq_telegram_accounts_chat_user
				ON telegram_accounts(chat_id, ocserv_user_id);

			CREATE TABLE telegram_packages (
				id BIGSERIAL PRIMARY KEY,
				title VARCHAR(128) NOT NULL,
				days INTEGER NOT NULL,
				traffic_size_gb INTEGER NOT NULL,
				traffic_type VARCHAR(32) NOT NULL DEFAULT 'TotallyTransmit',
				price_text VARCHAR(64) NOT NULL DEFAULT '',
				is_active BOOLEAN NOT NULL DEFAULT TRUE,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
			);

			CREATE INDEX idx_telegram_packages_active ON telegram_packages(is_active);

			CREATE TABLE telegram_requests (
				id BIGSERIAL PRIMARY KEY,
				chat_id BIGINT NOT NULL,
				telegram_username VARCHAR(64) NOT NULL DEFAULT '',
				type VARCHAR(16) NOT NULL,
				package_id BIGINT NULL,
				target_ocserv_user_id BIGINT NULL,
				desired_username VARCHAR(64) NOT NULL DEFAULT '',
				status VARCHAR(32) NOT NULL DEFAULT 'pending',
				receipt_file_path VARCHAR(255) NOT NULL DEFAULT '',
				user_message TEXT,
				admin_note TEXT,
				delivered_at TIMESTAMPTZ NULL,
				awaiting_payment_message_id BIGINT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				CONSTRAINT fk_telegram_requests_package
					FOREIGN KEY (package_id) REFERENCES telegram_packages(id) ON DELETE SET NULL,
				CONSTRAINT fk_telegram_requests_ocserv_user
					FOREIGN KEY (target_ocserv_user_id) REFERENCES ocserv_users(id) ON DELETE SET NULL
			);

			CREATE INDEX idx_telegram_requests_status ON telegram_requests(status);
			CREATE INDEX idx_telegram_requests_chat_id ON telegram_requests(chat_id);
			CREATE INDEX idx_telegram_requests_package_id ON telegram_requests(package_id);
			CREATE INDEX idx_telegram_requests_target_ocserv_user_id
				ON telegram_requests(target_ocserv_user_id);
		`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			DROP TABLE IF EXISTS telegram_requests;
			DROP TABLE IF EXISTS telegram_packages;
			DROP TABLE IF EXISTS telegram_accounts;
			DROP TABLE IF EXISTS telegram_settings;
		`).Error
	},
}
