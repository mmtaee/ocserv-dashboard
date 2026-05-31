package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type OcservUserConfig struct {
	// Static IPv4 address to assign to the user. Example: '192.168.100.10'
	ExplicitIPv4 *string `json:"explicit-ipv4"`

	// The pool of addresses from which to assign to the user. Example: '192.168.1.0/24'
	IPv4Network *string `json:"ipv4-network"`

	// Comma-separated list of DNS servers to assign to the user. Example: '8.8.8.8,1.1.1.1'
	DNS *CSVStringList `json:"dns" gorm:"type:text"`

	// NetBIOS Name Servers (WINS) for Windows clients. Example: '192.168.1.1'
	NBNS *string `json:"nbns"`

	// Routes pushed to the user for routing traffic. Example: ['0.0.0.0/0', '10.10.0.0/16']
	Route *CSVStringList `json:"route" gorm:"type:text"`

	// List of networks to exclude from VPN routing. Example: ['192.168.0.0/16', '10.0.0.0/8']
	NoRoute *CSVStringList `json:"no-route" gorm:"type:text"`

	// Internal route available only via VPN. Example: '10.0.0.0/8'
	IRoute *string `json:"iroute"`

	// List of domains over which the provided DNS servers should be used. Example: ['example.com', 'internal.company.com']
	SplitDNS *CSVStringList `json:"split-dns" gorm:"type:text"`

	// Maximum session time in seconds before forced disconnect. Example: 3600
	SessionTimeout *int `json:"session-timeout"`

	// Time in seconds before disconnecting idle users. Example: 600
	IdleTimeout *int `json:"idle-timeout"`

	// Idle timeout in seconds for mobile users. Example: 900
	MobileIdleTimeout *int `json:"mobile-idle-timeout"`

	// Rekey time in seconds; triggers key renegotiation. Example: 86400 for 24 hours
	RekeyTime *int `json:"rekey-time"`

	// Allow user access only to defined routes. Example: true
	RestrictToRoutes *bool `json:"restrict-to-routes"`

	// Comma-separated list of allowed or blocked ports/protocols. Supports 'tcp(port)', 'udp(port)', 'icmp()', 'icmpv6()', and negation with '!()'. Example: 'tcp(443), udp(53)' or '!(tcp(22), udp(1194))'
	RestrictToPorts *string `json:"restrict-to-ports"`
}

type OcservUserCertificateBackup struct {
	Status    string `json:"status"` // active or suspended
	KeyPEM    string `json:"key_pem,omitempty"`
	CertPEM   string `json:"cert_pem,omitempty"`
	P12Base64 string `json:"p12_base64,omitempty"`
}

type OcservUser struct {
	ID                   uint                         `json:"-" gorm:"primaryKey;autoIncrement" `
	UID                  string                       `json:"uid" gorm:"gorm:type:char(26);not null;uniqueIndex" validate:"required"`
	Owner                string                       `json:"owner" gorm:"type:varchar(16);default:''" validate:"required"`
	Group                string                       `json:"group" gorm:"type:varchar(16);default:'defaults'" validate:"required"`
	Username             string                       `json:"username" gorm:"type:varchar(255);not null;uniqueIndex" validate:"required"`
	Password             string                       `json:"password" gorm:"type:varchar(255);not null" validate:"required"`
	IsLocked             bool                         `json:"is_locked" gorm:"default(false)" validate:"required"`
	CreatedAt            time.Time                    `json:"created_at" gorm:"autoCreateTime" validate:"required"`
	UpdatedAt            time.Time                    `json:"updated_at" gorm:"autoUpdateTime" validate:"omitempty"`
	ExpireAt             *time.Time                   `json:"expire_at" gorm:"type:date" validate:"omitempty"`
	DeactivatedAt        *time.Time                   `json:"deactivated_at" gorm:"type:date" validate:"omitempty"`
	TrafficType          string                       `json:"traffic_type" gorm:"type:varchar(32);not null;default:1" enums:"Free,MonthlyTransmit,MonthlyReceive,MonthlyRxTx,TotallyTransmit,TotallyReceive,TotallyRxTx" validate:"required"`
	TrafficSize          int64                        `json:"traffic_size" gorm:"not null" validate:"required"` // in bytes
	Rx                   int                          `json:"rx" gorm:"not null;default:0" validate:"required"` // Receive in bytes
	Tx                   int                          `json:"tx" gorm:"not null;default:0" validate:"required"` // Transmit in bytes
	Description          string                       `json:"description" gorm:"type:text" validate:"omitempty"`
	IsOnline             bool                         `json:"is_online" gorm:"-:migration;->" validate:"required"`
	OnlineUserSessions   []OnlineUserSession          `json:"online_sessions" gorm:"-" validate:"required"`
	Config               *OcservUserConfig            `json:"config" gorm:"type:text"`
	CertificateEnabled   bool                         `json:"certificate_enabled" gorm:"-"`
	CertificateAvailable bool                         `json:"certificate_available" gorm:"-"`
	Certificate          *OcservUserCertificateBackup `json:"certificate,omitempty" gorm:"-"`
}

type OcservUserTrafficStatistics struct {
	ID        uint      `json:"-" gorm:"primaryKey;autoIncrement"`
	OcUserID  uint      `json:"-" gorm:"index;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	Rx        int       `json:"rx" gorm:"default:0"` // in bytes
	Tx        int       `json:"tx" gorm:"default:0"` // in bytes
}

type DailyTraffic struct {
	Date string  `json:"date"` // Format: YYYY-MM-DD
	Rx   float64 `json:"rx"`   // in GiB
	Tx   float64 `json:"tx"`   // in GiB
}

const (
	EventUseragent     = "user-agent"
	EventHandshake     = "handshake"
	EventPeriodicStats = "periodic-stats"
	EventDisconnect    = "disconnect"
)

type OcservUserSessionLog struct {
	ID        uint      `json:"-" gorm:"primaryKey;autoIncrement"`
	Username  string    `json:"username" gorm:"type:varchar(64);index" validate:"required"`
	IP        string    `json:"ip" gorm:"type:varchar(45)" validate:"omitempty"`
	Event     string    `json:"event" gorm:"type:varchar(64)" enums:"user-agent,handshake,periodic-stats,disconnect" validate:"required"`
	Message   string    `json:"message" gorm:"type:text" validate:"required"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}

func (c *OcservUserConfig) Value() (driver.Value, error) {
	return json.Marshal(&c)
}

func (c *OcservUserConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {

	case []byte:
		return json.Unmarshal(v, c)

	case string:
		return json.Unmarshal([]byte(v), c)

	default:
		return fmt.Errorf("unsupported type for OcservUserConfig: %T", value)
	}
}

func (o *OcservUser) BeforeUpdate(tx *gorm.DB) (err error) {
	if o.TrafficType == "" {
		o.TrafficType = Free
	}

	if !validateTrafficType(o.TrafficType) {
		return fmt.Errorf("invalid TrafficType: %s", o.TrafficType)
	}

	if o.TrafficType == Free {
		o.TrafficSize = 0
	}
	return nil
}

func (o *OcservUser) BeforeCreate(tx *gorm.DB) (err error) {
	if o.TrafficType == "" {
		o.TrafficType = Free
	}

	if !validateTrafficType(o.TrafficType) {
		return fmt.Errorf("invalid TrafficType: %s", o.TrafficType)
	}

	if o.TrafficType == Free {
		o.TrafficSize = 0
	}

	if o.UID == "" {
		o.UID = ulid.Make().String()
	}
	return
}

func validateTrafficType(trafficType string) bool {
	switch trafficType {
	case Free, MonthlyTransmit, MonthlyReceive, MonthlyRxTx, TotallyTransmit, TotallyReceive, TotallyRxTx:
		return true
	default:
		return false
	}
}
