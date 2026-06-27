package repository

import (
	"context"
	"time"

	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"gorm.io/gorm"
)

type ReportRepository struct {
	db                   *gorm.DB
	commonOcservUserRepo user.OcservUserInterface
}

type ReportRepositoryInterface interface {
	SessionLogs(ctx context.Context, pagination *request.Pagination, dateStart, dateEnd *time.Time) (*[]models.OcservUserSessionLog, int64, error)
	Statistics(ctx context.Context, dateStart, dateEnd *time.Time) (*[]models.DailyTraffic, error)
	TopBandwidthUser(ctx context.Context) (TopBandwidthUsers, error)
	TotalBandwidth(ctx context.Context) (TotalBandwidths, error)
	TotalUsers(ctx context.Context) (int64, error)
	TotalBandwidthDateRange(ctx context.Context, dateStart, dateEnd *time.Time) (TotalBandwidths, error)
	TotalBandWidthUser(ctx context.Context, uid string) (TotalBandwidths, error)
	TenDaysStats(ctx context.Context) ([]models.DailyTraffic, error)
	UsersStat(ctx context.Context) (UserStatsResult, error)
}

type UserStatsResult struct {
	Active      int64
	Deactivated int64
	Locked      int64
}

func NewtReportRepository() *ReportRepository {
	return &ReportRepository{
		db:                   database.GetConnection(),
		commonOcservUserRepo: user.NewOcservUser(),
	}
}

func (r *ReportRepository) SessionLogs(
	ctx context.Context,
	pagination *request.Pagination,
	dateStart, dateEnd *time.Time,
) (*[]models.OcservUserSessionLog, int64, error) {
	var totalRecords int64

	query := r.db.WithContext(ctx).Model(&models.OcservUserSessionLog{})

	if dateStart != nil {
		query = query.Where("created_at >= ?", *dateStart)
	}

	if dateEnd != nil {
		query = query.Where("created_at < ?", dateEnd.AddDate(0, 0, 1))
	}

	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	var logs []models.OcservUserSessionLog
	if err := request.Paginator(ctx, query, pagination).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return &logs, totalRecords, nil
}

func (r *ReportRepository) Statistics(ctx context.Context, dateStart, dateEnd *time.Time) (*[]models.DailyTraffic, error) {
	var results []models.DailyTraffic
	err := r.db.WithContext(ctx).
		Model(&models.OcservUserTrafficStatistics{}).
		Joins("JOIN ocserv_users ou ON ou.id = ocserv_user_traffic_statistics.oc_user_id").
		Select(`
		DATE(ocserv_user_traffic_statistics.created_at) AS date,
		SUM(ocserv_user_traffic_statistics.rx) / 1073741824.0 AS rx,
		SUM(ocserv_user_traffic_statistics.tx) / 1073741824.0 AS tx
	`).
		Where("ocserv_user_traffic_statistics.created_at >= ?", *dateStart).
		Where("ocserv_user_traffic_statistics.created_at <= ?", *dateEnd).
		Group("DATE(ocserv_user_traffic_statistics.created_at)").
		Order("DATE(ocserv_user_traffic_statistics.created_at)").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}
	return &results, nil
}

func (r *ReportRepository) TotalUsers(ctx context.Context) (int64, error) {
	var totalRecords int64

	err := r.db.WithContext(ctx).Model(&models.OcservUser{}).Count(&totalRecords).Error
	if err != nil {
		return 0, err
	}
	return totalRecords, nil
}

func (r *ReportRepository) TopBandwidthUser(ctx context.Context) (TopBandwidthUsers, error) {
	var (
		topRx []models.OcservUser
		topTx []models.OcservUser
	)

	result := TopBandwidthUsers{}

	// Top RX
	if err := r.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Select("uid, rx, tx, username, created_at").
		Where("rx > 0").
		Order("rx DESC, id DESC").
		Limit(4).
		Find(&topRx).Error; err != nil {
		return result, err
	}
	result.TopRX = topRx

	// Top TX
	if err := r.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Select("uid, rx, tx, username, created_at").
		Where("tx > 0").
		Order("tx DESC, id DESC").
		Limit(4).
		Find(&topTx).Error; err != nil {
		return result, err
	}
	result.TopTX = topTx

	return result, nil
}

func (r *ReportRepository) TotalBandwidth(ctx context.Context) (TotalBandwidths, error) {
	var total TotalBandwidths

	err := r.db.WithContext(ctx).
		Model(&models.OcservUserTrafficStatistics{}).
		Select(`
        COALESCE(SUM(rx),0) / 1073741824.0 AS rx,
        COALESCE(SUM(tx),0) / 1073741824.0 AS tx`).
		Scan(&total).Error
	if err != nil {
		return total, err
	}
	return total, nil
}

func (r *ReportRepository) TotalBandwidthDateRange(ctx context.Context, dateStart, dateEnd *time.Time) (TotalBandwidths, error) {
	var total TotalBandwidths

	query := r.db.WithContext(ctx).
		Model(&models.OcservUserTrafficStatistics{}).
		Select(`
			COALESCE(SUM(rx),0) / 1073741824.0 AS rx,
			COALESCE(SUM(tx),0) / 1073741824.0 AS tx`)

	// Apply filters based on dateStart and dateEnd
	if dateStart != nil {
		query = query.Where("created_at >= ?", *dateStart)
	}
	if dateEnd != nil {
		query = query.Where("created_at <= ?", *dateEnd)
	}

	err := query.Scan(&total).Error
	if err != nil {
		return total, err
	}
	return total, nil
}

func (r *ReportRepository) TenDaysStats(ctx context.Context) ([]models.DailyTraffic, error) {
	var results []models.DailyTraffic

	start := time.Now().AddDate(0, 0, -10).Truncate(24 * time.Hour)

	err := r.db.WithContext(ctx).
		Model(&models.OcservUserTrafficStatistics{}).
		Select(`
		DATE(created_at) AS date,
		SUM(rx) / 1073741824.0 AS rx,
		SUM(tx) / 1073741824.0 AS tx`).
		Where("created_at >= ?", start).
		Group("DATE(created_at)").
		Order("DATE(created_at)").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *ReportRepository) UsersStat(ctx context.Context) (UserStatsResult, error) {
	var result UserStatsResult

	err := r.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Select(`
			COUNT(*) FILTER (WHERE deactivated_at IS NULL AND is_locked = false) AS active,
			COUNT(*) FILTER (WHERE deactivated_at IS NOT NULL) AS deactivated,
			COUNT(*) FILTER (WHERE is_locked = true) AS locked
		`).
		Scan(&result).Error

	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *ReportRepository) TotalBandWidthUser(ctx context.Context, uid string) (TotalBandwidths, error) {
	var total TotalBandwidths

	//err := o.db.WithContext(ctx).
	//	Model(&models.OcservUserTrafficStatistics{}).
	//	Joins("JOIN ocserv_users ou ON ou.id = ocserv_user_traffic_statistics.oc_user_id").
	//	Where("ou.uid = ?", uid).
	//	Select(`
	//    COALESCE(SUM(rx),0) / 1073741824.0 AS rx,
	//    COALESCE(SUM(tx),0) / 1073741824.0 AS tx`).
	//	Scan(&total).Error

	err := r.db.WithContext(ctx).
		Table("ocserv_user_traffic_statistics AS t").
		Joins("JOIN ocserv_users ou ON ou.id = t.oc_user_id").
		Where("ou.uid = ?", uid).
		Select(`
            COALESCE(SUM(t.rx),0) / 1073741824.0 AS rx,
            COALESCE(SUM(t.tx),0) / 1073741824.0 AS tx
        `).
		Scan(&total).Error

	if err != nil {
		return total, err
	}
	return total, nil
}
