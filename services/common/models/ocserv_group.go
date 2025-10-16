package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type OcservGroupConfig struct {
	// Comma-separated list of DNS servers to assign to the client. Example: '8.8.8.8,1.1.1.1'
	DNS *CSVStringList `json:"dns" gorm:"type:text"`

	// NetBIOS Name Servers (WINS) for Windows clients. Example: '192.168.1.1'
	NBNS *string `json:"nbns"`

	// The pool of addresses that leases will be given from. Example: '192.168.1.0/24'
	IPv4Network *string `json:"ipv4-network"`

	// Maximum receive bandwidth in bytes per second. Example: '100000' for 100 KB/s
	RxDataPerSec *int `json:"rx-data-per-sec"`

	// Maximum transmit bandwidth in bytes per second. Example: '200000' for 200 KB/s
	TxDataPerSec *int `json:"tx-data-per-sec"`

	// Static IPv4 address to assign to client. Example: '192.168.100.10'
	ExplicitIPv4 *string `json:"explicit-ipv4"`

	// Linux control group to assign the VPN worker process to. Format: 'controller,subsystem:name'. Example: 'cpuset,cpu:test'
	CGroup *string `json:"cgroup"`

	// Internal route available only via VPN. Format: 'IP/prefix'. Example: '10.0.0.0/8'
	IRoute *string `json:"iroute"`

	// Routes pushed to the client for routing traffic. Example: ['0.0.0.0/0', '10.10.0.0/16']
	Route *CSVStringList `json:"route" gorm:"type:text"`

	// List of networks to exclude from VPN routing. Each entry should be in 'IP/prefix' format. Example: ['192.168.0.0/16', '10.0.0.0/8']
	NoRoute *CSVStringList `json:"no-route" gorm:"type:text"`

	// Priority for routes; lower is higher priority. Example: 1
	NetPriority *int `json:"net-priority"`

	// Disconnect client if its IP changes (e.g., due to network switch). Example: true
	DenyRoaming *bool `json:"deny-roaming"`

	// Disables UDP, enforcing TCP-only VPN connection. Example: true
	NoUDP *bool `json:"no-udp"`

	// Interval in seconds to send keep-alive pings. Example: 60
	KeepAlive *int `json:"keepalive"`

	// Dead Peer Detection timeout in seconds. Example: 90
	DPD *int `json:"dpd"`

	// DPD timeout specifically for mobile clients. Example: 300
	MobileDPD *int `json:"mobile-dpd"`

	// Maximum simultaneous logins per user. Example: 2
	MaxSameClients *int `json:"max-same-clients"`

	// Force all DNS traffic through the VPN tunnel. Example: true
	TunnelAllDNS *bool `json:"tunnel-all-dns"`

	// Interval in seconds for stats reporting. Example: 300
	StatsReportTime *int `json:"stats-report-time"`

	// Tunnel interface MTU to avoid fragmentation. Example: 1400
	MTU *int `json:"mtu"`

	// Time in seconds before disconnecting idle clients. Example: 600
	IdleTimeout *int `json:"idle-timeout"`

	// Idle timeout for mobile clients. Example: 900
	MobileIdleTimeout *int `json:"mobile-idle-timeout"`

	// Allow client access only to defined routes. Example: true
	RestrictUserToRoutes *bool `json:"restrict-user-to-routes"`

	// Comma-separated list of allowed (or blocked, if negated) protocols and ports. Supports 'tcp(port)', 'udp(port)', 'icmp()', 'icmpv6()', and negation with '!()'. Example: 'tcp(443), tcp(80), udp(53)', or '!(tcp(22), udp(1194))'
	RestrictUserToPorts *string `json:"restrict-user-to-ports"`

	// List of domains over which the provided DNS servers should be used. Example: ['example.com', 'internal.company.com']
	SplitDNS *CSVStringList `json:"split-dns" gorm:"type:text"`

	// Max session time in seconds before forced disconnect. Example: 3600
	SessionTimeout *int `json:"session-timeout"`
}

type OcservGroup struct {
	ID     uint               `json:"id" gorm:"primaryKey;autoIncrement"`
	Name   string             `json:"name" gorm:"type:varchar(255);not null;unique" validate:"required"`
	Config *OcservGroupConfig `json:"config" gorm:"type:json"`
}

func (c *OcservGroupConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *OcservGroupConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}
	return json.Unmarshal(bytes, c)
}
