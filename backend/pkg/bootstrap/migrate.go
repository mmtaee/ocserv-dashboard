package bootstrap

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/migrations"
	"gorm.io/gorm"
)

var Migrations = []*gormigrate.Migration{
	migrations.Migration001,
	migrations.Migration002,
	migrations.Migration003,
	migrations.Migration004,
	migrations.Migration005,
	migrations.Migration006,
	migrations.Migration007,
	migrations.Migration008,
	migrations.Migration009,
	migrations.Migration010,
	migrations.Migration011,
}

func Migrate(db *gorm.DB) error {
	if err := gormigrate.New(db, gormigrate.DefaultOptions, Migrations).Migrate(); err != nil {
		return fmt.Errorf("run database migrations: %w", err)
	}
	return nil
}
