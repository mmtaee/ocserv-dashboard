package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/user_expiry/internal/service"
)

var (
	debug      bool
	dockerMode bool
)

func main() {
	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.BoolVar(&dockerMode, "docker-mode", true, "Docker Mode")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	logger.Init(ctx, 100)

	config.Init(debug, "", 8888)
	database.Connect()

	cronService := service.NewCornService(dockerMode)

	logger.Info("Start checking missing cron jobs")
	cronService.MissedCron()
	logger.Info("Checking missing cron jobs completed")

	go func() {
		cronService.UserExpiryCron(ctx)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Warn("Received signal: %s ", sig)
	cancel()

	logger.Info("User expiry service shutting down completed")
}
