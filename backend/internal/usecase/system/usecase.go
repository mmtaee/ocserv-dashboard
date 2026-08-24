package system

import (
	"context"
	"fmt"
	"net"
	"strings"
)

type Usecase struct {
	runtime Runtime
	config  ConfigStore
}

func New(runtime Runtime, config ConfigStore) *Usecase {
	return &Usecase{runtime: runtime, config: config}
}

func (u *Usecase) Status(ctx context.Context) (*Status, error) {
	return u.runtime.Status(ctx)
}

func (u *Usecase) Restart(ctx context.Context) (*ActionResult, error) {
	if err := u.runtime.Restart(ctx); err != nil {
		return nil, err
	}
	return &ActionResult{Message: "service restarting started successfully"}, nil
}

func (u *Usecase) Enable(ctx context.Context) (*ActionResult, error) {
	status, err := u.runtime.Status(ctx)
	if err != nil {
		return nil, err
	}
	if status.UnitFileState == "enabled" {
		return &ActionResult{Message: "service already enabled"}, nil
	}
	if err := u.runtime.Enable(ctx); err != nil {
		return nil, err
	}
	return &ActionResult{Message: "service enabling started successfully"}, nil
}

func (u *Usecase) Disable(ctx context.Context) (*ActionResult, error) {
	status, err := u.runtime.Status(ctx)
	if err != nil {
		return nil, err
	}
	if status.UnitFileState == "disabled" {
		return &ActionResult{Message: "service already disabled"}, nil
	}
	if err := u.runtime.Disable(ctx); err != nil {
		return nil, err
	}
	return &ActionResult{Message: "service disabling started successfully"}, nil
}

func (u *Usecase) Config(ctx context.Context) (*OcservConfig, error) {
	return u.config.Read(ctx)
}

func (u *Usecase) UpdateConfig(ctx context.Context, changes OcservConfig) (*OcservConfig, error) {
	if err := ValidateConfig(changes); err != nil {
		return nil, err
	}
	updated, err := u.config.Write(ctx, changes)
	if err != nil {
		return nil, err
	}
	if err := u.runtime.Restart(ctx); err != nil {
		return nil, fmt.Errorf("ocserv configuration saved but runtime restart failed: %w", err)
	}
	return updated, nil
}

func ValidateConfig(config OcservConfig) error {
	if !hasChanges(config) {
		return ErrNoConfigChanges
	}
	for name, port := range map[string]*int{"tcp_port": config.TCPPort, "udp_port": config.UDPPort} {
		if port != nil && (*port < 1 || *port > 65535) {
			return invalid(name, "must be between 1 and 65535")
		}
	}
	for name, value := range map[string]*int{
		"max_clients": config.MaxClients, "max_same_clients": config.MaxSameClients,
	} {
		if value != nil && (*value < 1 || *value > 100000) {
			return invalid(name, "must be between 1 and 100000")
		}
	}
	for name, value := range map[string]*int{
		"keepalive": config.Keepalive, "dpd": config.DPD, "mobile_dpd": config.MobileDPD,
		"switch_to_tcp_timeout": config.SwitchToTCPTimeout, "auth_timeout": config.AuthTimeout,
		"min_reauth_time": config.MinReauthTime, "ban_reset_time": config.BanResetTime,
		"cookie_timeout": config.CookieTimeout, "rekey_time": config.RekeyTime,
	} {
		if value != nil && (*value < 0 || *value > 2678400) {
			return invalid(name, "must be between 0 and 2678400")
		}
	}
	if config.MaxBanScore != nil && (*config.MaxBanScore < 0 || *config.MaxBanScore > 100000) {
		return invalid("max_ban_score", "must be between 0 and 100000")
	}
	if config.MTU != nil && (*config.MTU < 576 || *config.MTU > 9000) {
		return invalid("mtu", "must be between 576 and 9000")
	}
	if config.LogLevel != nil && (*config.LogLevel < 0 || *config.LogLevel > 4) {
		return invalid("log_level", "must be between 0 and 4")
	}
	if config.RateLimitMS != nil && (*config.RateLimitMS < 0 || *config.RateLimitMS > 60000) {
		return invalid("rate_limit_ms", "must be between 0 and 60000")
	}
	if config.RekeyMethod != nil && !config.RekeyMethod.IsValid() {
		return invalid("rekey_method", "must be ssl or new-tunnel")
	}
	if config.IPv4Network != nil {
		ip, _, err := net.ParseCIDR(strings.TrimSpace(*config.IPv4Network))
		if err != nil || ip.To4() == nil {
			return invalid("ipv4_network", "must be a valid IPv4 CIDR")
		}
	}
	if config.DNS != nil {
		if len(*config.DNS) == 0 || len(*config.DNS) > 8 {
			return invalid("dns", "must contain between 1 and 8 IP addresses")
		}
		for _, address := range *config.DNS {
			if net.ParseIP(strings.TrimSpace(address)) == nil {
				return invalid("dns", "contains an invalid IP address")
			}
		}
	}
	for name, value := range map[string]*string{"banner": config.Banner, "pre_login_banner": config.PreLoginBanner} {
		if value != nil && (len(*value) > 2048 || strings.ContainsRune(*value, '\x00')) {
			return invalid(name, "must be at most 2048 characters without NUL bytes")
		}
	}
	return nil
}

func hasChanges(config OcservConfig) bool {
	return config.TCPPort != nil || config.UDPPort != nil || config.IPv4Network != nil || config.DNS != nil ||
		config.MaxClients != nil || config.MaxSameClients != nil || config.Keepalive != nil || config.DPD != nil ||
		config.MobileDPD != nil || config.SwitchToTCPTimeout != nil || config.TryMTUDiscovery != nil ||
		config.AuthTimeout != nil || config.MinReauthTime != nil || config.MaxBanScore != nil ||
		config.BanResetTime != nil || config.CookieTimeout != nil || config.DenyRoaming != nil ||
		config.RekeyTime != nil || config.RekeyMethod != nil || config.PredictableIPs != nil ||
		config.TunnelAllDNS != nil || config.PingLeases != nil || config.MTU != nil ||
		config.CiscoClientCompat != nil || config.DTLSLegacy != nil || config.LogLevel != nil ||
		config.RateLimitMS != nil || config.PreLoginBanner != nil || config.Banner != nil
}

func invalid(field, message string) error {
	return fmt.Errorf("%w: %s %s", ErrInvalidConfig, field, message)
}
