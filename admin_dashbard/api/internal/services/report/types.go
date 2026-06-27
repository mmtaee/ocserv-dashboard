package report

import (
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/models"
)

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

type OcservUserReportResponse struct {
	Online      int   `json:"online"`
	Active      int64 `json:"active"`
	Deactivated int64 `json:"deactivated"`
	Locked      int64 `json:"locked"`
}
