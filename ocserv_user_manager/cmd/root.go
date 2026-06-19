package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "ocserv-user-manager",
	Short: "Ocserv User Manager Service",
	Long:  "Ocserv User Manager handles user expiration, activation, and auto-deletion",
}

func Execute() error {
	return rootCmd.Execute()
}
