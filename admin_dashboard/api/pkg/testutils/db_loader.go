package testutils

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB initializes an in-memory SQLite database for testing and runs migrations for provided models
func SetupTestDB(t *testing.T, models ...interface{}) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if len(models) > 0 {
		err = db.AutoMigrate(models...)
		if err != nil {
			t.Fatalf("failed to migrate test database: %v", err)
		}
	}

	return db
}
