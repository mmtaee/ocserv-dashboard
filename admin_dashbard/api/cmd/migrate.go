package cmd

import (
	"github.com/mmtaee/ocserv-dashboard/api/pkg/bootstrap"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run the migrate database schemas",
	Run: func(cmd *cobra.Command, args []string) {
		bootstrap.Migrate()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
