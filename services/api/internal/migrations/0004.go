package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"gorm.io/gorm"
)

var Migration004 = &gormigrate.Migration{
	ID: "004_update_user_role",
	Migrate: func(tx *gorm.DB) error {
		// 1️⃣ Backfill role from is_admin
		if err := tx.Exec(`
			UPDATE users
			SET role = CASE
				WHEN is_admin = 1 THEN 'super-admin'
				ELSE 'admin'
			END
		`).Error; err != nil {
			return err
		}

		// 2️⃣ Drop the is_admin column
		if err := tx.Migrator().DropColumn("users", "is_admin"); err != nil {
			return err
		}

		logger.Info("migration 0004 complete successfully")
		return nil
	},

	Rollback: func(tx *gorm.DB) error {
		// In rollback, recreate the is_admin column with default false
		if err := tx.Migrator().AddColumn("users", "is_admin"); err != nil {
			return err
		}

		// Optionally, set is_admin based on role
		if err := tx.Exec(`
			UPDATE users
			SET is_admin = CASE
				WHEN role = 'super-admin' THEN 1
				ELSE 0
			END
		`).Error; err != nil {
			return err
		}

		return nil
	},
}
