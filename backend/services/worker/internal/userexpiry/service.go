package userexpiry

import (
	"context"
	"errors"
	"fmt"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	commonModels "github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	occtlDocker "github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/docker"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/occtl"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	stateManager "github.com/mmtaee/ocserv-dashboard/backend/services/worker/internal/expirystate"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"sync"
	"time"
)

// CronService handles all scheduled background jobs related to
// user expiration, monthly reactivation and auto-deletion.
//
// It supports both docker-mode and native ocserv mode.
type CronService struct {
	occtlHandler      occtl.OcservOcctlInterface
	ocservUserHandler user.OcservUserInterface
	occtlDockerRepo   occtlDocker.OcservOcctlUsersDocker
	dockerMode        bool
	stateMu           sync.Mutex
}

// NewCronService initializes the cron service.
// If dockerMode is true, docker-based occtl commands will be used.
// Otherwise, native ocserv handlers are used.
func NewCronService(dockerMode bool) *CronService {
	s := &CronService{
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
func (c *CronService) MissedCron(ctx context.Context) error {
	db := database.GetConnection()

	state := stateManager.NewCronState()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	lastRun := state.DailyLastRun.Truncate(24 * time.Hour)

	// daily missed job
	logger.Info("Start checking missing daily cron jobs")
	if state.DailyLastRun.IsZero() || lastRun.Before(today) {
		logger.Info("Running missed DAILY cron...")
		c.ExpireUsers(ctx, db)
		c.DeleteExpiredUsers(ctx, db)
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
		c.ActiveMonthlyUsers(ctx, db)
		state.MonthlyLastRun = today
	}
	logger.Info("Checking missing monthly cron jobs completed")

	if err := state.Save(); err != nil {
		return fmt.Errorf("save missed cron state: %w", err)
	}
	logger.Info("Saving missing cron jobs completed")
	return nil
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
func (c *CronService) UserExpiryCron(ctx context.Context) error {
	cronJob := cron.New(cron.WithSeconds())
	db := database.GetConnection()

	state := stateManager.NewCronState()

	// Every day at 00:01:00 — expire users
	_, err := cronJob.AddFunc("0 1 0 * * *", func() {
		c.ExpireUsers(ctx, db)

		c.stateMu.Lock()
		defer c.stateMu.Unlock()
		state.DailyLastRun = time.Now().Truncate(24 * time.Hour)
		if err := state.Save(); err != nil {
			logger.Error("Failed to save state: %v", err)
		}
	})
	if err != nil {
		return fmt.Errorf("register expire-users cron: %w", err)
	}
	logger.Info("Running user expiry cron...")

	// First and second day of each month at 00:01:00 — activate monthly users
	_, err = cronJob.AddFunc("0 1 0 1,2 * *", func() {
		c.ActiveMonthlyUsers(ctx, db)

		c.stateMu.Lock()
		defer c.stateMu.Unlock()
		state.MonthlyLastRun = time.Now().Truncate(24 * time.Hour)
		if saveErr := state.Save(); saveErr != nil {
			logger.Error("Failed to update state: %v", saveErr)
		}
	})
	if err != nil {
		return fmt.Errorf("register monthly-users cron: %w", err)
	}

	logger.Info("User activating Cron starting...")

	// Every day at 00:02:00 — delete expired users
	_, err3 := cronJob.AddFunc("0 2 0 * * *", func() {
		c.DeleteExpiredUsers(ctx, db)

		c.stateMu.Lock()
		defer c.stateMu.Unlock()
		state.DeleteInactiveUserLastRun = time.Now().Truncate(24 * time.Hour)
		if errSave := state.Save(); errSave != nil {
			logger.Error("Failed to save state: %v", errSave)
		}
	})
	if err3 != nil {
		return fmt.Errorf("register delete-users cron: %w", err3)
	}
	logger.Info("Running delete expired users cron...")

	//// Test: run every minute at second 0
	//_, err = cronJob.AddFunc("0 * * * * *", func() {
	//	c.DeleteExpiredUsers(ctx, db)
	//})

	cronJob.Start()

	<-ctx.Done()
	logger.Warn("Received context cancel, shutting down...")
	stopCtx := cronJob.Stop()
	select {
	case <-stopCtx.Done():
	case <-time.After(10 * time.Second):
		return errors.New("timed out waiting for user-expiry jobs")
	}
	logger.Info("User activating Cron stopped...")
	return nil
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
func (c *CronService) ExpireUsers(ctx context.Context, db *gorm.DB) {
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
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			wg.Done()
			wg.Wait()
			return
		}

		go func(u commonModels.OcservUser) {
			defer wg.Done()
			defer func() { <-sem }()

			// Update DB user
			if err2 := db.WithContext(ctx).Model(&u).Updates(map[string]interface{}{
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
func (c *CronService) ActiveMonthlyUsers(ctx context.Context, db *gorm.DB) {
	var users []commonModels.OcservUser
	today := time.Now().Truncate(24 * time.Hour)

	err := db.WithContext(ctx).
		Where("(expire_at IS NULL OR expire_at > ?)", today).
		Where("deactivated_at IS NOT NULL").
		Where("traffic_type IN ?", []string{
			commonModels.MonthlyReceive,
			commonModels.MonthlyTransmit,
			commonModels.MonthlyRxTx,
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
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			wg.Done()
			wg.Wait()
			return
		}

		go func(u commonModels.OcservUser) {
			defer wg.Done()
			defer func() { <-sem }()

			now := time.Now()

			if err2 := db.WithContext(ctx).Model(&u).Updates(map[string]interface{}{
				"rx":             0,
				"tx":             0,
				"usage_reset_at": &now,
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
func (c *CronService) DeleteExpiredUsers(ctx context.Context, db *gorm.DB) {
	var system models.System
	err := db.WithContext(ctx).First(&system).Error
	if err != nil {
		logger.Error("Failed to get system: %v", err)
		logger.Warn("set KeepInactiveUserDays to `30` and AutoDeleteInactiveUsers to `false`")
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
