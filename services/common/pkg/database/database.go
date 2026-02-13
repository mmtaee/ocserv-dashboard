package database

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DB   *gorm.DB
	once sync.Once
)

// Connect initializes the database connection only once
func Connect() {
	once.Do(func() {
		conf := config.Get()

		dbPath := "./db"
		if conf.Debug {
			home, err := os.UserHomeDir()
			if err != nil {
				logger.Fatal("error getting user home directory: %v", err)
			}
			dbPath = filepath.Join(home, "ocserv_db")
		}

		if err := os.MkdirAll(dbPath, os.ModePerm); err != nil {
			logger.Fatal("error creating db path: %v", err)
		}

		dbPath = filepath.Join(dbPath, "ocserv.db")

		absPath, err := filepath.Abs(dbPath)
		if err != nil {
			logger.Fatal("error getting abs path: %v", err)
		}

		logger.Info("Connecting to database [%s] ...", absPath)

		dsn := absPath + "?_journal_mode=WAL&_busy_timeout=5000"

		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			logger.Fatal("error connecting to database: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			logger.Fatal("error getting sql db: %v", err)
		}

		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetConnMaxLifetime(time.Hour)

		if conf.Debug {
			db = db.Debug()
		}

		DB = db

		logger.Info("Connected to database [%s] successfully ...", absPath)
	})
}

func GetConnection() *gorm.DB {
	if DB == nil {
		logger.Fatal("database not initialized. Call Connect() first.")
	}
	return DB
}

func CloseConnection() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.Fatal("error closing database connection: %v", err)
			}
			logger.Info("Closed database connection")
		}
	}
}
