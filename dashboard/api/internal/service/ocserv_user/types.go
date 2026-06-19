package ocserv_user

import (
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/request"
)

type CreateOcservUserRequest struct {
	Group       string                   `json:"group" validate:"required"`
	Username    string                   `json:"username" validate:"required,min=2,max=32"`
	Password    string                   `json:"password" validate:"required,min=2,max=32"`
	ExpireAt    string                   `json:"expire_at" validate:"omitempty" example:"2025-12-31"`
	Unlimited   bool                     `json:"unlimited" validate:"omitempty" example:"false" default:"false"`
	TrafficType string                   `json:"traffic_type" validate:"required,oneof=Free MonthlyTransmit MonthlyReceive TotallyTransmit TotallyReceive"`
	TrafficSize int                      `json:"traffic_size" validate:"omitempty,gte=0" example:"10"` // in GiB
	Description string                   `json:"description" validate:"omitempty,max=1024" example:"User for testing VPN access"`
	Config      *models.OcservUserConfig `json:"config" validate:"required"`
}

type UpdateOcservUserRequest struct {
	Group       *string                  `json:"group" example:"defaults"`
	Password    *string                  `json:"password" validate:"min=2,max=32"`
	ExpireAt    *string                  `json:"expire_at" validate:"omitempty" example:"2025-12-31"`
	Unlimited   bool                     `json:"unlimited" validate:"omitempty" example:"false" default:"false"`
	TrafficType *string                  `json:"traffic_type" validate:"oneof=Free MonthlyTransmit MonthlyReceive TotallyTransmit TotallyReceive"`
	TrafficSize *int                     `json:"traffic_size" validate:"gte=0" example:"10"` // in GiB
	Description *string                  `json:"description" validate:"omitempty,max=1024" example:"User for testing VPN access"`
	Config      *models.OcservUserConfig `json:"config" validate:"omitempty"`
}

type OcservUsersResponse struct {
	Meta   request.Pagination  `json:"meta"`
	Result []models.OcservUser `json:"result"`
}

type GetOnlineSessionsRequest struct {
	Usernames []string `json:"usernames" validate:"required,min=1"`
}

type Ocpasswd struct {
	Username string `json:"username"`
	Group    string `json:"group"`
}

type SyncOcpasswdRequest struct {
	Users       []Ocpasswd               `json:"users" validate:"required"`
	ExpireAt    *string                  `json:"expire_at" validate:"omitempty" example:"2025-12-31"`
	TrafficType *string                  `json:"traffic_type" validate:"required,oneof=Free MonthlyTransmit MonthlyReceive TotallyTransmit TotallyReceive"`
	TrafficSize *int                     `json:"traffic_size" validate:"required,gte=0" example:"10"` // in GiB
	Description *string                  `json:"description" validate:"omitempty,max=1024" example:"User for testing VPN access"`
	Config      *models.OcservUserConfig `json:"config" validate:"omitempty"`
}

type OcservUsersSyncResponse struct {
	Meta   request.Pagination `json:"meta"`
	Result []Ocpasswd         `json:"result"`
}

type ActivateUserData struct {
	ExpireAt *string `json:"expire_at" validate:"omitempty" example:"2025-12-31"`
}

type SessionLogsData struct {
	DateStart string `json:"date_start" query:"date_start" validate:"omitempty" example:"2025-01-31"`
	DateEnd   string `json:"date_end" query:"date_end" validate:"omitempty" example:"2025-12-31"`
}

type SessionLogsResponse struct {
	Meta   request.Pagination            `json:"meta"`
	Result []models.OcservUserSessionLog `json:"result"`
}

type StatisticsData struct {
	DateStart string `json:"date_start" query:"date_start" validate:"omitempty" example:"2025-01-31"`
	DateEnd   string `json:"date_end" query:"date_end" validate:"omitempty" example:"2025-12-31"`
}

type StatisticsResponse struct {
	Statistics []models.DailyTraffic `json:"statistics" validate:"required"`
	TotalBandwidths TotalBandwidths `json:"total_bandwidths" validate:"required"`
}

type TotalBandwidths struct {
	Rx float64 `json:"rx"` // in GiB
	Tx float64 `json:"tx"` // in GiB
}
