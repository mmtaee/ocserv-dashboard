package database

import (
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
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
			log.Fatal(err)
		}
		dbPath = filepath.Join(home, "ocserv_db")
	}

	err := os.MkdirAll(dbPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	dbPath = filepath.Join(dbPath, "ocserv.db")

	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connecting to database %s ...", absPath)
	db, err := gorm.Open(sqlite.Open(absPath), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}
	if conf.Debug {
		db = db.Debug()
	}
	DB = db
	log.Println("Connected to database")
}

func GetConnection() *gorm.DB {
	return DB
}

func CloseConnection() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		err := sqlDB.Close()
		if err != nil {
			log.Fatal("failed to close database")
		}
		log.Println("database closed successfully")
	}
}
