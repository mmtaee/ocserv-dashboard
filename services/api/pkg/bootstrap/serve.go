package bootstrap

import (
	"context"
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing"
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	Migrate()

	defer database.CloseConnection()

	go routing.Serve(cfg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Warn("Forcing shutting down...")
		os.Exit(1)
	}()

	sig := <-quit

	logger.Warn("Shutting down... Reason:", sig)

	routing.Shutdown(ctx)
	database.CloseConnection()

	logger.Info("Api service shutdown complete")
}
