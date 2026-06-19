package cmd

import (
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/pkg/bootstrap"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start API server",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Starting API server...")
		bootstrap.Serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
