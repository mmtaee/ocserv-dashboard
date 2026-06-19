package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "ocserv-log-parser",
	Short: "Ocserv Log Parser Service",
	Long:  "Ocserv Log Parser parses ocserv logs to track sessions and traffic statistics",
}

func Execute() error {
	return rootCmd.Execute()
}
