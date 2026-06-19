package repository

import (
	"time"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"gorm.io/gorm"
)

type ReportRepository interface {
	SessionLogs(page, limit int, orderBy, sort string, startDate, endDate *time.Time) ([]models.OcservUserSessionLog, int64, error)
	Statistics(startDate, endDate *time.Time) ([]models.DailyTraffic, error)
	TotalBandwidthDateRange(startDate, endDate *time.Time) (TotalBandwidths, error)
	UsersStat(adminID uint, role string) (UserStatsResult, error)
}

type reportRepository struct {
	db *gorm.DB
}

type UserStatsResult struct {
	Active      int64
	Deactivated int64
	Locked      int64
}

type TotalBandwidths struct {
	Rx float64 `json:"rx"` // in GiB
	Tx float64 `json:"tx"` // in GiB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) SessionLogs(page, limit int, orderBy, sort string, startDate, endDate *time.Time) ([]models.OcservUserSessionLog, int64, error) {
	var logs []models.OcservUserSessionLog
	var total int64

	query := r.db.Model(&models.OcservUserSessionLog{})

	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at < ?", endDate.AddDate(0, 0, 1))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	validOrderFields := map[string]bool{"id": true, "created_at": true}
	validSortOrders := map[string]bool{"asc": true, "desc": true}
	if !validOrderFields[orderBy] {
		orderBy = "id"
	}
	if !validSortOrders[sort] {
		sort = "desc"
	}

	offset := (page - 1) * limit
	err := query.Order(orderBy + " " + sort).Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *reportRepository) Statistics(startDate, endDate *time.Time) ([]models.DailyTraffic, error) {
	var results []models.DailyTraffic
	err := r.db.
		Model(&models.OcservUserTrafficStatistics{}).
		Select(`
		DATE(ocserv_user_traffic_statistics.created_at) AS date,
		SUM(ocserv_user_traffic_statistics.rx) / 1073741824.0 AS rx,
		SUM(ocserv_user_traffic_statistics.tx) / 1073741824.0 AS tx
	`).
		Where("ocserv_user_traffic_statistics.created_at >= ?", *startDate).
		Where("ocserv_user_traffic_statistics.created_at <= ?", *endDate).
		Group("DATE(ocserv_user_traffic_statistics.created_at)").
		Order("DATE(ocserv_user_traffic_statistics.created_at)").
		Scan(&results).Error
	return results, err
}

func (r *reportRepository) TotalBandwidthDateRange(startDate, endDate *time.Time) (TotalBandwidths, error) {
	var total TotalBandwidths
	query := r.db.
		Model(&models.OcservUserTrafficStatistics{}).
		Select(`
		COALESCE(SUM(rx),0) / 1073741824.0 AS rx,
		COALESCE(SUM(tx),0) / 1073741824.0 AS tx`)

	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}

	err := query.Scan(&total).Error
	return total, err
}

func (r *reportRepository) UsersStat(adminID uint, role string) (UserStatsResult, error) {
	var result UserStatsResult

	query := r.db.Model(&models.OcservUser{})

	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}

	err := query.
		Select(`
			COUNT(*) FILTER (WHERE deactivated_at IS NULL AND is_locked = false) AS active,
			COUNT(*) FILTER (WHERE deactivated_at IS NOT NULL) AS deactivated,
			COUNT(*) FILTER (WHERE is_locked = true) AS locked
		`).
		Scan(&result).Error

	return result, err
}
