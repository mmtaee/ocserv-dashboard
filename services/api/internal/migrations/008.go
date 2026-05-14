// Package migrations holds gormigrate database migrations.
//
// Migration 008 adds telegram_requests.awaiting_payment_message_id (Telegram message id),
// not user password column length.
package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	"gorm.io/gorm"
)

var Migration008 = &gormigrate.Migration{
	ID: "008_telegram_request_awaiting_payment_message_id",

	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			ALTER TABLE telegram_requests
			ADD COLUMN IF NOT EXISTS awaiting_payment_message_id BIGINT NULL;
		`).Error; err != nil {
			return err
		}
		logger.Info("migration 008 (telegram_requests.awaiting_payment_message_id) complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE telegram_requests
			DROP COLUMN IF EXISTS awaiting_payment_message_id;
		`).Error
	},
}
