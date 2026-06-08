package bootstrap

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/api/internal/migrations"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
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

func Migrate() {
	logger.Info("Starting database migrations...")

	config.Init(false, "", 0)

	database.Connect()
	defer database.Close()

	db := database.GetConnection()

	m := gormigrate.New(db, gormigrate.DefaultOptions, Migrations)
	if err := m.Migrate(); err != nil {
		logger.Fatal("Failed to run migrations: %v", err)
	}

	logger.Info("Database migrations complete")
}
