package notifier

import (
	"context"
	"time"

	"github.com/mmtaee/ocserv-dashboard/common/models"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/i18n"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/repository"
)

const (
	checkInterval    = 30 * time.Minute
	notifyCooldown   = 24 * time.Hour
	bytesPerMegabyte = 1024 * 1024
	bytesPerGigabyte = 1024 * 1024 * 1024
)

// Sender is implemented by anything capable of delivering a chat message.
// We accept it as an interface to avoid a hard dependency on the bot Manager
// type, which keeps unit tests trivial.
type Sender interface {
	Send(chatID int64, text string) error
}

type Notifier struct {
	sender Sender
	repo   *repository.Repository
}

func New(sender Sender, repo *repository.Repository) *Notifier {
	return &Notifier{
		sender: sender,
		repo:   repo,
	}
}

// Run performs an initial scan and then a periodic scan every checkInterval.
// Returns when ctx is cancelled.
func (n *Notifier) Run(ctx context.Context) {
	tick := time.NewTicker(checkInterval)
	defer tick.Stop()

	n.scan(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			n.scan(ctx)
		}
	}
}

func (n *Notifier) scan(ctx context.Context) {
	settings, err := n.repo.Settings(ctx)
	if err != nil || !settings.Enabled {
		return
	}
	thresholdBytes := int64(settings.LowQuotaThresholdMB) * bytesPerMegabyte

	accounts, err := n.repo.AllAccounts(ctx)
	if err != nil {
		logger.Warn("telegram_bot: notifier list accounts: %v", err)
		return
	}

	now := time.Now()
	for _, account := range accounts {
		user, err := n.repo.OcservUserByID(ctx, account.OcservUserID)
		if err != nil {
			continue
		}
		if user.IsLocked || user.DeactivatedAt != nil {
			continue
		}
		if user.TrafficType == models.Free {
			continue
		}

		quotaBytes := int64(user.TrafficSize)
		var usedBytes int64
		switch user.TrafficType {
		case models.MonthlyTransmit, models.TotallyTransmit:
			usedBytes = int64(user.Tx)
		case models.MonthlyReceive, models.TotallyReceive:
			usedBytes = int64(user.Rx)
		case models.MonthlyRxTx, models.TotallyRxTx:
			usedBytes = int64(user.Rx) + int64(user.Tx)
		default:
			continue
		}
		remaining := quotaBytes - usedBytes
		if remaining <= 0 || remaining >= thresholdBytes {
			continue
		}
		if account.LastLowQuotaNotifiedAt != nil && now.Sub(*account.LastLowQuotaNotifiedAt) < notifyCooldown {
			continue
		}

		remainingMB := int(remaining / bytesPerMegabyte)
		text := i18n.T(account.Language, i18n.LowQuotaWarning, user.Username, remainingMB)
		if err := n.sender.Send(account.ChatID, text); err != nil {
			logger.Warn("telegram_bot: notifier send failed: %v", err)
			continue
		}
		if err := n.repo.MarkLowQuotaNotified(ctx, account.ID, now); err != nil {
			logger.Warn("telegram_bot: notifier mark notified: %v", err)
		}
	}
}
