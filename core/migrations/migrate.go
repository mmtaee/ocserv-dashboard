package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
)

var Migrations = []*gormigrate.Migration{
	Migration001,
	Migration002,
	Migration003,
	Migration004,
	Migration005,
	Migration006,
	Migration007,
	Migration008, Migration009, Migration010,
	Migration011,
	Migration012,
	Migration013,
	Migration014,
}

func Migrate() {
	logger.Info("Starting database migrations...")

	config.Init()

	database.Connect()
	defer database.Close()

	db := database.GetConnection()

	m := gormigrate.New(db, gormigrate.DefaultOptions, Migrations)
	if err := m.Migrate(); err != nil {
		logger.Fatal("Failed to run migrations: %v", err)
	}

	logger.Info("Database migrations complete")
}
