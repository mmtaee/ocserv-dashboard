package ocserv_user

import (
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	"github.com/mmtaee/ocserv-dashboard/common/ocserv/user"
)

type CreateOcservUserData struct {
	Group       string                   `json:"group" validate:"required"`
	Username    string                   `json:"username" validate:"required,min=2,max=32"`
	Password    string                   `json:"password" validate:"required,min=2,max=32"`
	ExpireAt    string                   `json:"expire_at" validate:"omitempty" example:"2025-12-31"`
	Unlimited   bool                     `json:"unlimited" validate:"omitempty" example:"false" default:"false"`
	TrafficType string                   `json:"traffic_type" validate:"required,oneof=Free MonthlyTransmit MonthlyReceive MonthlyRxTx TotallyTransmit TotallyReceive TotallyRxTx"`
	TrafficSize int                      `json:"traffic_size" validate:"omitempty,gte=0" example:"10737418240"` // 10 GiB
	Description string                   `json:"description" validate:"omitempty,max=1024" example:"User for testing VPN access"`
	Config      *models.OcservUserConfig `json:"config" validate:"required"`
}

type UpdateOcservUserData struct {
	Group       *string                  `json:"group" example:"default"`
	Password    *string                  `json:"password" validate:"min=2,max=32"`
	ExpireAt    *string                  `json:"expire_at"  validate:"omitempty" example:"2025-12-31"`
	Unlimited   bool                     `json:"unlimited" validate:"omitempty" example:"false" default:"false"`
	TrafficType *string                  `json:"traffic_type" validate:"oneof=Free MonthlyTransmit MonthlyReceive MonthlyRxTx TotallyTransmit TotallyReceive TotallyRxTx"`
	TrafficSize *int                     `json:"traffic_size" validate:"gte=0" example:"10737418240"` // 10 GiB
	Description *string                  `json:"description" validate:"omitempty,max=1024" example:"User for testing VPN access"`
	Config      *models.OcservUserConfig `json:"config" validate:"omitempty"`
}

type OcservUsersResponse struct {
	Meta   request.Meta        `json:"meta" validate:"required"`
	Result []models.OcservUser `json:"result" validate:"omitempty"`
}

type SyncOcpasswdRequest struct {
	Users       []user.Ocpasswd          `json:"users" validate:"required"`
	ExpireAt    *string                  `json:"expire_at" validate:"omitempty" example:"2025-12-31"`
	TrafficType *string                  `json:"traffic_type" validate:"oneof=Free MonthlyTransmit MonthlyReceive MonthlyRxTx TotallyTransmit TotallyReceive TotallyRxTx"`
	TrafficSize *int                     `json:"traffic_size" validate:"gte=0" example:"10737418240"` // 10 GiB
	Description *string                  `json:"description" validate:"omitempty,max=1024" example:"User for testing VPN access"`
	Config      *models.OcservUserConfig `json:"config" validate:"omitempty"`
}

type OcservUsersSyncResponse struct {
	Meta   request.Meta    `json:"meta" validate:"required"`
	Result []user.Ocpasswd `json:"result" validate:"omitempty"`
}

type ActivateUserData struct {
	ExpireAt *string `json:"expire_at" validate:"omitempty" example:"2025-12-31"`
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
