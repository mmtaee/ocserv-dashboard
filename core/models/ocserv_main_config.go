package models

// OcservMainConfig represents the main ocserv configuration file
type OcservMainConfig struct {
	Auth                   *string `json:"auth"`
	RunAsUser              *string `json:"run-as-user"`
	RunAsGroup             *string `json:"run-as-group"`
	SocketFile             *string `json:"socket-file"`
	IsolateWorkers         *bool   `json:"isolate-workers"`
	MaxClients             *int    `json:"max-clients" validate:"omitempty,min=1"`
	KeepAlive              *int    `json:"keepalive" validate:"omitempty,min=1"`
	DPD                    *int    `json:"dpd" validate:"omitempty,min=1"`
	MobileDPD              *int    `json:"mobile-dpd" validate:"omitempty,min=1"`
	SwitchToTCPTimeout     *int    `json:"switch-to-tcp-timeout" validate:"omitempty,min=1"`
	TryMTUDiscovery        *bool   `json:"try-mtu-discovery"`
	ServerCert             *string `json:"server-cert"`
	ServerKey              *string `json:"server-key"`
	TLSPriorities          *string `json:"tls-priorities"`
	AuthTimeout            *int    `json:"auth-timeout" validate:"omitempty,min=1"`
	MinReauthTime          *int    `json:"min-reauth-time" validate:"omitempty,min=1"`
	MaxBanScore            *int    `json:"max-ban-score" validate:"omitempty,min=1"`
	BanResetTime           *int    `json:"ban-reset-time" validate:"omitempty,min=1"`
	CookieTimeout          *int    `json:"cookie-timeout" validate:"omitempty,min=1"`
	DenyRoaming            *bool   `json:"deny-roaming"`
	RekeyTime              *int    `json:"rekey-time" validate:"omitempty,min=1"`
	RekeyMethod            *string `json:"rekey-method" validate:"omitempty,oneof=ssl data"`
	UseOcctl               *bool   `json:"use-occtl"`
	PidFile                *string `json:"pid-file"`
	Device                 *string `json:"device"`
	PredictableIPs         *bool   `json:"predictable-ips"`
	TunnelAllDNS           *bool   `json:"tunnel-all-dns"`
	DNS                    *string `json:"dns"`
	PingLeases             *bool   `json:"ping-leases"`
	MTU                    *int    `json:"mtu" validate:"omitempty,min=576,max=65535"`
	CiscoClientCompat      *bool   `json:"cisco-client-compat"`
	DTLSLegacy             *bool   `json:"dtls-legacy"`
	TCPPort                *int    `json:"tcp-port" validate:"omitempty,min=1,max=65535"`
	UDPPort                *int    `json:"udp-port" validate:"omitempty,min=1,max=65535"`
	MaxSameClients         *int    `json:"max-same-clients" validate:"omitempty,min=1"`
	IPv4Network            *string `json:"ipv4-network"`
	ConfigPerGroup         *string `json:"config-per-group"`
	ConfigPerUser          *string `json:"config-per-user"`
	LogLevel               *int    `json:"log-level" validate:"omitempty,min=0,max=5"`
	RateLimitMS            *int    `json:"rate-limit-ms" validate:"omitempty,min=1"`
	PreLoginBanner         *string `json:"pre-login-banner" validate:"omitempty,max=100"`
	Banner                 *string `json:"banner" validate:"omitempty,max=190"`
}
