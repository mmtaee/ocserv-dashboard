package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	"gorm.io/gorm"
)

var Migration009 = &gormigrate.Migration{
	ID: "009_store_ocserv_traffic_size_in_bytes",

	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec(`
			UPDATE ocserv_users
			SET traffic_size = traffic_size * 1073741824
			WHERE traffic_type <> 'Free'
			  AND traffic_size > 0
			  AND traffic_size < 1048576;
		`).Error; err != nil {
			return err
		}

		logger.Info("migration 009 converted ocserv traffic_size from GiB units to bytes successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			UPDATE ocserv_users
			SET traffic_size = traffic_size / 1073741824
			WHERE traffic_type <> 'Free'
			  AND traffic_size >= 1048576;
		`).Error
	},
}
