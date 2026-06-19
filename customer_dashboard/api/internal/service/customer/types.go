package customer

import (
	"time"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
)

type ModelCustomer struct {
	Owner                string                 `json:"owner"`
	Username             string                 `json:"username"`
	IsLocked             bool                   `json:"is_locked"`
	CertificateEnabled   bool                   `json:"certificate_enabled"`
	CertificateAvailable bool                   `json:"certificate_available"`
	ExpireAt             *time.Time             `json:"expire_at"`
	DeactivatedAt        *time.Time             `json:"deactivated_at"`
	TrafficType          string                 `json:"traffic_type"`
	TrafficSize          int64                  `json:"traffic_size"`
	Rx                   int                    `json:"rx"`
	Tx                   int                    `json:"tx"`
}

type UsageResponse struct {
	DateStart  time.Time               `json:"date_start" validate:"required"`
	DateEnd    time.Time               `json:"date_end" validate:"required"`
	Bandwidths []models.DailyTraffic   `json:"bandwidths" validate:"required"`
}

type SummaryResponse struct {
	OcservUser ModelCustomer  `json:"ocserv_user" validate:"required"`
	Usage      UsageResponse  `json:"usage" validate:"required"`
}

type CiscoSetupResponse struct {
	CertificateImportURI string                 `json:"certificate_import_uri" validate:"required"`
	ConnectionCreateURI  string                 `json:"connection_create_uri" validate:"required"`
	CertificatePassword  string                 `json:"certificate_password" validate:"required"`
	ConnectionName       string                 `json:"connection_name" validate:"required"`
	ServerAddress        string                 `json:"server_address" validate:"required"`
	ServerPort           int                    `json:"server_port" validate:"required"`
	ExpiresAt            time.Time              `json:"expires_at" validate:"required"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=2,max=32"`
	NewPassword string `json:"new_password" validate:"required,min=2,max=32"`
}

type UserStatisticsRequest struct {
	DateStart string `query:"date_start" validate:"omitempty"`
	DateEnd   string `query:"date_end" validate:"omitempty"`
}

type SessionLogsRequest struct {
	DateStart string `query:"date_start" validate:"omitempty"`
	DateEnd   string `query:"date_end" validate:"omitempty"`
	OrderBy   string `query:"order_by" validate:"omitempty,oneof=id created_at"`
	Sort      string `query:"sort" validate:"omitempty,oneof=asc desc"`
	Page      int    `query:"page" validate:"required,min=1"`
	Limit     int    `query:"limit" validate:"required,min=1,max=100"`
}

type UserStatisticsResponse struct {
	Bandwidths []models.DailyTraffic `json:"bandwidths"`
}

type SessionLogsResponse struct {
	Meta   request.Pagination               `json:"meta"`
	Result []models.OcservUserSessionLog    `json:"result"`
}

type OnlineSessionsResponse struct {
	Result []models.OnlineUserSession `json:"result"`
}
