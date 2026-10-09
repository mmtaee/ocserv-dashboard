<div align="center">

# 🖥️ /devops

**Your server. Plain English. Safe by default.**

[![Agent Skill](https://img.shields.io/badge/Agent%20Skill-skills.sh-orange)](https://skills.sh)
[![Platform](https://img.shields.io/badge/Platform-Debian%20%7C%20Ubuntu-blue?logo=linux)](https://ubuntu.com)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

_Security audits · Service health · Nginx & SSL · Docker · Backups · Auto-rollback_

</div>

---

An agent skill that gives you an expert DevOps engineer on demand — audits security, manages services, configures nginx, renews SSL certificates, handles Docker containers, and applies changes safely with automatic backups, diffs, and rollback on failure.

Works with any agent that supports the [skills.sh](https://skills.sh) open standard: Claude Code, Cursor, VS Code Copilot, Codex, and more.

```
/devops
```

```
📊 Server Health — myserver — 2026-05-18

⚠ Issues Found:
  1. [CRITICAL] SSL cert for api.example.com expires in 5 days
  2. [WARNING]  Container "my-app" exited 3h ago
  3. [INFO]     8 security updates pending

✓ All good: nginx, sshd, fail2ban, memory (38% used), disk

Run `/devops renew ssl api.example.com` to act on the most urgent issue.
```

---

## Quick Start

```bash
# 1. Install
npx skills add H3nSte1n/skills --skill devops

# 2. Set your server (add to ~/.zshrc or ~/.bashrc)
export DEVOPS_HOST=agentops@your-server-ip

# 3. Run the setup wizard
/devops setup
```

The wizard checks prerequisites, optionally installs security tools (ufw, fail2ban, certbot), auto-discovers your services and Docker containers, and asks you to tag each as `production` or `experimental` — production services require an extra confirmation step before any change.

---

## Features

### 🔒 Security Hardening

- SSH config audit (PermitRootLogin, PasswordAuthentication, MaxAuthTries, AllowUsers)
- UFW firewall rules — check, add, enable safely (warns before locking you out)
- fail2ban — install, configure, inspect active bans
- Unattended security upgrades — verify and enable
- Full CVE/package audit with security-only update support

### 🖥️ Service Health & Monitoring

- System resources: CPU load, memory, disk usage with configurable alert thresholds
- systemd service status — all running, all failed
- SSL certificate expiry across all nginx vhosts (alerts at 30 days, critical at 7)
- Docker container health, resource stats, and log streaming

### 🌐 Nginx & SSL Management

- Config audit and syntax validation (`nginx -t`)
- Let's Encrypt renewal via certbot (dry-run first, always)
- Guided new vhost creation — reverse proxy template included
- vhost removal with optional cert revocation

### 🐳 Docker

- Full container lifecycle: start, stop, restart (with graceful 30s timeout)
- Docker Compose support — finds compose files from container labels
- Log streaming and resource stats
- Docker UI agnostic — manages via CLI over SSH; any Docker UI (Portainer, Dockge, Yacht, etc.) works as read-only reference

### 🛡️ Built-in Safety Rails

- **Diff before every edit** — you approve before anything changes
- **Auto-backup** before every file modification (stored in `/var/backups/devops/`)
- **Auto-rollback** if validation fails — the backup is automatically restored
- **Production safety** — services tagged `production` require `yes, production` to modify
- **Audit log** — every action timestamped locally at `~/.agents/state/devops/audit.log`

---

## Usage

### Health check

```
/devops
```

Runs a proactive health check across disk, SSL certs, failed services, Docker containers, and pending security updates. Surfaces the top 3–5 issues with severity levels and suggested next commands.

### Security

```
/devops audit security          # full SSH/UFW/fail2ban/CVE report
/devops harden ssh              # guided SSH config hardening
/devops check firewall          # UFW rules review
/devops check fail2ban          # jail status and config
/devops check updates           # pending packages (audit only by default)
/devops upgrade packages        # apply all pending updates (targeted for production services)
/devops apply security updates  # apply security patches only
```

### Service management

```
/devops status nginx
/devops restart nginx
/devops logs my-app
/devops check disk
/devops what did you do recently
```

### SSL & Nginx

```
/devops renew ssl                    # renew all certs
/devops renew ssl api.example.com    # renew one cert
/devops add vhost app.example.com    # guided new site setup
/devops check ssl app.example.com    # expiry check
```

### Docker

```
/devops docker status         # all containers with health
/devops logs my-container     # last 50 log lines
/devops restart my-container  # graceful restart
/devops docker stats          # CPU/memory per container
```

### Override host for one session

```
/devops root@1.2.3.4 check disk
/devops myserver audit security
```

---

## How it works

### Safety model

Every write operation follows a strict 10-step workflow:

```
Read → Diff → Confirm → [Production gate] → Backup → Apply → Validate → Reload → Verify → Log
```

If validation fails at any point, the skill **automatically restores the backup** and reports what went wrong.

### Production tagging

During setup, you tag each service as `production` or `experimental`. Production services require extra confirmation (`yes, production`) before any write. After each change the skill stays in a rollback-ready state — type `ok` to confirm or `rollback` to revert instantly.

---

## Installation

### Via skills.sh (recommended)

```bash
npx skills add H3nSte1n/skills --skill devops
```

Installs to `~/.agents/skills/devops/` and symlinks into each AI agent automatically.

### Manual install

```bash
git clone https://github.com/H3nSte1n/skills
cp -r skills/devops ~/.agents/skills/devops
```

### Set your server host

```bash
export DEVOPS_HOST=agentops@your-server-ip
```

Add to your shell profile (`~/.zshrc` or `~/.bashrc`) to make it permanent.

---

## Recommended Server Setup

Using `root@hostname` works but is a security antipattern. The recommended approach is a **dedicated `agentops` user** with its own SSH keypair and passwordless sudo in a separate sudoers file. This gives you independent revocation, a separate audit trail, and instant one-command revocation.

> **Why not restrict individual sudo commands?** The skill uses `sudo bash -c '...'` for scripted operations. Any sudoers whitelist that permits `bash` is equivalent to `NOPASSWD:ALL` — command-level restrictions provide no real protection. The dedicated-user model achieves actual containment.

**Step 1 — On your local machine, generate a dedicated keypair:**

```bash
ssh-keygen -t ed25519 -f ~/.ssh/agent_devops_key -C "agent-devops"
```

This creates two files: `~/.ssh/agent_devops_key` (private, never share) and `~/.ssh/agent_devops_key.pub` (public, goes on the server).

**Step 2 — On your server, create the user and install the public key:**

```bash
# Create dedicated user
sudo useradd -m -s /bin/bash agentops

# Unlock the account for SSH key auth (useradd locks accounts by default)
sudo passwd -d agentops

# Set up the .ssh directory
sudo mkdir -p /home/agentops/.ssh
sudo chmod 700 /home/agentops/.ssh

# Copy your public key — run this on your LOCAL machine, not the server:
# ssh-copy-id -i ~/.ssh/agent_devops_key.pub agentops@your-server-ip
#
# Or manually: paste the output of `cat ~/.ssh/agent_devops_key.pub` below,
# replacing the placeholder — the entire key must be on ONE line:
echo 'ssh-ed25519 AAAA...your-full-public-key... agent-devops' | sudo tee /home/agentops/.ssh/authorized_keys
sudo chmod 600 /home/agentops/.ssh/authorized_keys
sudo chown -R agentops:agentops /home/agentops/.ssh

# Grant passwordless sudo to agentops only
echo "agentops ALL=(ALL) NOPASSWD:ALL" | sudo tee /etc/sudoers.d/agentops
sudo chmod 440 /etc/sudoers.d/agentops

# Add to docker group so Docker commands work without sudo
sudo usermod -aG docker agentops
```

**Step 3 — If your server uses `AllowUsers` in sshd_config, add `agentops`:**

Many hardened servers whitelist SSH users explicitly. Check and update if needed:

```bash
# Check if AllowUsers is set
sudo grep AllowUsers /etc/ssh/sshd_config

# If it exists, add agentops to the list:
sudo sed -i 's/^AllowUsers.*/& agentops/' /etc/ssh/sshd_config
sudo systemctl reload ssh
```

If `grep` returns nothing, you can skip this step.

**Step 4 — On your local machine, add the host to `~/.ssh/config`:**

```
Host myserver-agent
  HostName your-server-ip
  Port 22
  User agentops
  IdentityFile ~/.ssh/agent_devops_key
```

Change `Port 22` to your server's SSH port if it's non-standard.

**Step 5 — Test:**

```bash
ssh myserver-agent "sudo echo ok"
```

You should see `ok`. If you do, the setup is complete.

To revoke agent access at any time: `sudo rm /etc/sudoers.d/agentops` on the server.

---

## Prerequisites

| Requirement                          | Notes                                                                   |
| ------------------------------------ | ----------------------------------------------------------------------- |
| SSH access to a Debian/Ubuntu server | Key-based auth — see setup above                                        |
| Passwordless sudo on the server      | Required — see setup above                                              |
| `ssh` and `openssl` locally          | macOS and Linux have both by default; Windows users should use WSL2     |
| systemd on the server                | Required for service management; OpenRC and SysV init are not supported |

**Tested on:** Ubuntu 22.04 LTS, Ubuntu 24.04 LTS, Debian 12 (Bookworm)

---

## Platform notes

| Limitation                          | Detail                                                                                                     |
| ----------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| **Debian/Ubuntu only**              | Uses `apt-get` for packages. RHEL, Fedora, Arch, Alpine, and other distributions are not supported.        |
| **UFW for firewall management**     | The skill manages UFW rules. `firewalld` and `nftables` are detected and flagged but not managed.          |
| **nginx for web / certbot for SSL** | Vhost management and Let's Encrypt target nginx + certbot. Apache, Caddy, and `acme.sh` are not supported. |
| **Docker CLI**                      | Container management uses the Docker CLI. Podman, containerd, and LXC are not supported.                   |
| **Single server**                   | Manages one server at a time via `DEVOPS_HOST`. Multi-server orchestration is out of scope.                |

---

## Domain coverage

| Domain            | Capabilities                                                 |
| ----------------- | ------------------------------------------------------------ |
| **Security**      | SSH hardening, UFW, fail2ban, unattended-upgrades, CVE audit |
| **Health**        | Disk/CPU/RAM, systemd status, SSL expiry, Docker health      |
| **Nginx & SSL**   | Config audit, Let's Encrypt, vhost create/remove             |
| **Docker**        | Container lifecycle, Compose, logs, stats                    |
| **Cron & Timers** | Crontab inspection, systemd timers (inspection only — creating/editing jobs is out of scope) |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to add new domains, coding conventions, and how to open a PR.

---

## License

MIT — use freely, modify as needed, share back improvements.
