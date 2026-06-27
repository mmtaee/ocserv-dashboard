package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/bot"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/i18n"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/notifier"
)

const defaultReceiptsDir = "/opt/ocserv_dashboard/uploads/receipts"

func receiptStorageDir() string {
	if d := strings.TrimSpace(os.Getenv("TELEGRAM_RECEIPTS_DIR")); d != "" {
		return filepath.Clean(d)
	}
	return defaultReceiptsDir
}

var debug bool

func main() {
	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	logger.Init(ctx, 100)
	config.Init(debug, "", 0)
	i18n.Init()
	database.Connect()

	dir := receiptStorageDir()
	if err := os.MkdirAll(dir, 0o750); err != nil {
		logger.Warn("failed to create receipt directory %s: %v", dir, err)
	}

	manager := bot.NewManager(dir)
	go manager.Run(ctx)

	notif := notifier.New(manager, manager.Repo())
	go notif.Run(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	logger.Warn("Received signal: %s", sig)
	cancel()

	manager.Stop()
	logger.Info("telegram_bot service shutdown complete")
}
