package bootstrap

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/ocserv_user_manager/internal/service"
)

func Serve(debug bool) {
	// Initialize config (we don't need host/port here, just debug)
	config.Init(debug, "", 0)

	// Initialize database
	database.Connect()

	// Initialize cron service
	cronService := service.NewCronService()

	// Check and run missed cron jobs
	logger.Info("Start checking missing cron jobs")
	cronService.MissedCron()
	logger.Info("Checking missing cron jobs completed")

	// Start cron scheduler
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		cronService.UserExpiryCron(ctx)
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Warn("Received signal: %s", sig)
	cancel()

	logger.Info("User manager service shutdown complete")
}
