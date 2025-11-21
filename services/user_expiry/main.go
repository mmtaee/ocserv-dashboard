package main

import (
	"context"
	"flag"
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"github.com/mmtaee/ocserv-users-management/user_expiry/internal/service"
	"os"
	"os/signal"
	"syscall"
)

var debug bool

func main() {
	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	logger.Init(ctx, 100)

	config.Init(debug, "", 8888)
	database.Connect()

	cronService := service.NewCornService()

	cronService.MissedCron()

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
