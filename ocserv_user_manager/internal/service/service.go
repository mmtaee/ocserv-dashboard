package service

import (
	"context"
	"sync"
	"time"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/occtl"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/ocserv_user_manager/pkg/state"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type CronService struct {
	occtlHandler      occtl.OcservOcctlInterface
	ocservUserHandler user.OcservUserInterface
	// TODO: occtlDockerRepo (for webhook-based docker mode) is temporarily disabled
	// occtlDockerRepo   occtlDocker.OcservOcctlUsersDocker
}

func NewCronService() *CronService {
	s := &CronService{}

	// TODO: Docker mode (webhook-based occtl/lock operations) is temporarily disabled
	// For now, only non-docker mode (direct occtl/ocpasswd calls) is supported
	s.occtlHandler = occtl.NewOcservOcctl()
	s.ocservUserHandler = user.NewOcservUser()

	return s
}

// MissedCron checks and runs any missed cron jobs
func (c *CronService) MissedCron() {
	db := database.GetConnection()

	stateManager := state.NewCronState()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	lastDailyRun := stateManager.DailyLastRun.Truncate(24 * time.Hour)

	// Daily missed job
	logger.Info("Start checking missing daily cron jobs")
	if stateManager.DailyLastRun.IsZero() || lastDailyRun.Before(today) {
		logger.Info("Running missed DAILY cron...")
		c.ExpireUsers(context.Background(), db)
		c.DeleteExpiredUsers(context.Background(), db)
		stateManager.DailyLastRun = today
	} else {
		logger.Info("Daily cron already ran today, skipping.")
	}
	logger.Info("Checking missing daily cron jobs completed")

	// Monthly missed job
	logger.Info("Start checking missing monthly cron jobs")
	firstDay := today.Day() == 1
	newMonth := stateManager.MonthlyLastRun.IsZero() || stateManager.MonthlyLastRun.Month() != today.Month()

	if firstDay && newMonth {
		logger.Info("Running missed MONTHLY cron...")
		c.ActivateMonthlyUsers(context.Background(), db)
		stateManager.MonthlyLastRun = today
	}
	logger.Info("Checking missing monthly cron jobs completed")

	if err := stateManager.Save(); err != nil {
		logger.Fatal("Failed to save state: %v", err)
	}
	logger.Info("Saving missing cron jobs completed")
}

// UserExpiryCron starts the cron scheduler
func (c *CronService) UserExpiryCron(ctx context.Context) {
	cronJob := cron.New(cron.WithSeconds())
	db := database.GetConnection()

	stateManager := state.NewCronState()

	// Every day at 00:01:00 — expire users
	_, err := cronJob.AddFunc("0 1 0 * * *", func() {
		c.ExpireUsers(ctx, db)

		stateManager.DailyLastRun = time.Now().Truncate(24 * time.Hour)
		if err := stateManager.Save(); err != nil {
			logger.Error("Failed to save state: %v", err)
		}
	})
	if err != nil {
		logger.Fatal("Failed to add cron job: %v", err)
	}
	logger.Info("Running user expiry cron...")

	// First and second day of each month at 00:01:00 — activate monthly users
	_, err = cronJob.AddFunc("0 1 0 1,2 * *", func() {
		c.ActivateMonthlyUsers(ctx, db)

		stateManager.MonthlyLastRun = time.Now().Truncate(24 * time.Hour)
		if err := stateManager.Save(); err != nil {
			logger.Error("Failed to update state: %v", err)
		}
	})
	if err != nil {
		logger.Fatal("Failed to add cron job: %v", err)
	}

	logger.Info("User activating Cron starting...")

	// Every day at 00:02:00 — delete expired users
	_, err = cronJob.AddFunc("0 2 0 * * *", func() {
		c.DeleteExpiredUsers(ctx, db)

		stateManager.DailyLastRun = time.Now().Truncate(24 * time.Hour)
		if errSave := stateManager.Save(); errSave != nil {
			logger.Error("Failed to save state: %v", errSave)
		}
	})
	if err != nil {
		logger.Fatal("Failed to add cron job: %v", err)
	}
	logger.Info("Running delete expired users cron...")

	cronJob.Start()

	<-ctx.Done()
	logger.Warn("Received context cancel, shutting down...")
	cronJob.Stop()
	logger.Info("User activating Cron stopped...")
}

// ExpireUsers finds expired users and deactivates them
func (c *CronService) ExpireUsers(ctx context.Context, db *gorm.DB) {
	var users []models.OcservUser

	pastDay := time.Now().UTC().AddDate(0, 0, -1)
	err := db.WithContext(ctx).
		Select("id", "username", "expire_at").
		Where("expire_at IS NOT NULL").
		Where("deactivated_at IS NULL").
		Where("expire_at < ?", pastDay).
		Find(&users).Error
	if err != nil {
		logger.Error("Failed to get users: %v", err)
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
			now := time.Now()
			if err2 := db.Model(&u).Updates(map[string]interface{}{
				"deactivated_at": now,
				"is_locked":      true,
			}).Error; err2 != nil {
				logger.Error("Failed to update user: %v", err2)
				return
			}

			if _, err3 := c.occtlHandler.DisconnectUser(u.Username); err3 != nil {
				logger.Error("Failed to disconnect user %s: %v", u.Username, err3)
			}
			if _, err4 := c.ocservUserHandler.Lock(u.Username); err4 != nil {
				logger.Error("Failed to lock user %s: %v", u.Username, err4)
			}
		}(u)
	}

	wg.Wait()
}

// ActivateMonthlyUsers reactivates monthly traffic users
func (c *CronService) ActivateMonthlyUsers(ctx context.Context, db *gorm.DB) {
	var users []models.OcservUser
	today := time.Now().Truncate(24 * time.Hour)

	err := db.WithContext(ctx).
		Where("(expire_at IS NULL OR expire_at > ?)", today).
		Where("deactivated_at IS NOT NULL").
		Where("traffic_type IN ?", []string{
			models.MonthlyTransmit,
			models.MonthlyReceive,
			models.MonthlyRxTx,
		}).
		Find(&users).Error
	if err != nil {
		logger.Error("Failed to get users: %v", err)
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

			now := time.Now()

			if err2 := db.Model(&u).Updates(map[string]interface{}{
				"rx":             0,
				"tx":             0,
				"usage_reset_at": &now,
				"deactivated_at": nil,
				"is_locked":      false,
			}).Error; err2 != nil {
				logger.Error("Failed to update user %s: %v", u.Username, err2)
				return
			}

			if _, err2 := c.ocservUserHandler.UnLock(u.Username); err2 != nil {
				logger.Error("Failed to unlock user %s: %v", u.Username, err2)
			}

		}(u)
	}

	wg.Wait()
}

// DeleteExpiredUsers permanently deletes inactive users
func (c *CronService) DeleteExpiredUsers(ctx context.Context, db *gorm.DB) {
	var system models.System
	err := db.WithContext(ctx).First(&system).Error
	if err != nil {
		logger.Error("Failed to get system: %v", err)
		logger.Warn("Set KeepInactiveUserDays to 30 and AutoDeleteInactiveUsers to false")
		system.KeepInactiveUserDays = 30
		system.AutoDeleteInactiveUsers = false
	}

	if !system.AutoDeleteInactiveUsers {
		logger.Warn("User auto-delete is disabled")
		return
	}

	if system.KeepInactiveUserDays < 1 {
		logger.Warn("User keep inactive days is lower than 1 day")
		return
	}

	cutoffDate := time.Now().AddDate(0, 0, -system.KeepInactiveUserDays).UTC()
	result := db.WithContext(ctx).
		Where("expire_at IS NOT NULL AND expire_at <= ?", cutoffDate).
		Delete(&models.OcservUser{})

	if result.Error != nil {
		logger.Error("Failed to delete inactive users: %v", result.Error)
		return
	}

	if result.RowsAffected == 0 {
		logger.Info("No inactive users found for deletion")
		return
	}

	logger.Info("Deleted %d inactive users", result.RowsAffected)
}
