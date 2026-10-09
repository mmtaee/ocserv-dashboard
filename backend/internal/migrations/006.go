package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration006 = &gormigrate.Migration{
	ID: "006_rename_ocserv_user_running_usage",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE ocserv_users RENAME COLUMN rx TO running_rx;
			ALTER TABLE ocserv_users RENAME COLUMN tx TO running_tx;
		`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE ocserv_users RENAME COLUMN running_rx TO rx;
			ALTER TABLE ocserv_users RENAME COLUMN running_tx TO tx;
		`).Error
	},
}
