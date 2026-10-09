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

// RecordFirstConnection atomically claims the first observed connection. The
// predicate prevents concurrent or later log events from extending the term.
func (r *WorkerUserExpiryRepository) RecordFirstConnection(ctx context.Context, username string, connectedAt time.Time) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Where("username = ?", username).
		Where("expiry_mode = ?", models.ExpiryModeFirstConnection).
		Where("first_connected_at IS NULL AND expire_at IS NULL").
		Updates(map[string]interface{}{
			"first_connected_at": connectedAt.UTC(),
			"expire_at": gorm.Expr(
				"? + (expire_days_after_first_connection * INTERVAL '1 day')",
				connectedAt.UTC(),
			),
		})
	return result.RowsAffected == 1, result.Error
}

func (r *WorkerUserExpiryRepository) ExpiredUsers(ctx context.Context, at time.Time) ([]models.OcservUser, error) {
	var users []models.OcservUser
	err := r.db.WithContext(ctx).Select("id", "username", "expiry_mode", "expire_at", "first_connected_at").
		Where("expire_at IS NOT NULL").Where("deactivated_at IS NULL").Where("expire_at <= ?", at.UTC()).
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
		Where("traffic_type IN ?", []models.TrafficType{models.MonthlyReceive, models.MonthlyTransmit, models.MonthlyRxTx}).
		Find(&users).Error
	return users, err
}

func (r *WorkerUserExpiryRepository) Reactivate(ctx context.Context, id uint, at time.Time) error {
	return r.db.WithContext(ctx).Model(&models.OcservUser{}).Where("id = ?", id).
		Updates(map[string]interface{}{"running_rx": 0, "running_tx": 0, "usage_reset_at": &at, "deactivated_at": nil, "is_locked": false}).Error
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
