package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migration010 = &gormigrate.Migration{
	ID: "010_add_system_first_init",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE systems
				ADD COLUMN first_init BOOLEAN NOT NULL DEFAULT FALSE;
		`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec(`
			ALTER TABLE systems
				DROP COLUMN IF EXISTS first_init;
		`).Error
	},
}
