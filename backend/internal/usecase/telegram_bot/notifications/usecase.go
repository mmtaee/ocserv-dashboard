package notifier

import (
	"context"
	"fmt"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/telegram_bot/i18n"
)

const (
	checkInterval    = 30 * time.Minute
	notifyCooldown   = 24 * time.Hour
	bytesPerMegabyte = 1024 * 1024
	bytesPerGigabyte = 1024 * 1024 * 1024
)

type Notifier struct {
	sender Sender
	repo   Repository
}

func New(sender Sender, repo Repository) *Notifier {
	return &Notifier{
		sender: sender,
		repo:   repo,
	}
}

// Run performs an initial scan and then a periodic scan every checkInterval.
// Returns when ctx is cancelled.
func (n *Notifier) Run(ctx context.Context) error {
	tick := time.NewTicker(checkInterval)
	defer tick.Stop()

	if err := n.scan(ctx); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
			if err := n.scan(ctx); err != nil {
				return err
			}
		}
	}
}

func (n *Notifier) scan(ctx context.Context) error {
	settings, err := n.repo.Settings(ctx)
	if err != nil {
		return fmt.Errorf("load notification settings: %w", err)
	}
	if !settings.Enabled {
		return nil
	}
	thresholdBytes := int64(settings.LowQuotaThresholdMB) * bytesPerMegabyte

	accounts, err := n.repo.AllAccounts(ctx)
	if err != nil {
		return fmt.Errorf("list notification accounts: %w", err)
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
	return nil
}
