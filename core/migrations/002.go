package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/gorm"
)

var Migration002 = &gormigrate.Migration{
	ID: "002_create_ocserv_user_session_logs",

	Migrate: func(tx *gorm.DB) error {

		// =========================
		// SESSION LOGS TABLE
		// =========================
		if err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS ocserv_user_session_logs (
				id BIGSERIAL PRIMARY KEY,
				username VARCHAR(64),
				ip VARCHAR(45),
				event VARCHAR(64),
				message TEXT,
				created_at TIMESTAMP DEFAULT NOW()
			);
		`).Error; err != nil {
			return err
		}

		// =========================
		// INDEXES
		// =========================
		if err := tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_ocserv_logs_username
			ON ocserv_user_session_logs(username);
		`).Error; err != nil {
			return err
		}

		logger.Info("migration 002 (Postgres) complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			DROP TABLE IF EXISTS ocserv_user_session_logs;
		`).Error
	},
}
