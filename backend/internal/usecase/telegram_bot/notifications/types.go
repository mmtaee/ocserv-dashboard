package notifier

import (
	"context"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

type Sender interface {
	Send(chatID int64, text string) error
}

type Repository interface {
	Settings(ctx context.Context) (*models.TelegramSettings, error)
	AllAccounts(ctx context.Context) ([]models.TelegramAccount, error)
	OcservUserByID(ctx context.Context, id uint) (*models.OcservUser, error)
	MarkLowQuotaNotified(ctx context.Context, id uint, at time.Time) error
}
