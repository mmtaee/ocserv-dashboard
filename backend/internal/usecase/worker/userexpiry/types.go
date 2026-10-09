package userexpiry

import (
	"context"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

type Repository interface {
	ExpiredUsers(ctx context.Context, before time.Time) ([]models.OcservUser, error)
	Deactivate(ctx context.Context, id uint, at time.Time) error
	MonthlyUsers(ctx context.Context, today time.Time) ([]models.OcservUser, error)
	Reactivate(ctx context.Context, id uint, at time.Time) error
	SystemSettings(ctx context.Context) (*models.System, error)
	DeleteExpired(ctx context.Context, cutoff time.Time) (int64, error)
}

type AccessController interface {
	Disconnect(username string) error
	Lock(username string) error
	Unlock(username string) error
}
