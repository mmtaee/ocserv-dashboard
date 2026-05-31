package customer

import (
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"time"
)

type SummaryData struct {
	Username string `json:"username" validate:"required,min=2,max=32"`
	Password string `json:"password" validate:"required,min=2,max=32"`
}

type ModelCustomer struct {
	Owner                string     `json:"owner" gorm:"type:varchar(16);default:''" validate:"required"`
	Username             string     `json:"username" gorm:"type:varchar(16);not null;uniqueIndex" validate:"required"`
	IsLocked             bool       `json:"is_locked" gorm:"default(false)" validate:"required"`
	CertificateEnabled   bool       `json:"certificate_enabled" validate:"required"`
	CertificateAvailable bool       `json:"certificate_available" validate:"required"`
	ExpireAt             *time.Time `json:"expire_at" gorm:"type:date" validate:"required"`
	DeactivatedAt        *time.Time `json:"deactivated_at" gorm:"type:date" validate:"required"`
	TrafficType          string     `json:"traffic_type" gorm:"type:varchar(32);not null;default:1" enums:"Free,MonthlyTransmit,MonthlyReceive,MonthlyRxTx,TotallyTransmit,TotallyReceive,TotallyRxTx" validate:"required"`
	TrafficSize          int64      `json:"traffic_size" gorm:"not null" validate:"required"` // in GiB  >> x * 1024 ** 3
	Rx                   int        `json:"rx" gorm:"not null;default:0" validate:"required"` // Receive in bytes
	Tx                   int        `json:"tx" gorm:"not null;default:0" validate:"required"` // Transmit in bytes
}

type UsageResponse struct {
	DateStart  time.Time                  `json:"date_start" validate:"required"`
	DateEnd    time.Time                  `json:"date_end" validate:"required"`
	Bandwidths repository.TotalBandwidths `json:"bandwidths" validate:"required"`
}

type SummaryResponse struct {
	OcservUser ModelCustomer `json:"ocserv_user" validate:"required"`
	Usage      UsageResponse `json:"usage" validate:"required"`
}
