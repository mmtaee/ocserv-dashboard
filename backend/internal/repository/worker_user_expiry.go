package repository

import (
	"context"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"gorm.io/gorm"
)

type WorkerUserExpiryRepository struct {
	db *gorm.DB
}

func NewWorkerUserExpiryRepository(db *gorm.DB) *WorkerUserExpiryRepository {
	return &WorkerUserExpiryRepository{db: db}
}

func (r *WorkerUserExpiryRepository) ExpiredUsers(ctx context.Context, before time.Time) ([]models.OcservUser, error) {
	var users []models.OcservUser
	err := r.db.WithContext(ctx).Select("id", "username", "expire_at").
		Where("expire_at IS NOT NULL").Where("deactivated_at IS NULL").Where("expire_at < ?", before).
		Find(&users).Error
	return users, err
}

func (r *WorkerUserExpiryRepository) Deactivate(ctx context.Context, id uint, at time.Time) error {
	return r.db.WithContext(ctx).Model(&models.OcservUser{}).Where("id = ?", id).
		Updates(map[string]interface{}{"deactivated_at": at, "is_locked": true}).Error
}

func (r *WorkerUserExpiryRepository) MonthlyUsers(ctx context.Context, today time.Time) ([]models.OcservUser, error) {
	var users []models.OcservUser
	err := r.db.WithContext(ctx).Where("(expire_at IS NULL OR expire_at > ?)", today).
		Where("deactivated_at IS NOT NULL").
		Where("traffic_type IN ?", []string{models.MonthlyReceive, models.MonthlyTransmit, models.MonthlyRxTx}).
		Find(&users).Error
	return users, err
}

func (r *WorkerUserExpiryRepository) Reactivate(ctx context.Context, id uint, at time.Time) error {
	return r.db.WithContext(ctx).Model(&models.OcservUser{}).Where("id = ?", id).
		Updates(map[string]interface{}{"rx": 0, "tx": 0, "usage_reset_at": &at, "deactivated_at": nil, "is_locked": false}).Error
}

func (r *WorkerUserExpiryRepository) SystemSettings(ctx context.Context) (*models.System, error) {
	var settings models.System
	if err := r.db.WithContext(ctx).First(&settings).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *WorkerUserExpiryRepository) DeleteExpired(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("expire_at IS NOT NULL AND expire_at <= ?", cutoff).Delete(&models.OcservUser{})
	return result.RowsAffected, result.Error
}
