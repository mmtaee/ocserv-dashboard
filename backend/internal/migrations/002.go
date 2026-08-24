package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration002 = &gormigrate.Migration{
	ID: "002_create_ocserv_account_tables",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			CREATE TABLE ocserv_groups (
				id BIGSERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL UNIQUE,
				config JSON
			);

			CREATE TABLE ocserv_users (
				id BIGSERIAL PRIMARY KEY,
				owner_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
				"group" VARCHAR(16) NOT NULL DEFAULT 'defaults',
				username VARCHAR(255) NOT NULL UNIQUE,
				password VARCHAR(255) NOT NULL,
				is_locked BOOLEAN NOT NULL DEFAULT FALSE,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				expire_at DATE NULL,
				deactivated_at DATE NULL,
				usage_reset_at TIMESTAMPTZ NULL,
				traffic_type VARCHAR(32) NOT NULL DEFAULT 'Free'
					CHECK (traffic_type IN ('Free', 'MonthlyTransmit', 'MonthlyReceive', 'MonthlyRxTx', 'TotallyTransmit', 'TotallyReceive', 'TotallyRxTx')),
				traffic_size BIGINT NOT NULL DEFAULT 0,
				rx BIGINT NOT NULL DEFAULT 0,
				tx BIGINT NOT NULL DEFAULT 0,
				description TEXT,
				config TEXT
			);

			CREATE INDEX idx_ocserv_users_owner_id ON ocserv_users(owner_id);
			CREATE INDEX idx_ocserv_users_group ON ocserv_users("group");
			CREATE INDEX idx_ocserv_users_expire_at ON ocserv_users(expire_at);
			CREATE INDEX idx_ocserv_users_deactivated_at ON ocserv_users(deactivated_at);
		`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			DROP TABLE IF EXISTS ocserv_users;
			DROP TABLE IF EXISTS ocserv_groups;
		`).Error
	},
}
