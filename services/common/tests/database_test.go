// go test ./common/tests -run TestConcurrentWrites -v

package tests

import (
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"sync"
	"testing"
)

type User struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

func TestConcurrentWrites(t *testing.T) {
	config.Init(false, "", 0)
	database.Connect()
	db := database.GetConnection()

	// Auto migrate
	err := db.AutoMigrate(&User{})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errors := 0
	mutex := sync.Mutex{}

	// 100 concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			err := db.Create(&User{Name: "User"}).Error
			if err != nil {
				mutex.Lock()
				errors++
				mutex.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if errors > 0 {
		t.Fatalf("Got %d errors (database likely locked)", errors)
	}

	t.Log("No locking errors detected ✅")
}
