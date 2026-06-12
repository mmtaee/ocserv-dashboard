package bootstrap

import (
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/migrations"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
)

var coreMigrations = []*gormigrate.Migration{
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
	migrations.Migration012,
	migrations.Migration013,
}

func RunMigrations() {
	db := database.GetConnection()
	m := gormigrate.New(db, gormigrate.DefaultOptions, coreMigrations)

	if err := m.Migrate(); err != nil {
		log.Fatalf("Could not migrate: %v", err)
	}
	log.Println("Migration run successfully")
}
