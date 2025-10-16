package repository

import (
	"context"
	"github.com/mmtaee/ocserv-users-management/api/internal/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"gorm.io/gorm"
)

type SystemRepository struct {
	db *gorm.DB
}

type SystemRepositoryInterface interface {
	SystemSetup(ctx context.Context, user *models.User, system *models.System) (*models.User, *models.System, error)
	System(ctx context.Context) (*models.System, error)
	SystemUpdate(ctx context.Context, system *models.System) (*models.System, error)
}

func NewSystemRepository() *SystemRepository {
	return &SystemRepository{
		db: database.GetConnection(),
	}
}

func (s *SystemRepository) SystemSetup(ctx context.Context, user *models.User, system *models.System) (*models.User, *models.System, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&system).Error
		if err != nil {
			return err
		}
		err = tx.Create(&user).Error
		if err != nil {
			return err
		}
		return nil
	})
	return user, system, err
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
				"google_captcha_secret_key": system.GoogleCaptchaSecretKey,
				"google_captcha_site_key":   system.GoogleCaptchaSiteKey,
			},
		).Error; err != nil {
		return nil, err
	}

	return system, nil
}
