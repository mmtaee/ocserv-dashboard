package repository

import (
	"context"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	"gorm.io/gorm"
)

type SystemRepository struct {
	db *gorm.DB
}

type SystemRepositoryInterface interface {
	System(ctx context.Context) (*models.System, error)
	SystemUpdate(ctx context.Context, system *models.System) (*models.System, error)
}

func NewSystemRepository() *SystemRepository {
	return &SystemRepository{
		db: database.GetConnection(),
	}
}

func (s *SystemRepository) System(ctx context.Context) (*models.System, error) {
	var system models.System
	err := s.db.WithContext(ctx).First(&system).Error
	if err != nil {
		return nil, err
	}
	return &system, nil
}

func (s *SystemRepository) SystemUpdate(ctx context.Context, system *models.System) (*models.System, error) {
	var latest models.System
	if err := s.db.WithContext(ctx).Order("id desc").First(&latest).Error; err != nil {
		return nil, err
	}

	// Update the latest system record with new values
	if err := s.db.WithContext(ctx).
		Model(&models.System{}).
		Where("id = ?", latest.ID).
		Updates(
			map[string]interface{}{
				"first_init":                     system.FirstInit,
				"google_captcha_secret_key":      system.GoogleCaptchaSecretKey,
				"google_captcha_site_key":        system.GoogleCaptchaSiteKey,
				"auto_delete_inactive_users":     system.AutoDeleteInactiveUsers,
				"keep_inactive_user_days":        system.KeepInactiveUserDays,
				"client_profile_server_address":  system.ClientProfileServerAddress,
				"client_profile_server_port":     system.ClientProfileServerPort,
				"client_profile_connection_name": system.ClientProfileConnectionName,
			},
		).Error; err != nil {
		return nil, err
	}

	return system, nil
}
