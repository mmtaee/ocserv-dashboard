package service

import (
	"context"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/occtl"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/user"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	stateManager "github.com/mmtaee/ocserv-users-management/user_expiry/pkg/state"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"log"
	"sync"
	"time"
)

type CornService struct {
	occtlHandler      occtl.OcservOcctlInterface
	ocservUserHandler user.OcservUserInterface
}

func NewCornService() *CornService {
	return &CornService{
		occtlHandler:      occtl.NewOcservOcctl(),
		ocservUserHandler: user.NewOcservUser(),
	}
}

func (c *CornService) MissedCron() {
	db := database.GetConnection()

	state := stateManager.NewCronState()
	today := time.Now().Truncate(24 * time.Hour)

	// daily missed job
	if state.DailyLastRun.IsZero() || state.DailyLastRun.Before(today) {
		log.Println("Running missed DAILY cron...")
		c.ExpireUsers(context.Background(), db)
		state.DailyLastRun = today
	}

	// monthly missed job
	firstDay := today.Day() == 1
	newMonth := state.MonthlyLastRun.IsZero() || state.MonthlyLastRun.Month() != today.Month()

	if firstDay && newMonth {
		log.Println("Running missed MONTHLY cron...")
		c.ActiveMonthlyUsers(context.Background(), db)
		state.MonthlyLastRun = today
	}

	if err := state.Save(); err != nil {
		log.Fatalf("Failed to save state: %v", err)
	}
}

func (c *CornService) UserExpiryCron(ctx context.Context) {
	cronJob := cron.New(cron.WithSeconds())
	db := database.GetConnection()

	state := stateManager.NewCronState()

	// Every day at 00:01:00 — expire users
	_, err := cronJob.AddFunc("0 1 0 * * *", func() {
		c.ExpireUsers(ctx, db)

		state.DailyLastRun = time.Now().Truncate(24 * time.Hour)
		if err := state.Save(); err != nil {
			log.Println("Failed to update state file:", err)
		}
	})
	if err != nil {
		log.Printf("Failed to schedule cron: %v", err)
		return
	}
	log.Println("UserExpiry Cron starting...")

	// First and second day of each month at 00:01:00 — activate monthly users
	_, err = cronJob.AddFunc("0 1 0 1,2 * *", func() {
		c.ActiveMonthlyUsers(ctx, db)

		state.MonthlyLastRun = time.Now().Truncate(24 * time.Hour)
		if err = state.Save(); err != nil {
			log.Println("Failed to update state file:", err)
		}
	})
	log.Println("User activating Cron starting...")

	//// Test: run every minute at second 0
	//_, err = c.AddFunc("0 * * * * *", func() {
	//	ActiveMonthlyUsers(ctx, db)
	//})

	cronJob.Start()

	<-ctx.Done()
	log.Println("Stopping Cron service ...")
	cronJob.Stop()
	log.Println("Cron stopped")
}

func (c *CornService) ExpireUsers(ctx context.Context, db *gorm.DB) {
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
			if _, err2 := c.occtlHandler.DisconnectUser(u.Username); err2 != nil {
				log.Printf("Failed to disconnect user %s: %v", u.Username, err2)
				return
			}

			// Lock user in ocserv
			if _, err2 := c.ocservUserHandler.Lock(u.Username); err2 != nil {
				log.Printf("Failed to lock user %s: %v", u.Username, err2)
				return
			}

		}(u)
	}

	wg.Wait()
}

func (c *CornService) ActiveMonthlyUsers(ctx context.Context, db *gorm.DB) {
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

			if _, err2 := c.ocservUserHandler.UnLock(u.Username); err2 != nil {
				log.Printf("Failed to unlock user %s: %v", u.Username, err2)
			}

		}(u)
	}

	wg.Wait()
}
