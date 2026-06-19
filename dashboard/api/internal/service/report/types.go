package report

import (
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/request"
)

type SessionLogsData struct {
	DateStart string `json:"date_start" query:"date_start" validate:"omitempty" example:"2025-01-31"`
	DateEnd   string `json:"date_end" query:"date_end" validate:"omitempty" example:"2025-12-31"`
}

type SessionLogsResponse struct {
	Meta   request.Pagination            `json:"meta"`
	Result []models.OcservUserSessionLog `json:"result"`
}

type StatisticsData struct {
	DateStart string `json:"date_start" query:"date_start" validate:"required" example:"2025-01-31"`
	DateEnd   string `json:"date_end" query:"date_end" validate:"required" example:"2025-12-31"`
}

type TotalBandwidthData struct {
	DateStart string `json:"date_start" query:"date_start" validate:"omitempty" example:"2025-01-31"`
	DateEnd   string `json:"date_end" query:"date_end" validate:"omitempty" example:"2025-12-31"`
}

type OcservUserReportResponse struct {
	Online      int   `json:"online"`
	Active      int64 `json:"active"`
	Deactivated int64 `json:"deactivated"`
	Locked      int64 `json:"locked"`
}
