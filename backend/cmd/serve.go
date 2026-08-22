package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/httpserver"
	logging "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	adminapi "github.com/mmtaee/ocserv-dashboard/backend/services/admin_api"
	customerapi "github.com/mmtaee/ocserv-dashboard/backend/services/customer_api"
	telegrambot "github.com/mmtaee/ocserv-dashboard/backend/services/telegram_bot"
	"github.com/mmtaee/ocserv-dashboard/backend/services/worker"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	debugMode  bool
	serverHost string
	serverPort int
	dockerMode bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run all backend services",
	RunE: func(_ *cobra.Command, _ []string) error {
		_ = godotenv.Load()
		return serve()
	},
}

func init() {
	serveCmd.Flags().BoolVarP(&debugMode, "debug", "d", false, "enable debug mode")
	serveCmd.Flags().StringVar(&serverHost, "host", "0.0.0.0", "server host")
	serveCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "server port")
	serveCmd.Flags().BoolVar(&dockerMode, "docker-mode", false, "use Docker-based ocserv integrations")
	rootCmd.AddCommand(serveCmd)
}

func serve() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	logging.Init(ctx, 256)
	defer func() {
		stop()
		logging.Wait()
	}()

	config.Init(debugMode, serverHost, serverPort)
	cfg := config.Get()
	if err := database.Connect(); err != nil {
		return err
	}
	defer database.Close()

	telegramEnabled := strings.EqualFold(strings.TrimSpace(os.Getenv("TELEGRAM_BOT_ENABLED")), "true")
	admin := adminapi.New(telegramEnabled || cfg.Debug)
	customer := customerapi.New(cfg)
	apiServer := httpserver.New(cfg, admin, customer)
	backgroundWorker := worker.New(dockerMode)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error { return apiServer.Run(groupCtx) })
	group.Go(func() error { return backgroundWorker.Run(groupCtx) })
	if telegramEnabled {
		receiptsDir := strings.TrimSpace(os.Getenv("TELEGRAM_RECEIPTS_DIR"))
		if receiptsDir == "" {
			receiptsDir = filepath.Join("uploads", "receipts")
		}
		botService := telegrambot.New(filepath.Clean(receiptsDir))
		group.Go(func() error { return botService.Run(groupCtx) })
	}

	return group.Wait()
}
