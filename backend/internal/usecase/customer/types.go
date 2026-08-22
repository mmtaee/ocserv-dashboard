package customer

import (
	"context"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
)

type SystemRepository interface {
	System(ctx context.Context) (*models.System, error)
}

type OcservUserRepository interface {
	GetByUsername(ctx context.Context, username string) (*models.OcservUser, error)
	TotalBandwidthUserDateRange(ctx context.Context, uid string, dateStart, dateEnd *time.Time) (repository.TotalBandwidths, error)
	CertificatePathByUsername(ctx context.Context, username string) (string, error)
	CreateCertificate(ctx context.Context, uid string) error
}

type OcctlRepository interface {
	Disconnect(username string) (string, error)
}

type Credentials struct {
	Username string `json:"username" validate:"required,min=2,max=32"`
	Password string `json:"password" validate:"required,min=2,max=32"`
}

type Customer struct {
	Owner                string     `json:"owner"`
	Username             string     `json:"username"`
	IsLocked             bool       `json:"is_locked"`
	CertificateEnabled   bool       `json:"certificate_enabled"`
	CertificateAvailable bool       `json:"certificate_available"`
	ExpireAt             *time.Time `json:"expire_at"`
	DeactivatedAt        *time.Time `json:"deactivated_at"`
	TrafficType          string     `json:"traffic_type"`
	TrafficSize          int64      `json:"traffic_size"`
	Rx                   int        `json:"rx"`
	Tx                   int        `json:"tx"`
}

type Usage struct {
	DateStart  time.Time                  `json:"date_start"`
	DateEnd    time.Time                  `json:"date_end"`
	Bandwidths repository.TotalBandwidths `json:"bandwidths"`
}

type Summary struct {
	OcservUser Customer `json:"ocserv_user"`
	Usage      Usage    `json:"usage"`
}

type CiscoSetup struct {
	CertificateImportURI string    `json:"certificate_import_uri"`
	ConnectionCreateURI  string    `json:"connection_create_uri"`
	CertificatePassword  string    `json:"certificate_password"`
	ConnectionName       string    `json:"connection_name"`
	ServerAddress        string    `json:"server_address"`
	ServerPort           int       `json:"server_port"`
	ExpiresAt            time.Time `json:"expires_at"`
}
