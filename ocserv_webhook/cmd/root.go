package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ocserv-webhook",
	Short: "Ocserv Webhook Service",
	Long:  "Ocserv Webhook Service handles webhook requests for user management actions",
}

func Execute() error {
	return rootCmd.Execute()
}
