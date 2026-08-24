package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration003 = &gormigrate.Migration{
	ID: "003_create_ocserv_activity_tables",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			CREATE TABLE ocserv_user_traffic_statistics (
				id BIGSERIAL PRIMARY KEY,
				ocserv_user_id BIGINT NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				rx BIGINT NOT NULL DEFAULT 0,
				tx BIGINT NOT NULL DEFAULT 0,
				CONSTRAINT fk_traffic_statistics_ocserv_user
					FOREIGN KEY (ocserv_user_id) REFERENCES ocserv_users(id) ON DELETE CASCADE
			);

			CREATE INDEX idx_traffic_statistics_ocserv_user_id
				ON ocserv_user_traffic_statistics(ocserv_user_id);
			CREATE INDEX idx_traffic_statistics_created_at
				ON ocserv_user_traffic_statistics(created_at);

			CREATE TABLE ocserv_user_session_logs (
				id BIGSERIAL PRIMARY KEY,
				username VARCHAR(64) NOT NULL,
				ip VARCHAR(45),
				event VARCHAR(64) NOT NULL,
				message TEXT NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
			);

			CREATE INDEX idx_ocserv_session_logs_username
				ON ocserv_user_session_logs(username);
			CREATE INDEX idx_ocserv_session_logs_created_at
				ON ocserv_user_session_logs(created_at);
		`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			DROP TABLE IF EXISTS ocserv_user_session_logs;
			DROP TABLE IF EXISTS ocserv_user_traffic_statistics;
		`).Error
	},
}
