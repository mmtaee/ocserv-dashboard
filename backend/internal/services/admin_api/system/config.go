package system

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	systemusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/system"
)

const DefaultOcservConfigPath = "/etc/ocserv/ocserv.conf"

var supportedConfigKeys = []string{
	"tcp-port", "udp-port", "ipv4-network", "dns", "max-clients", "max-same-clients",
	"keepalive", "dpd", "mobile-dpd", "switch-to-tcp-timeout", "try-mtu-discovery",
	"auth-timeout", "min-reauth-time", "max-ban-score", "ban-reset-time", "cookie-timeout",
	"deny-roaming", "rekey-time", "rekey-method", "predictable-ips", "tunnel-all-dns",
	"ping-leases", "mtu", "cisco-client-compat", "dtls-legacy", "log-level", "rate-limit-ms",
	"pre-login-banner", "banner",
}

type ConfigFile struct {
	path string
}

func NewConfigFile(path string) *ConfigFile {
	return &ConfigFile{path: filepath.Clean(path)}
}

func (f *ConfigFile) Read(ctx context.Context) (*systemusecase.OcservConfig, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	content, err := os.ReadFile(f.path)
	if err != nil {
		return nil, err
	}
	return ParseConfig(content)
}

func (f *ConfigFile) Write(ctx context.Context, changes systemusecase.OcservConfig) (*systemusecase.OcservConfig, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("OCSERV_REGENERATE_CONFIG")), "true") {
		return nil, errors.New("OCSERV_REGENERATE_CONFIG must be disabled before API-managed configuration can be saved")
	}
	content, err := os.ReadFile(f.path)
	if err != nil {
		return nil, err
	}
	updated, err := applyConfigChanges(content, changes)
	if err != nil {
		return nil, err
	}
	if err := writeAtomic(f.path, updated); err != nil {
		return nil, err
	}
	return ParseConfig(updated)
}

// ParseConfig parses only the API-managed allowlist. Unsupported directives
// remain outside the schema and are preserved verbatim when writing.
func ParseConfig(content []byte) (*systemusecase.OcservConfig, error) {
	var result systemusecase.OcservConfig
	var dns []string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		key, value, ok := parseDirective(scanner.Text())
		if !ok {
			continue
		}
		var err error
		switch key {
		case "tcp-port":
			result.TCPPort, err = parseInt(value)
		case "udp-port":
			result.UDPPort, err = parseInt(value)
		case "ipv4-network":
			result.IPv4Network = stringPointer(unquote(value))
		case "dns":
			dns = append(dns, unquote(value))
		case "max-clients":
			result.MaxClients, err = parseInt(value)
		case "max-same-clients":
			result.MaxSameClients, err = parseInt(value)
		case "keepalive":
			result.Keepalive, err = parseInt(value)
		case "dpd":
			result.DPD, err = parseInt(value)
		case "mobile-dpd":
			result.MobileDPD, err = parseInt(value)
		case "switch-to-tcp-timeout":
			result.SwitchToTCPTimeout, err = parseInt(value)
		case "try-mtu-discovery":
			result.TryMTUDiscovery, err = parseBool(value)
		case "auth-timeout":
			result.AuthTimeout, err = parseInt(value)
		case "min-reauth-time":
			result.MinReauthTime, err = parseInt(value)
		case "max-ban-score":
			result.MaxBanScore, err = parseInt(value)
		case "ban-reset-time":
			result.BanResetTime, err = parseInt(value)
		case "cookie-timeout":
			result.CookieTimeout, err = parseInt(value)
		case "deny-roaming":
			result.DenyRoaming, err = parseBool(value)
		case "rekey-time":
			result.RekeyTime, err = parseInt(value)
		case "rekey-method":
			method := systemusecase.RekeyMethod(unquote(value))
			result.RekeyMethod = &method
		case "predictable-ips":
			result.PredictableIPs, err = parseBool(value)
		case "tunnel-all-dns":
			result.TunnelAllDNS, err = parseBool(value)
		case "ping-leases":
			result.PingLeases, err = parseBool(value)
		case "mtu":
			result.MTU, err = parseInt(value)
		case "cisco-client-compat":
			result.CiscoClientCompat, err = parseBool(value)
		case "dtls-legacy":
			result.DTLSLegacy, err = parseBool(value)
		case "log-level":
			result.LogLevel, err = parseInt(value)
		case "rate-limit-ms":
			result.RateLimitMS, err = parseInt(value)
		case "pre-login-banner":
			result.PreLoginBanner = stringPointer(unquote(value))
		case "banner":
			result.Banner = stringPointer(unquote(value))
		}
		if err != nil {
			return nil, fmt.Errorf("parse ocserv directive %s: %w", key, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(dns) > 0 {
		result.DNS = &dns
	}
	if err := systemusecase.ValidateConfig(result); err != nil && !errors.Is(err, systemusecase.ErrNoConfigChanges) {
		return nil, err
	}
	return &result, nil
}

func applyConfigChanges(content []byte, changes systemusecase.OcservConfig) ([]byte, error) {
	updates := configUpdates(changes)
	lines := strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")
	output := make([]string, 0, len(lines)+len(updates))
	written := make(map[string]bool, len(updates))
	for _, line := range lines {
		key, _, ok := parseDirective(line)
		values, managed := updates[key]
		if !ok || !managed {
			output = append(output, line)
			continue
		}
		if !written[key] {
			output = append(output, values...)
			written[key] = true
		}
	}
	for _, key := range supportedConfigKeys {
		if values, ok := updates[key]; ok && !written[key] {
			output = append(output, values...)
		}
	}
	return []byte(strings.Join(output, "\n") + "\n"), nil
}

func configUpdates(config systemusecase.OcservConfig) map[string][]string {
	updates := make(map[string][]string)
	addInt := func(key string, value *int) {
		if value != nil {
			updates[key] = []string{key + " = " + strconv.Itoa(*value)}
		}
	}
	addBool := func(key string, value *bool) {
		if value != nil {
			updates[key] = []string{key + " = " + strconv.FormatBool(*value)}
		}
	}
	addString := func(key string, value *string, quoted bool) {
		if value != nil {
			rendered := *value
			if quoted {
				rendered = strconv.Quote(rendered)
			}
			updates[key] = []string{key + " = " + rendered}
		}
	}
	addInt("tcp-port", config.TCPPort)
	addInt("udp-port", config.UDPPort)
	addString("ipv4-network", config.IPv4Network, false)
	if config.DNS != nil {
		lines := make([]string, 0, len(*config.DNS))
		for _, address := range *config.DNS {
			lines = append(lines, "dns = "+strings.TrimSpace(address))
		}
		updates["dns"] = lines
	}
	addInt("max-clients", config.MaxClients)
	addInt("max-same-clients", config.MaxSameClients)
	addInt("keepalive", config.Keepalive)
	addInt("dpd", config.DPD)
	addInt("mobile-dpd", config.MobileDPD)
	addInt("switch-to-tcp-timeout", config.SwitchToTCPTimeout)
	addBool("try-mtu-discovery", config.TryMTUDiscovery)
	addInt("auth-timeout", config.AuthTimeout)
	addInt("min-reauth-time", config.MinReauthTime)
	addInt("max-ban-score", config.MaxBanScore)
	addInt("ban-reset-time", config.BanResetTime)
	addInt("cookie-timeout", config.CookieTimeout)
	addBool("deny-roaming", config.DenyRoaming)
	addInt("rekey-time", config.RekeyTime)
	if config.RekeyMethod != nil {
		value := string(*config.RekeyMethod)
		addString("rekey-method", &value, false)
	}
	addBool("predictable-ips", config.PredictableIPs)
	addBool("tunnel-all-dns", config.TunnelAllDNS)
	addBool("ping-leases", config.PingLeases)
	addInt("mtu", config.MTU)
	addBool("cisco-client-compat", config.CiscoClientCompat)
	addBool("dtls-legacy", config.DTLSLegacy)
	addInt("log-level", config.LogLevel)
	addInt("rate-limit-ms", config.RateLimitMS)
	addString("pre-login-banner", config.PreLoginBanner, true)
	addString("banner", config.Banner, true)
	return updates
}

func writeAtomic(path string, content []byte) (returnErr error) {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("refusing to replace non-regular ocserv config %s", path)
	}
	directory := filepath.Dir(path)
	temp, err := os.CreateTemp(directory, ".ocserv.conf-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() {
		_ = temp.Close()
		if returnErr != nil {
			_ = os.Remove(tempPath)
		}
	}()
	if err := temp.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		if err := temp.Chown(int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}
	}
	if _, err := temp.Write(content); err != nil {
		return err
	}
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	dir, err := os.Open(directory)
	if err != nil {
		return err
	}
	defer dir.Close()
	if err := dir.Sync(); err != nil {
		return err
	}
	return nil
}

func parseDirective(line string) (string, string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", "", false
	}
	parts := strings.SplitN(trimmed, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return strings.TrimSpace(parts[0]), stripInlineComment(strings.TrimSpace(parts[1])), true
}

func stripInlineComment(value string) string {
	var quote rune
	escaped := false
	for index, char := range value {
		switch {
		case quote != 0 && escaped:
			escaped = false
		case quote != 0 && char == '\\':
			escaped = true
		case quote == 0 && (char == '\'' || char == '"'):
			quote = char
		case quote != 0 && char == quote:
			quote = 0
		case quote == 0 && char == '#':
			return strings.TrimSpace(value[:index])
		}
	}
	return strings.TrimSpace(value)
}

func parseInt(value string) (*int, error) {
	parsed, err := strconv.Atoi(unquote(value))
	return &parsed, err
}

func parseBool(value string) (*bool, error) {
	parsed, err := strconv.ParseBool(unquote(value))
	return &parsed, err
}

func unquote(value string) string {
	value = strings.TrimSpace(value)
	if parsed, err := strconv.Unquote(value); err == nil {
		return parsed
	}
	return value
}

func stringPointer(value string) *string {
	return &value
}
