package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "ocserv-telegram-bot",
	Short: "Ocserv Telegram Bot Service",
	Long:  "Ocserv Telegram Bot provides a self-service interface for VPN users",
}

func Execute() error {
	return rootCmd.Execute()
}
