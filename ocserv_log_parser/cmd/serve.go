package cmd

import (
	"github.com/mmtaee/ocserv-dashboard/ocserv_log_parser/pkg/bootstrap"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	debug      bool
	host       string
	port       int
	dockerMode bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start log parser service",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Starting log parser service...")
		bootstrap.Serve(debug, host, port, dockerMode)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	serveCmd.Flags().StringVarP(&host, "host", "H", "0.0.0.0", "Server host")
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Server port")
	serveCmd.Flags().BoolVar(&dockerMode, "docker-mode", false, "Docker mode")
}
