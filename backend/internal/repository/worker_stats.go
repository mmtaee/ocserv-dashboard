package repository

import (
	"context"
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

func (r *WorkerStatsRepository) CreateTraffic(ctx context.Context, traffic *models.OcservUserTrafficStatistics) error {
	return r.db.WithContext(ctx).Create(traffic).Error
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

func (r *WorkerStatsRepository) SaveUser(ctx context.Context, user *models.OcservUser) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *WorkerStatsRepository) SaveSessionLog(ctx context.Context, log *models.OcservUserSessionLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}
