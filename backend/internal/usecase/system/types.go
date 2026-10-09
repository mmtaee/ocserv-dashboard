package system

import (
	"context"
	"errors"
)

var (
	ErrInvalidConfig    = errors.New("invalid ocserv configuration")
	ErrNoConfigChanges  = errors.New("no ocserv configuration changes provided")
	ErrUpdateNotAllowed = errors.New("ocserv configuration updates are not supported in this deployment mode")
)

// RekeyMethod is the supported Ocserv key renegotiation strategy.
type RekeyMethod string

const (
	RekeyMethodSSL       RekeyMethod = "ssl"
	RekeyMethodNewTunnel RekeyMethod = "new-tunnel"
)

func (method RekeyMethod) IsValid() bool {
	return method == RekeyMethodSSL || method == RekeyMethodNewTunnel
}

// OcservConfig is the explicit allowlist of main ocserv.conf settings managed
// by the API. Pointer fields distinguish omitted PATCH values from zero values.
type OcservConfig struct {
	TCPPort            *int         `json:"tcp_port,omitempty"`
	UDPPort            *int         `json:"udp_port,omitempty"`
	IPv4Network        *string      `json:"ipv4_network,omitempty"`
	DNS                *[]string    `json:"dns,omitempty"`
	MaxClients         *int         `json:"max_clients,omitempty"`
	MaxSameClients     *int         `json:"max_same_clients,omitempty"`
	Keepalive          *int         `json:"keepalive,omitempty"`
	DPD                *int         `json:"dpd,omitempty"`
	MobileDPD          *int         `json:"mobile_dpd,omitempty"`
	SwitchToTCPTimeout *int         `json:"switch_to_tcp_timeout,omitempty"`
	TryMTUDiscovery    *bool        `json:"try_mtu_discovery,omitempty"`
	AuthTimeout        *int         `json:"auth_timeout,omitempty"`
	MinReauthTime      *int         `json:"min_reauth_time,omitempty"`
	MaxBanScore        *int         `json:"max_ban_score,omitempty"`
	BanResetTime       *int         `json:"ban_reset_time,omitempty"`
	CookieTimeout      *int         `json:"cookie_timeout,omitempty"`
	DenyRoaming        *bool        `json:"deny_roaming,omitempty"`
	RekeyTime          *int         `json:"rekey_time,omitempty"`
	RekeyMethod        *RekeyMethod `json:"rekey_method,omitempty" enums:"ssl,new-tunnel"`
	PredictableIPs     *bool        `json:"predictable_ips,omitempty"`
	TunnelAllDNS       *bool        `json:"tunnel_all_dns,omitempty"`
	PingLeases         *bool        `json:"ping_leases,omitempty"`
	MTU                *int         `json:"mtu,omitempty"`
	CiscoClientCompat  *bool        `json:"cisco_client_compat,omitempty"`
	DTLSLegacy         *bool        `json:"dtls_legacy,omitempty"`
	LogLevel           *int         `json:"log_level,omitempty"`
	RateLimitMS        *int         `json:"rate_limit_ms,omitempty"`
	PreLoginBanner     *string      `json:"pre_login_banner,omitempty"`
	Banner             *string      `json:"banner,omitempty"`
}

type ConfigResponse struct {
	OcservConfig
	AllowUpdate bool `json:"allow_update"`
}

type Runtime interface {
	Restart(ctx context.Context) error
}

type ConfigStore interface {
	Read(ctx context.Context) (*OcservConfig, error)
	Write(ctx context.Context, changes OcservConfig) (*OcservConfig, error)
}
