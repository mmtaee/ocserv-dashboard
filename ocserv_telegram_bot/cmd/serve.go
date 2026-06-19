package cmd

import (
	"github.com/mmtaee/ocserv-dashboard/ocserv_telegram_bot/pkg/bootstrap"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	debug bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start telegram bot service",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Starting telegram bot service...")
		bootstrap.Serve(debug)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
}
