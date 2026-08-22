package telegrambot

import (
	"context"
	"fmt"
	"os"

	"github.com/mmtaee/ocserv-dashboard/backend/services/telegram_bot/internal/bot"
	"github.com/mmtaee/ocserv-dashboard/backend/services/telegram_bot/internal/i18n"
	"github.com/mmtaee/ocserv-dashboard/backend/services/telegram_bot/internal/notifier"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	receiptsDir string
}

func New(receiptsDir string) *Service {
	return &Service{receiptsDir: receiptsDir}
}

func (s *Service) Run(ctx context.Context) error {
	if err := os.MkdirAll(s.receiptsDir, 0o750); err != nil {
		return fmt.Errorf("create telegram receipt directory: %w", err)
	}
	i18n.Init()
	manager := bot.NewManager(s.receiptsDir)
	notifications := notifier.New(manager, manager.Repo())
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		manager.Run(groupCtx)
		return nil
	})
	group.Go(func() error {
		notifications.Run(groupCtx)
		return nil
	})
	return group.Wait()
}
