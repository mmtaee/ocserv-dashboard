package database

import (
	"fmt"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

var (
	PostgresDB *gorm.DB
)

func ConnectPostgres() {
	cfg := config.Get()
	db, err := connectPostgres(cfg.DB)
	if err != nil {
		logger.Fatal("postgres connection error: %v", err)
	}

	finalizePostgres(db, cfg.Debug)
}

func connectPostgres(cfg config.PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode,
	)

	logger.Info("Connecting Postgres [%s:%s/%s]", cfg.Host, cfg.Port, cfg.DBName)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func finalizePostgres(db *gorm.DB, debug bool) {
	sqlDB, _ := db.DB()

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if debug {
		db = db.Debug()
	}

	PostgresDB = db

	logger.Info("Postgres connected successfully")
}

func GetPostgres() *gorm.DB {
	if PostgresDB == nil {
		logger.Fatal("Postgres not initialized")
	}
	return PostgresDB
}

func ClosePostgres() {
	if PostgresDB != nil {
		if db, _ := PostgresDB.DB(); db != nil {
			_ = db.Close()
		}
	}

	logger.Info("Postgres databases closed")
}

func Connect() {
	ConnectPostgres()
}

func GetConnection() *gorm.DB {
	return PostgresDB
}

func Close() {
	ClosePostgres()
}
