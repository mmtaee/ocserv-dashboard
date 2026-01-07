package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration002 = &gormigrate.Migration{
	ID: "002_add_roles_and_admin",
	Migrate: func(tx *gorm.DB) error {
		// 1️⃣ Add Role column if it doesn't exist
		if err := tx.Exec(`
			ALTER TABLE users ADD COLUMN role VARCHAR(16) NOT NULL DEFAULT 'staff';
		`).Error; err != nil {
			return err
		}

		// 2️⃣ Add AdminID column if it doesn't exist
		if err := tx.Exec(`
			ALTER TABLE users ADD COLUMN admin_id INTEGER;
		`).Error; err != nil {
			return err
		}

		// 3️⃣ Create index for role
		if err := tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
		`).Error; err != nil {
			return err
		}

		// 4️⃣ Backfill Role from is_admin
		// is_admin = true  → superadmin
		// is_admin = false → admin
		if err := tx.Exec(`
			UPDATE users
			SET role = CASE
				WHEN is_admin = 1 THEN 'superadmin'
				ELSE 'admin'
			END
		`).Error; err != nil {
			return err
		}

		return nil
	},
}
