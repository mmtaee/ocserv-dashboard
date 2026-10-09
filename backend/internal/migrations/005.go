package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration005 = &gormigrate.Migration{
	ID: "005_add_ocserv_user_expiry_modes",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE ocserv_users
				ALTER COLUMN expire_at TYPE TIMESTAMPTZ
				USING expire_at::timestamp AT TIME ZONE 'UTC';

			ALTER TABLE ocserv_users
				ADD COLUMN expiry_mode VARCHAR(32),
				ADD COLUMN expire_days_after_first_connection INTEGER NULL,
				ADD COLUMN first_connected_at TIMESTAMPTZ NULL;

			UPDATE ocserv_users
			SET expiry_mode = CASE
				WHEN expire_at IS NULL THEN 'unlimited'
				ELSE 'fixed'
			END;

			ALTER TABLE ocserv_users
				ALTER COLUMN expiry_mode SET NOT NULL,
				ALTER COLUMN expiry_mode SET DEFAULT 'unlimited',
				ADD CONSTRAINT chk_ocserv_users_expiry_configuration CHECK (
					(expiry_mode = 'unlimited'
						AND expire_at IS NULL
						AND expire_days_after_first_connection IS NULL
						AND first_connected_at IS NULL)
					OR (expiry_mode = 'fixed'
						AND expire_at IS NOT NULL
						AND expire_days_after_first_connection IS NULL
						AND first_connected_at IS NULL)
					OR (expiry_mode = 'first_connection'
						AND expire_days_after_first_connection > 0
						AND ((first_connected_at IS NULL AND expire_at IS NULL)
							OR (first_connected_at IS NOT NULL AND expire_at IS NOT NULL)))
				);
		`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE ocserv_users DROP CONSTRAINT IF EXISTS chk_ocserv_users_expiry_configuration;
			ALTER TABLE ocserv_users
				DROP COLUMN IF EXISTS first_connected_at,
				DROP COLUMN IF EXISTS expire_days_after_first_connection,
				DROP COLUMN IF EXISTS expiry_mode;
			ALTER TABLE ocserv_users
				ALTER COLUMN expire_at TYPE DATE USING (expire_at AT TIME ZONE 'UTC')::date;
		`).Error
	},
}
