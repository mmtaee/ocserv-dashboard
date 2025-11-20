package main

import (
	"context"
	"flag"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/occtl"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/user"
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	occtlHandler      = occtl.NewOcservOcctl()
	ocservUserHandler = user.NewOcservUser()
	debug             bool
)

func main() {
	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	config.Init(debug, "", 8888)
	database.Connect()

	go func() {
		UserExpiryCron(ctx)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("\nReceived signal: %s\n", sig)
	cancel()

}

func UserExpiryCron(ctx context.Context) {
	c := cron.New(cron.WithSeconds())
	db := database.GetConnection()
	
	// Every day at 00:01:00 — expire users
	_, err := c.AddFunc("0 1 0 * * *", func() {
		ExpireUsers(ctx, db)
	})
	if err != nil {
		log.Printf("Failed to schedule cron: %v", err)
		return
	}
	log.Println("UserExpiry Cron starting...")

	// First and second day of each month at 00:01:00 — activate monthly users
	_, err = c.AddFunc("0 1 0 1,2 * *", func() {
		ActiveMonthlyUsers(ctx, db)
	})
	log.Println("User activating Cron starting...")

	//// Test: run every minute at second 0
	//_, err = c.AddFunc("0 * * * * *", func() {
	//	ActiveMonthlyUsers(ctx, db)
	//})

	c.Start()

	<-ctx.Done()
	log.Println("Stopping Cron service ...")
	c.Stop()
	log.Println("Cron stopped")
}

func ExpireUsers(ctx context.Context, db *gorm.DB) {
	var users []models.OcservUser

	pastDay := time.Now().AddDate(0, 0, -1)
	err := db.WithContext(ctx).
		Where("expire_at IS NOT NULL").
		Where("deactivated_at IS NULL").
		Where("expire_at < ?", pastDay).
		Find(&users).Error
	if err != nil {
		log.Printf("Failed to find users: %v", err)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for _, u := range users {
		wg.Add(1)
		sem <- struct{}{}

		go func(u models.OcservUser) {
			defer wg.Done()
			defer func() { <-sem }()

			// Update DB user
			if err2 := db.Model(&u).Updates(map[string]interface{}{ // CHANGED: using &u (copied)
				"deactivated_at": time.Now(),
				"is_locked":      true,
			}).Error; err2 != nil {
				log.Printf("Failed to update user %s: %v", u.Username, err2)
				return
			}

			// Disconnect user from ocserv
			if _, err2 := occtlHandler.DisconnectUser(u.Username); err2 != nil {
				log.Printf("Failed to disconnect user %s: %v", u.Username, err2)
				return
			}

			// Lock user in ocserv
			if _, err2 := ocservUserHandler.Lock(u.Username); err2 != nil {
				log.Printf("Failed to lock user %s: %v", u.Username, err2)
				return
			}

		}(u)
	}

	wg.Wait()
}

func ActiveMonthlyUsers(ctx context.Context, db *gorm.DB) {
	var users []models.OcservUser
	today := time.Now().Truncate(24 * time.Hour)

	err := db.WithContext(ctx).
		Where("expire_at IS NOT NULL").
		Where("expire_at > ?", today).
		Where("deactivated_at IS NOT NULL").
		Where("traffic_type IN ?", []string{
			models.MonthlyReceive,
			models.MonthlyTransmit,
		}).
		Find(&users).Error
	if err != nil {
		log.Printf("Failed to find users: %v", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for _, u := range users {
		wg.Add(1)
		sem <- struct{}{}

		go func(u models.OcservUser) {
			defer wg.Done()
			defer func() { <-sem }()

			if err2 := db.Model(&u).Updates(map[string]interface{}{
				"rx":             0,
				"tx":             0,
				"deactivated_at": nil,
				"is_locked":      false,
			}).Error; err2 != nil {
				log.Printf("Failed to activate user %s: %v", u.Username, err2)
				return
			}

			if _, err2 := ocservUserHandler.UnLock(u.Username); err2 != nil {
				log.Printf("Failed to unlock user %s: %v", u.Username, err2)
			}

		}(u)
	}

	wg.Wait()
}
