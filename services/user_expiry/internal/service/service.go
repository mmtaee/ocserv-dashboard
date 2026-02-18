package service

import (
	"context"
	commonModels "github.com/mmtaee/ocserv-users-management/common/models"
	occtlDocker "github.com/mmtaee/ocserv-users-management/common/occtl_docker"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/occtl"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/user"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"github.com/mmtaee/ocserv-users-management/user_expiry/internal/models"
	stateManager "github.com/mmtaee/ocserv-users-management/user_expiry/pkg/state"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"sync"
	"time"
)

// CornService handles all scheduled background jobs related to
// user expiration, monthly reactivation and auto-deletion.
//
// It supports both docker-mode and native ocserv mode.
type CornService struct {
	occtlHandler      occtl.OcservOcctlInterface
	ocservUserHandler user.OcservUserInterface
	occtlDockerRepo   occtlDocker.OcservOcctlUsersDocker
	dockerMode        bool
}

// NewCornService initializes cron service.
// If dockerMode is true, docker-based occtl commands will be used.
// Otherwise, native ocserv handlers are used.
func NewCornService(dockerMode bool) *CornService {
	s := &CornService{
		dockerMode: dockerMode,
	}

	if dockerMode {
		s.occtlDockerRepo = occtlDocker.NewOcservOcctlDocker()
	} else {
		s.occtlHandler = occtl.NewOcservOcctl()
		s.ocservUserHandler = user.NewOcservUser()
	}
	return s
}

// MissedCron checks whether daily or monthly cron jobs were missed
// (for example if the service was down) and executes them manually.
//
// It ensures:
// - ExpireUsers runs once per day
// - ActiveMonthlyUsers runs once per month (on first day)
func (c *CornService) MissedCron() {
	db := database.GetConnection()

	state := stateManager.NewCronState()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	lastRun := state.DailyLastRun.Truncate(24 * time.Hour)

	// daily missed job
	logger.Info("Start checking missing daily cron jobs")
	if state.DailyLastRun.IsZero() || lastRun.Before(today) {
		logger.Info("Running missed DAILY cron...")
		c.ExpireUsers(context.Background(), db)
		c.DeleteExpiredUsers(context.Background(), db)
		state.DailyLastRun = today
	} else {
		logger.Info("Daily cron already ran today, skipping.")
	}
	logger.Info("Checking missing daily cron jobs completed")

	// monthly missed job
	logger.Info("start checking missing monthly cron jobs completed")
	firstDay := today.Day() == 1
	newMonth := state.MonthlyLastRun.IsZero() || state.MonthlyLastRun.Month() != today.Month()

	if firstDay && newMonth {
		logger.Info("Running missed MONTHLY cron...")
		c.ActiveMonthlyUsers(context.Background(), db)
		state.MonthlyLastRun = today
	}
	logger.Info("Checking missing monthly cron jobs completed")

	if err := state.Save(); err != nil {
		logger.Fatal("Failed to save state: %v", err)
	}
	logger.Info("Saving missing cron jobs completed")
}

// UserExpiryCron registers and starts all cron jobs:
//
// Daily (00:01:00):
//   - ExpireUsers
//
// Daily (00:02:00):
//   - DeleteExpiredUsers
//
// Monthly (1st & 2nd day at 00:01:00):
//   - ActiveMonthlyUsers
//
// The cron stops when context is canceled.
func (c *CornService) UserExpiryCron(ctx context.Context) {
	cronJob := cron.New(cron.WithSeconds())
	db := database.GetConnection()

	state := stateManager.NewCronState()

	// Every day at 00:01:00 — expire users
	_, err := cronJob.AddFunc("0 1 0 * * *", func() {
		c.ExpireUsers(ctx, db)

		state.DailyLastRun = time.Now().Truncate(24 * time.Hour)
		if err := state.Save(); err != nil {
			logger.Error("Failed to save state: %v", err)
		}
	})
	if err != nil {
		logger.Fatal("Failed to add cron job: %v", err)
	}
	logger.Info("Running user expiry cron...")

	// First and second day of each month at 00:01:00 — activate monthly users
	_, err = cronJob.AddFunc("0 1 0 1,2 * *", func() {
		c.ActiveMonthlyUsers(ctx, db)

		state.MonthlyLastRun = time.Now().Truncate(24 * time.Hour)
		if err = state.Save(); err != nil {
			logger.Error("Failed to update state: %v", err)
		}
	})
	if err != nil {
		logger.Fatal("Failed to add cron job: %v", err)
	}

	logger.Info("User activating Cron starting...")

	// Every day at 00:02:00 — delete expired users
	_, err3 := cronJob.AddFunc("0 2 0 * * *", func() {
		c.DeleteExpiredUsers(ctx, db)

		state.DailyLastRun = time.Now().Truncate(24 * time.Hour)
		if errSave := state.Save(); errSave != nil {
			logger.Error("Failed to save state: %v", errSave)
		}
	})
	if err3 != nil {
		logger.Fatal("Failed to add cron job: %v", err3)
	}
	logger.Info("Running delete expired users cron...")

	//// Test: run every minute at second 0
	//_, err = cronJob.AddFunc("0 * * * * *", func() {
	//	c.DeleteExpiredUsers(ctx, db)
	//})

	cronJob.Start()

	<-ctx.Done()
	logger.Warn("Received context cancel, shutting down...")
	cronJob.Stop()
	logger.Info("User activating Cron stopped...")
}

// ExpireUsers finds users whose expire_at has passed
// and deactivates them.
//
// Actions performed per user:
//   - Set deactivated_at = now
//   - Set is_locked = true
//   - Disconnect active session
//   - Lock user in ocserv
//
// Runs concurrently with max 10 workers.
func (c *CornService) ExpireUsers(ctx context.Context, db *gorm.DB) {
	var users []commonModels.OcservUser

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

		go func(u commonModels.OcservUser) {
			defer wg.Done()
			defer func() { <-sem }()

			// Update DB user
			if err2 := db.Model(&u).Updates(map[string]interface{}{ // CHANGED: using &u (copied)
				"deactivated_at": time.Now(),
				"is_locked":      true,
			}).Error; err2 != nil {
				logger.Error("Failed to update user: %v", err2)
				return
			}

			var (
				disconnect func(string) (string, error)
				lock       func(string) (string, error)
			)

			if c.dockerMode {
				disconnect = c.occtlDockerRepo.DisconnectUser
				lock = c.occtlDockerRepo.Lock
			} else {
				disconnect = c.occtlHandler.DisconnectUser
				lock = c.ocservUserHandler.Lock
			}

			if _, err3 := disconnect(u.Username); err3 != nil {
				logger.Error("Failed to disconnect user %s: %v", u.Username, err3)
			}
			if _, err4 := lock(u.Username); err4 != nil {
				logger.Error("Failed to lock user %s: %v", u.Username, err4)
			}
			return
		}(u)
	}

	wg.Wait()
}

// ActiveMonthlyUsers reactivates monthly traffic users
// at the beginning of a new month.
//
// Conditions:
//   - User is currently deactivated
//   - Traffic type is MonthlyReceive or MonthlyTransmit
//   - User is not expired
//
// Actions:
//   - Reset rx and tx counters
//   - Remove deactivated_at
//   - Unlock user
//
// Runs concurrently with max 10 workers.
func (c *CornService) ActiveMonthlyUsers(ctx context.Context, db *gorm.DB) {
	var users []commonModels.OcservUser
	today := time.Now().Truncate(24 * time.Hour)

	err := db.WithContext(ctx).
		Where("(expire_at IS NULL OR expire_at > ?)", today).
		Where("deactivated_at IS NOT NULL").
		Where("traffic_type IN ?", []string{
			commonModels.MonthlyReceive,
			commonModels.MonthlyTransmit,
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

		go func(u commonModels.OcservUser) {
			defer wg.Done()
			defer func() { <-sem }()

			if err2 := db.Model(&u).Updates(map[string]interface{}{
				"rx":             0,
				"tx":             0,
				"deactivated_at": nil,
				"is_locked":      false,
			}).Error; err2 != nil {
				logger.Error("Failed to update user %s: %v", u.Username, err2)
				return
			}

			var unlock func(string) (string, error)

			if c.dockerMode {
				unlock = c.occtlDockerRepo.Unlock
			} else {
				unlock = c.ocservUserHandler.UnLock
			}
			if _, err2 := unlock(u.Username); err2 != nil {
				logger.Error("Failed to unlock user %s: %v", u.Username, err2)
			}

		}(u)
	}

	wg.Wait()
}

// DeleteExpiredUsers permanently deletes users who:
//
//   - Are deactivated
//   - Have been inactive longer than system.KeepInactiveUserDays
//   - AutoDeleteInactiveUsers setting is enabled
//
// Uses bulk delete for performance and logs number of deleted rows.
func (c *CornService) DeleteExpiredUsers(ctx context.Context, db *gorm.DB) {
	var system models.System
	err := db.WithContext(ctx).First(&system).Error
	if err != nil {
		logger.Error("Failed to get system: %v", err)
		return
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
		Delete(&commonModels.OcservUser{})

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
