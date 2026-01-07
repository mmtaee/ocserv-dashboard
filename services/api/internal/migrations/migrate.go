package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"gorm.io/gorm"
)

var Migrations = []*gormigrate.Migration{
	Migration001,
	Migration002,
	Migration003,
}

func Migrate() {
	logger.Info("Starting database migrations...")
	db := database.GetConnection()

	m := gormigrate.New(db, gormigrate.DefaultOptions, Migrations)
	m.InitSchema(func(tx *gorm.DB) error {
		return nil
	})

	if err := m.Migrate(); err != nil {
		logger.Fatal("Failed to run migrations: %v", err)
	}

	logger.Info("Database migrations complete")
}
