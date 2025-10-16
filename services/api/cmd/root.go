package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "web_api",
	Short: "Ocserv User Management API Service CLI",
	Long: `Ocserv User Management API Service CLI

This CLI provides tools to manage the Ocserv backend services, including:
  - Running the HTTP server
  - Managing admin users
  - Performing database operations
  - Other system-level tasks`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
