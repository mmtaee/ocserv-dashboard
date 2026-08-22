package repository

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	"gorm.io/gorm"
)

type BackupRepository struct {
	db *gorm.DB
}

type BackupRepositoryInterface interface {
	Groups(ctx context.Context) ([]models.OcservGroup, error)
	Users(ctx context.Context) ([]models.OcservUser, error)
	ExistingGroupNames(ctx context.Context, names []string) ([]string, error)
	ExistingUsernames(ctx context.Context, usernames []string) ([]string, error)
}

func NewBackupRepository() *BackupRepository {
	return &BackupRepository{db: database.GetConnection()}
}

func (r *BackupRepository) Groups(ctx context.Context) ([]models.OcservGroup, error) {
	var groups []models.OcservGroup
	return groups, r.db.WithContext(ctx).Find(&groups).Error
}

func (r *BackupRepository) Users(ctx context.Context) ([]models.OcservUser, error) {
	var users []models.OcservUser
	return users, r.db.WithContext(ctx).Find(&users).Error
}

func (r *BackupRepository) ExistingGroupNames(ctx context.Context, names []string) ([]string, error) {
	var existing []string
	err := r.db.WithContext(ctx).Model(&models.OcservGroup{}).Where("name IN ?", names).Pluck("name", &existing).Error
	return existing, err
}

func (r *BackupRepository) ExistingUsernames(ctx context.Context, usernames []string) ([]string, error) {
	var existing []string
	err := r.db.WithContext(ctx).Model(&models.OcservUser{}).Where("username IN ?", usernames).Pluck("username", &existing).Error
	return existing, err
}
