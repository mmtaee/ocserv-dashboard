package userexpiry

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	state "github.com/mmtaee/ocserv-dashboard/backend/internal/services/worker/cronstate"
	userexpiryusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/worker/userexpiry"
	"github.com/robfig/cron/v3"
)

type CronService struct {
	usecase *userexpiryusecase.Usecase
	stateMu sync.Mutex
}

func NewCronService(usecase *userexpiryusecase.Usecase) *CronService {
	return &CronService{usecase: usecase}
}

func (s *CronService) Run(ctx context.Context) error {
	if err := s.runMissed(ctx); err != nil {
		return err
	}

	journal := state.NewCronState()
	fatalErrors := make(chan error, 1)
	scheduler := cron.New(cron.WithSeconds())
	register := func(spec string, job func(context.Context, time.Time) error, updateState func(time.Time)) error {
		_, err := scheduler.AddFunc(spec, func() {
			now := time.Now()
			if err := job(ctx, now); err != nil {
				select {
				case fatalErrors <- err:
				default:
				}
				return
			}
			s.stateMu.Lock()
			defer s.stateMu.Unlock()
			updateState(now.Truncate(24 * time.Hour))
			if err := journal.Save(); err != nil {
				select {
				case fatalErrors <- fmt.Errorf("save cron state: %w", err):
				default:
				}
			}
		})
		return err
	}
	if err := register("0 1 0 * * *", func(ctx context.Context, now time.Time) error {
		return s.usecase.Expire(ctx, now)
	}, func(at time.Time) { journal.DailyLastRun = at }); err != nil {
		return fmt.Errorf("register expire-users cron: %w", err)
	}
	if err := register("0 1 0 1,2 * *", func(ctx context.Context, now time.Time) error {
		return s.usecase.ActivateMonthly(ctx, now)
	}, func(at time.Time) { journal.MonthlyLastRun = at }); err != nil {
		return fmt.Errorf("register monthly-users cron: %w", err)
	}
	if err := register("0 2 0 * * *", func(ctx context.Context, now time.Time) error {
		_, err := s.usecase.DeleteInactive(ctx, now)
		return err
	}, func(at time.Time) { journal.DeleteInactiveUserLastRun = at }); err != nil {
		return fmt.Errorf("register delete-users cron: %w", err)
	}

	scheduler.Start()
	select {
	case <-ctx.Done():
	case err := <-fatalErrors:
		stopCtx := scheduler.Stop()
		select {
		case <-stopCtx.Done():
			return err
		case <-time.After(10 * time.Second):
			return fmt.Errorf("%w; timed out waiting for user-expiry jobs", err)
		}
	}
	stopCtx := scheduler.Stop()
	select {
	case <-stopCtx.Done():
		return nil
	case <-time.After(10 * time.Second):
		return errors.New("timed out waiting for user-expiry jobs")
	}
}

func (s *CronService) runMissed(ctx context.Context) error {
	journal := state.NewCronState()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if journal.DailyLastRun.IsZero() || journal.DailyLastRun.Truncate(24*time.Hour).Before(today) {
		if err := s.usecase.Expire(ctx, today); err != nil {
			return err
		}
		if _, err := s.usecase.DeleteInactive(ctx, today); err != nil {
			return err
		}
		journal.DailyLastRun = today
		journal.DeleteInactiveUserLastRun = today
	}
	if today.Day() == 1 && (journal.MonthlyLastRun.IsZero() || journal.MonthlyLastRun.Month() != today.Month()) {
		if err := s.usecase.ActivateMonthly(ctx, today); err != nil {
			return err
		}
		journal.MonthlyLastRun = today
	}
	if err := journal.Save(); err != nil {
		return fmt.Errorf("save missed cron state: %w", err)
	}
	logger.Info("User expiry missed-job check completed")
	return nil
}
