package testutils

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB creates an in-memory SQLite DB, runs migrations for given models,
// and returns the DB instance. Used for testing.
func SetupTestDB(t interface{}, models ...interface{}) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test DB: " + err.Error())
	}

	if err := db.AutoMigrate(models...); err != nil {
		panic("failed to migrate test DB: " + err.Error())
	}
	return db
}
