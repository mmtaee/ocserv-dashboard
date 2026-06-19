package cmd

import (
	"errors"
	"log"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/internal/repository"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	username string
	password string
)

var createSuperAdminCmd = &cobra.Command{
	Use:   "create-super-admin",
	Short: "Create a new super admin user",
	Run: func(cmd *cobra.Command, args []string) {
		// Connect to database
		database.Connect()
		defer database.Close()

		db := database.GetConnection()
		adminRepo := repository.NewAdminRepository(db)

		// Check if super admin already exists
		var existingAdmin models.Administrator
		err := db.Where("role = ?", models.AdminRoleSuper).First(&existingAdmin).Error
		if err == nil {
			log.Fatal("Super admin already exists")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatalf("Failed to check for existing super admin: %v", err)
		}

		// Check if username already exists
		_, err = adminRepo.FindByUsername(username)
		if err == nil {
			log.Fatal("Username already exists")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatalf("Failed to check username: %v", err)
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		// Create admin
		admin := &models.Administrator{
			Username: username,
			Password: string(hashedPassword),
			Role:     models.AdminRoleSuper,
		}

		if err := adminRepo.Create(admin); err != nil {
			log.Fatalf("Failed to create super admin: %v", err)
		}

		log.Println("Super admin created successfully")
	},
}

func init() {
	createSuperAdminCmd.Flags().StringVarP(&username, "username", "u", "", "Username for the super admin (required)")
	createSuperAdminCmd.Flags().StringVarP(&password, "password", "p", "", "Password for the super admin (required)")
	createSuperAdminCmd.MarkFlagRequired("username")
	createSuperAdminCmd.MarkFlagRequired("password")
	rootCmd.AddCommand(createSuperAdminCmd)
}
