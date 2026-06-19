package bootstrap

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/ocserv_log_parser/internal/readers"
	"github.com/mmtaee/ocserv-dashboard/ocserv_log_parser/internal/stats"
)

func Serve(debug bool, host string, port int, dockerMode bool) {
	// Initialize config
	config.Init(debug, host, port)

	// Initialize database
	database.Connect()

	ctx, cancel := context.WithCancel(context.Background())
	serviceName := "ocserv"

	streamChan := make(chan string, 1000)

	// Start log reader
	if !dockerMode {
		logger.Info("Systemd mode")
		go func() {
			if err := readers.SystemdStreamLogs(ctx, serviceName, streamChan); err != nil {
				logger.Error("Systemd stream logs error: %v", err)
			}
		}()
	} else {
		logger.Info("Docker mode")
		go func() {
			if err := readers.DockerStreamLogs(ctx, serviceName, streamChan); err != nil {
				logger.Error("Docker stream logs error: %v", err)
			}
		}()
	}

	// Start stat service
	statService := stats.NewStatService(ctx, streamChan)
	go func() {
		statService.CalculateUserStats()
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Warn("Received shutdown signal %s", sig)
	cancel()

	<-ctx.Done()
	logger.Info("Log parser service shutdown successfully")
}
