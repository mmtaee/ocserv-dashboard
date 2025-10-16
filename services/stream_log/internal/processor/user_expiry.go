package processor

import (
	"context"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"log"
	"time"
)

func UserExpiryCron(ctx context.Context) {
	c := cron.New(cron.WithSeconds())
	db := database.GetConnection()

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
		Where("expire_at < ? AND deactivated_at IS NULL", pastDay).
		Find(&users).Error
	if err != nil {
		log.Printf("Failed to find users: %v", err)
	}

	now := time.Now()

	if len(users) > 0 {
		for _, user := range users {
			user.DeactivatedAt = &now
			user.IsLocked = true
			db.WithContext(ctx).Save(&user)
			_, err = ocservOcctlRepo.DisconnectUser(user.Username)
			if err != nil {
				log.Printf("Failed to disconnect user: %v", err)
				continue
			}
			_, err = ocservUserRepo.Lock(user.Username)
			if err != nil {
				log.Printf("Failed to lock user: %v", err)
				continue
			}
		}
	}

}

func ActiveMonthlyUsers(ctx context.Context, db *gorm.DB) {
	var users []models.OcservUser
	today := time.Now().Format("2006-01-02")

	err := db.WithContext(ctx).Where(
		"expire_at IS NOT NULL AND expire_at > ? AND deactivated_at IS NOT NULL AND traffic_type IN ?",
		today, []string{models.MonthlyReceive, models.MonthlyTransmit},
	).Find(&users).Error
	if err != nil {
		log.Printf("Failed to find users: %v", err)
	}

	if len(users) > 0 {
		for _, user := range users {
			user.DeactivatedAt = nil
			user.IsLocked = false
			db.WithContext(ctx).Save(&user)
			_, err = ocservUserRepo.UnLock(user.Username)
			if err != nil {
				log.Printf("Failed to unlock user: %v", err)
				continue
			}
		}
	}
}
