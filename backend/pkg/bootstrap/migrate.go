package bootstrap

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/migrations"
	"gorm.io/gorm"
)

var commonMigrations = []*gormigrate.Migration{
	migrations.Migration001,
	migrations.Migration002,
	migrations.Migration003,
	migrations.Migration004,
	migrations.Migration005,
	migrations.Migration006,
	migrations.Migration007,
}

func MigrationsFor(agentNode bool) []*gormigrate.Migration {
	selected := append([]*gormigrate.Migration(nil), commonMigrations...)
	if agentNode {
		return append(selected, migrations.Migration009)
	}
	return append(selected, migrations.Migration008)
}

func Migrate(db *gorm.DB) error {
	if err := gormigrate.New(db, gormigrate.DefaultOptions, MigrationsFor(config.Get().AgentNode)).Migrate(); err != nil {
		return fmt.Errorf("run database migrations: %w", err)
	}
	return nil
}
