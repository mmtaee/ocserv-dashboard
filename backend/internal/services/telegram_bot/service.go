package telegrambot

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/services/telegram_bot/bot"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/telegram_bot/auth"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/telegram_bot/conversation/session"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/telegram_bot/i18n"
	notifications "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/telegram_bot/notifications"
	"golang.org/x/sync/errgroup"
)

type Component interface {
	Run(ctx context.Context) error
}

type Service struct {
	receiptsDir string
	components  []Component
}

func New(receiptsDir string) *Service {
	repo := repository.NewTelegramBotRepository()
	sessions := session.NewStore(15 * time.Minute)
	verifier := auth.NewVerifier(repo)
	manager := bot.NewManager(receiptsDir, repo, repo, sessions, verifier)
	return &Service{receiptsDir: receiptsDir, components: []Component{manager, notifications.New(manager, repo)}}
}

func (s *Service) Run(ctx context.Context) error {
	if err := os.MkdirAll(s.receiptsDir, 0o750); err != nil {
		return fmt.Errorf("create telegram receipt directory: %w", err)
	}
	i18n.Init()
	group, groupCtx := errgroup.WithContext(ctx)
	for _, component := range s.components {
		component := component
		group.Go(func() error { return component.Run(groupCtx) })
	}
	return group.Wait()
}
