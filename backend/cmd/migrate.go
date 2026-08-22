package cmd

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	logging "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/bootstrap"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	RunE: func(_ *cobra.Command, _ []string) error {
		_ = godotenv.Load()
		ctx, cancel := context.WithCancel(context.Background())
		logging.Init(ctx, 128)
		defer func() {
			cancel()
			logging.Wait()
		}()
		config.Init(false, "", 0)
		if err := database.Connect(); err != nil {
			return err
		}
		defer database.Close()
		return bootstrap.Migrate(database.GetConnection())
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
