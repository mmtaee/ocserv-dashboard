# Changelog

All notable changes to this project will be documented in this file.

## [1.0.0] - 2026-05-22

### Added

- Initial public release
- SSH-based server management for Debian/Ubuntu Linux servers
- Health check: system resources, systemd status, SSL expiry, Docker health
- Security domain: SSH hardening, UFW, fail2ban, unattended-upgrades, CVE audit
- Nginx & SSL domain: config audit, Let's Encrypt renewal, vhost create/remove
- Docker domain: container lifecycle, Compose, logs, stats
- Cron & Timers domain: crontab and systemd timer inspection (read-only)
- 10-step config edit workflow with auto-backup and auto-rollback
- Production safety: STRICT mode with `yes, production` confirmation gate
- State file at `~/.agents/state/devops/server.md` for service tagging
- Append-only audit log at `~/.agents/state/devops/audit.log`
- Setup wizard with auto-discovery and service tagging
- Progressive domain loading — core always in context, domains loaded on demand
- Tool detection via `dpkg -l` throughout — reliable in non-interactive SSH PATH where `/usr/sbin` is absent
- Certbot detection handles both apt and snap installs (`[ -L /usr/bin/certbot ]`)
- SSL domain extraction filters to valid FQDNs — excludes `localhost`, `default_server`, IP addresses
- fail2ban jail name validated with regex and single-quoted in SSH commands
- sshd reload detects correct unit name at runtime (`ssh` on Debian, `sshd` on Ubuntu)
- Setup tagging heuristic: `docker`, `containerd`, `db`, `redis` default to experimental; user-facing containers suggested as production
- Certbot install guidance: apt-first (`python3-certbot-nginx`), snap as fallback
