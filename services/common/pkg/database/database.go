package database

import (
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

var DB *gorm.DB

func Connect() {
	conf := config.Get()

	dbPath := "./db"
	if conf.Debug {
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Fatal("error getting user home directory: %v", err)
		}
		dbPath = filepath.Join(home, "ocserv_db")
	}

	err := os.MkdirAll(dbPath, os.ModePerm)
	if err != nil {
		logger.Fatal("error creating db path: %v", err)
	}

	dbPath = filepath.Join(dbPath, "ocserv.db")

	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		logger.Fatal("error getting abs path: %v", err)
	}

	logger.Info("Connecting to database [%s] ...", absPath)
	db, err := gorm.Open(sqlite.Open(absPath), &gorm.Config{})
	if err != nil {
		logger.Fatal("error connecting to database: %v", err)
	}
	if conf.Debug {
		db = db.Debug()
	}
	DB = db
	logger.Info("Connected to database [%s] successfully ...", absPath)
}

func GetConnection() *gorm.DB {
	return DB
}

func CloseConnection() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		err := sqlDB.Close()
		if err != nil {
			logger.Fatal("error closing database connection: %v", err)
		}
		logger.Info("Closed database connection")
	}
}
