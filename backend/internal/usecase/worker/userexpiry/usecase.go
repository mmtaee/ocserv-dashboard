package userexpiry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Usecase struct {
	repository Repository
	access     AccessController
}

func New(repository Repository, access AccessController) *Usecase {
	return &Usecase{repository: repository, access: access}
}

func (u *Usecase) Expire(ctx context.Context, now time.Time) error {
	users, err := u.repository.ExpiredUsers(ctx, now.UTC())
	if err != nil {
		return fmt.Errorf("find expired users: %w", err)
	}
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(10)
	for _, item := range users {
		item := item
		group.Go(func() error {
			if err := u.repository.Deactivate(groupCtx, item.ID, now); err != nil {
				return fmt.Errorf("deactivate %s: %w", item.Username, err)
			}
			if err := u.access.Disconnect(item.Username); err != nil {
				return fmt.Errorf("disconnect %s: %w", item.Username, err)
			}
			if err := u.access.Lock(item.Username); err != nil {
				return fmt.Errorf("lock %s: %w", item.Username, err)
			}
			return nil
		})
	}
	return group.Wait()
}

func (u *Usecase) ActivateMonthly(ctx context.Context, now time.Time) error {
	users, err := u.repository.MonthlyUsers(ctx, now.Truncate(24*time.Hour))
	if err != nil {
		return fmt.Errorf("find monthly users: %w", err)
	}
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(10)
	for _, item := range users {
		item := item
		group.Go(func() error {
			if err := u.repository.Reactivate(groupCtx, item.ID, now); err != nil {
				return fmt.Errorf("reactivate %s: %w", item.Username, err)
			}
			if err := u.access.Unlock(item.Username); err != nil {
				return fmt.Errorf("unlock %s: %w", item.Username, err)
			}
			return nil
		})
	}
	return group.Wait()
}

func (u *Usecase) DeleteInactive(ctx context.Context, now time.Time) (int64, error) {
	settings, err := u.repository.SystemSettings(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("load system settings: %w", err)
	}
	if !settings.AutoDeleteInactiveUsers || settings.KeepInactiveUserDays < 1 {
		return 0, nil
	}
	return u.repository.DeleteExpired(ctx, now.UTC().AddDate(0, 0, -settings.KeepInactiveUserDays))
}
