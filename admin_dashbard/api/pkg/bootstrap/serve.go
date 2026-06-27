package bootstrap

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
)

func Serve(debug bool, host string, port int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Init(ctx, 100)

	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic recovered: %v", r)
		}
	}()

	config.Init(debug, host, port)
	cfg := config.Get()

	database.Connect()
	defer database.Close()

	go routing.Serve(cfg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Warn("Forcing shutting down...")
		os.Exit(1)
	}()

	sig := <-quit

	logger.Warn("Shutting down... Signal Reason: %s", sig.String())

	routing.Shutdown(ctx)
	database.Close()

	logger.Info("Api service shutdown complete")
}
