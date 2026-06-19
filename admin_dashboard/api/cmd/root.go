package cmd

import (
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "api",
	Short: "Ocserv Dashboard API",
}

func Execute() error {
	config.Init()
	return rootCmd.Execute()
}
