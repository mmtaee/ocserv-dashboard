package database

import (
	"fmt"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"time"
)

var (
	SQLiteDB   *gorm.DB
	PostgresDB *gorm.DB
)

// ==========================
// INIT BOTH DATABASES
// ==========================

func ConnectSQLite(debug bool) {
	db, err := connectSQLite(debug)
	if err != nil {
		logger.Fatal("sqlite connection error: %v", err)
	}

	finalizeSQLite(db, debug)
}

func ConnectPostgres() {
	cfg := config.Get()
	db, err := connectPostgres(cfg.DB)
	if err != nil {
		logger.Fatal("postgres connection error: %v", err)
	}

	finalizePostgres(db, cfg.Debug)
}

// ==========================
// SQLITE
// ==========================

func connectSQLite(debug bool) (*gorm.DB, error) {
	dbPath := "./db"

	if debug {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dbPath = filepath.Join(home, "ocserv_db")
	}

	if err := os.MkdirAll(dbPath, os.ModePerm); err != nil {
		return nil, err
	}

	dbFile := filepath.Join(dbPath, "ocserv.db")

	logger.Info("Connecting SQLite [%s]", dbFile)

	dsn := dbFile + "?_busy_timeout=5000"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.Exec("PRAGMA journal_mode=DELETE;")
	db.Exec("PRAGMA synchronous=FULL;")
	db.Exec("PRAGMA foreign_keys=ON;")

	return db, nil
}

func finalizeSQLite(db *gorm.DB, debug bool) {
	sqlDB, _ := db.DB()

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if debug {
		db = db.Debug()
	}

	SQLiteDB = db

	logger.Info("SQLite connected successfully")
}

// ==========================
// POSTGRES
// ==========================
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

// ==========================
// GETTERS
// ==========================

func GetSQLite() *gorm.DB {
	if SQLiteDB == nil {
		logger.Fatal("SQLite not initialized")
	}
	return SQLiteDB
}

func GetPostgres() *gorm.DB {
	if PostgresDB == nil {
		logger.Fatal("Postgres not initialized")
	}
	return PostgresDB
}

// ==========================
// CLOSE ALL
// ==========================

func CloseSQLite() {
	if SQLiteDB != nil {
		if db, _ := SQLiteDB.DB(); db != nil {
			_ = db.Close()
		}
	}

	logger.Info("SQLite databases closed")
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
