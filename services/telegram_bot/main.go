package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/mmtaee/ocserv-dashboard/common/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/bot"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/notifier"
)

const receiptStorageDir = "/opt/ocserv_dashboard/uploads/receipts"

var debug bool

func main() {
	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	logger.Init(ctx, 100)
	config.Init(debug, "", 0)
	database.Connect()

	if err := os.MkdirAll(filepath.Clean(receiptStorageDir), 0o750); err != nil {
		logger.Warn("failed to create receipt directory %s: %v", receiptStorageDir, err)
	}

	manager := bot.NewManager(receiptStorageDir)
	go manager.Run(ctx)

	notif := notifier.New(manager)
	go notif.Run(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	logger.Warn("Received signal: %s", sig)
	cancel()

	manager.Stop()
	logger.Info("telegram_bot service shutdown complete")
}
