package bootstrap

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/ocserv_telegram_bot/internal/bot"
	"github.com/mmtaee/ocserv-dashboard/ocserv_telegram_bot/internal/notifier"
)

func Serve(receiptsDir string) error {
	if receiptsDir == "" {
		receiptsDir = "./telegram-receipts"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if err := database.Connect(); err != nil {
		logger.Fatal("failed to connect to database: %v", err)
	}
	defer database.Close()

	mgr := bot.NewManager(receiptsDir)
	nfy := notifier.New(mgr, mgr.Repo())

	go mgr.Run(ctx)
	go nfy.Run(ctx)

	logger.Info("ocserv_telegram_bot: started, receipts dir: %s", receiptsDir)

	<-sigChan
	logger.Info("ocserv_telegram_bot: shutting down...")
	cancel()
	return nil
}
