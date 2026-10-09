package repository

import (
	"context"
	"errors"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"gorm.io/gorm"
)

type WorkerStatsRepository struct {
	db *gorm.DB
}

func NewWorkerStatsRepository(db *gorm.DB) *WorkerStatsRepository {
	return &WorkerStatsRepository{db: db}
}

func (r *WorkerStatsRepository) FindUser(ctx context.Context, username string) (*models.OcservUser, error) {
	var user models.OcservUser
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *WorkerStatsRepository) RecordUsage(ctx context.Context, traffic *models.OcservUserTrafficStatistics) (*models.OcservUser, error) {
	var user models.OcservUser
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(traffic).Error; err != nil {
			return err
		}
		result := tx.Raw(`
			UPDATE ocserv_users
			SET running_rx = running_rx + ?,
				running_tx = running_tx + ?,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
			RETURNING *
		`, traffic.Rx, traffic.Tx, traffic.OcservUserID).Scan(&user)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *WorkerStatsRepository) CurrentMonthTotals(ctx context.Context, userID uint, usageResetAt *time.Time) (int, int, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	if usageResetAt != nil && usageResetAt.After(start) {
		start = *usageResetAt
	}
	var totals struct {
		TotalRX int
		TotalTX int
	}
	err := r.db.WithContext(ctx).Model(&models.OcservUserTrafficStatistics{}).
		Select("COALESCE(SUM(rx), 0) AS total_rx, COALESCE(SUM(tx), 0) AS total_tx").
		Where("ocserv_user_id = ? AND created_at >= ? AND created_at < ?", userID, start, start.AddDate(0, 1, 0)).
		Scan(&totals).Error
	return totals.TotalRX, totals.TotalTX, err
}

func (r *WorkerStatsRepository) UpdateAccessState(ctx context.Context, userID uint, locked bool, deactivatedAt *time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{"is_locked": locked, "deactivated_at": deactivatedAt})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("ocserv user not found")
	}
	return nil
}

func (r *WorkerStatsRepository) SaveSessionLog(ctx context.Context, log *models.OcservUserSessionLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}
