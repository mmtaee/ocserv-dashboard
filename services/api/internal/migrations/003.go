package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"gorm.io/gorm"
)

var Migration003 = &gormigrate.Migration{
	ID: "003_create_permissions",
	Migrate: func(tx *gorm.DB) error {
		err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS permissions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				service VARCHAR(64) NOT NULL,
				action VARCHAR(1) NOT NULL,
				created_at DATETIME,
				updated_at DATETIME,
				FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
			);
		`).Error
		if err != nil {
			return err
		}
		logger.Info("migration 0003 complete successfully")
		return nil
	},
}
