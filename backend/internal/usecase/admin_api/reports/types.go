package report

import (
	"context"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Repository interface {
	SessionLogs(ctx context.Context, pagination *request.Pagination, dateStart, dateEnd *time.Time) (*[]models.OcservUserSessionLog, int64, error)
	Statistics(ctx context.Context, dateStart, dateEnd *time.Time) (*[]models.DailyTraffic, error)
	TotalBandwidthDateRange(ctx context.Context, dateStart, dateEnd *time.Time) (repository.TotalBandwidths, error)
	UsersStat(ctx context.Context) (repository.UserStatsResult, error)
	TotalBandWidthUser(ctx context.Context, id uint) (repository.TotalBandwidths, error)
	TenDaysStats(ctx context.Context) ([]models.DailyTraffic, error)
	TotalUsers(ctx context.Context) (int64, error)
	TopBandwidthUser(ctx context.Context) (repository.TopBandwidthUsers, error)
	TotalBandwidth(ctx context.Context) (repository.TotalBandwidths, error)
}

type SessionProvider interface {
	OnlineSessions() ([]models.OnlineUserSession, error)
}

type UserReport struct {
	Online      int   `json:"online"`
	Active      int64 `json:"active"`
	Deactivated int64 `json:"deactivated"`
	Locked      int64 `json:"locked"`
}

type SessionLogsData struct {
	DateStart string `json:"date_start" query:"date_start" validate:"omitempty" example:"2025-1-31"`
	DateEnd   string `json:"date_end" query:"date_end" validate:"omitempty" example:"2025-12-31"`
}

type SessionLogsResponse struct {
	Meta   request.Meta                   `json:"meta" validate:"required"`
	Result *[]models.OcservUserSessionLog `json:"result" validate:"omitempty"`
}

type StatisticsData struct {
	DateStart string `json:"date_start" query:"date_start" validate:"omitempty" example:"2025-1-31"`
	DateEnd   string `json:"date_end" query:"date_end" validate:"omitempty" example:"2025-12-31"`
}

type StatisticsResponse struct {
	Statistics      []models.DailyTraffic      `json:"statistics" validate:"required"`
	TotalBandwidths repository.TotalBandwidths `json:"total_bandwidths" validate:"required"`
}

type TotalBandwidthData struct {
	DateStart string `json:"date_start" query:"date_start" validate:"omitempty" example:"2025-1-31"`
	DateEnd   string `json:"date_end" query:"date_end" validate:"omitempty" example:"2025-12-31"`
}

type OcservUserReportResponse = UserReport
