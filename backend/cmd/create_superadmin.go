package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	logging "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/bootstrap/superadmin"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
	"github.com/spf13/cobra"
)

var createSuperadminCmd = &cobra.Command{
	Use:   "create-superadmin",
	Short: "Create or update the configured initial superadmin",
	RunE: func(_ *cobra.Command, _ []string) error {
		_ = godotenv.Load()
		username := os.Getenv("SUPERADMIN_USERNAME")
		password := os.Getenv("SUPERADMIN_PASSWORD")
		if username == "" || password == "" {
			return errors.New("SUPERADMIN_USERNAME and SUPERADMIN_PASSWORD are required")
		}
		ctx, cancel := context.WithCancel(context.Background())
		logging.Init(ctx, 128)
		defer func() { cancel(); logging.Wait() }()
		config.Init(false, "", 0)
		if err := database.Connect(); err != nil {
			return err
		}
		defer database.Close()
		_, err := superadmin.New(repository.NewUserRepository(), crypto.NewCustomPassword()).Ensure(ctx, username, password)
		return err
	},
}

func init() {
	rootCmd.AddCommand(createSuperadminCmd)
}
