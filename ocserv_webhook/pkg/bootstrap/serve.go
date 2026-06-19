package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/ocserv_webhook/internal/webhook"
)

func Serve(debug bool, host string, port int, dockerMode bool) {
	// Initialize config
	config.Init(debug, host, port)
	cfg := config.Get()

	// Initialize Echo
	e := echo.New()

	// Initialize webhook handler
	webhookHandler := webhook.NewHandler(dockerMode)
	webhookHandler.RegisterRoutes(e)

	// Start server
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		logger.Info("Webhook server listening on %s", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown server
	logger.Warn("Shutting down webhook server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logger.Fatal("Server shutdown failed: %v", err)
	}
	logger.Info("Webhook server shutdown successfully")
}
