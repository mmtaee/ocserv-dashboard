# Domain: Security

This section provides instructions for auditing and hardening a Debian/Ubuntu server's security posture. Run these checks proactively during any security audit, or on-demand when the user asks about hardening, SSH, firewall, fail2ban, or updates.

## SSH Hardening Audit

Retrieve the effective sshd configuration (includes defaults merged with sshd_config) in one command:

```bash
ssh "$DEVOPS_HOST" "sudo sshd -T | grep -E 'permitrootlogin|passwordauthentication|maxauthtries|port'"
```

Evaluate each value against these thresholds:

- **PermitRootLogin**: must be `no`; `prohibit-password` is acceptable as a minimum; anything else is a finding.
- **PasswordAuthentication**: must be `no` (keys only). If `yes`, flag as critical.
- **MaxAuthTries**: must be 3 or less. Default is 6 — flag if not explicitly reduced.
- **AllowUsers / AllowGroups**: check whether either directive is set (limits which accounts can log in). Absence of both is a finding:
  ```bash
  ssh "$DEVOPS_HOST" "sudo sshd -T | grep -E 'allowusers|allowgroups'"
  ```
  If no output is returned, note that all system users with valid keys can log in.
- **Port**: note the current port. Non-default ports (not 22) add minor obscurity — not a security control, but worth noting in the report.

To harden a setting, use the **Config Edit Workflow** — read `/etc/ssh/sshd_config`, show the diff, confirm, backup, apply, validate with `sudo sshd -t`, then reload sshd (never restart — you'd lose your connection if it fails). Before reloading, detect the correct unit name: run `ssh "$DEVOPS_HOST" "systemctl list-units --plain | grep -E 'sshd?\.service'"` — use `sshd` if found, otherwise `ssh`. The Config Edit Workflow handles all of this automatically.

## UFW Firewall

**Note:** UFW is assumed for firewall management. Servers using `firewalld`, raw `iptables`, or `nftables` are not supported — UFW commands will fail on those systems.

Check firewall status and all active rules:

```bash
# Check for alternative firewalls before assuming UFW
if ssh "$DEVOPS_HOST" "systemctl is-active firewalld 2>/dev/null" | grep -q "^active"; then
  echo "firewalld is active — UFW management not supported; review rules with: sudo firewall-cmd --list-all"
elif ssh "$DEVOPS_HOST" "nft list ruleset 2>/dev/null" | grep -q .; then
  echo "nftables rules detected — UFW management not supported; review rules with: sudo nft list ruleset"
elif ! ssh "$DEVOPS_HOST" "dpkg -l ufw 2>/dev/null | grep -q '^ii'" 2>/dev/null; then
  echo "UFW not installed — firewall status unknown"
else
  ssh "$DEVOPS_HOST" "sudo ufw status verbose"
fi
```

Key rules to verify:
- SSH port (22 or custom) must be open — confirm before making any changes.
- HTTP (80/tcp) and HTTPS (443/tcp) must be open if a web server is running.
- Default inbound policy must be `deny` (shown as `Default: deny (incoming)`).
- Outbound policy is typically `allow` by default — acceptable.

To add a rule (always confirm the port with the user before running):
```bash
ssh "$DEVOPS_HOST" "sudo ufw allow <port>/tcp"
```

To enable UFW safely when it is inactive:

**Step 1** — Detect the active SSH port using two independent sources and cross-check:
```bash
# Source 1: the live session's actual connection port (most reliable)
ssh "$DEVOPS_HOST" "echo \$SSH_CONNECTION | awk '{print \$4}'"
# Source 2: the configured port from sshd
ssh "$DEVOPS_HOST" "sudo sshd -T | grep '^port'"
```
If both values agree, use that port. If they differ, show both to the user:
"⚠ Your live session is on port <A> (from SSH_CONNECTION) but sshd -T reports port <B>. Which port is your active connection using? (type the number)"
Do not proceed until you have a confirmed port.

**Step 2** — Show the computed allow rule to the user and ask for confirmation:
"I will add the following rule before enabling UFW: `sudo ufw allow <confirmed_port>/tcp`. Confirm? (yes/no)"
Do not proceed without explicit "yes".

**Step 3** — Add the SSH allow rule, verify it appears in the rule list, then enable:
```bash
ssh "$DEVOPS_HOST" "sudo ufw allow <detected_port>/tcp"
ssh "$DEVOPS_HOST" "sudo ufw show added"
```
Show the user the output of `ufw show added` and ask: "SSH port <N> is in the allow list. Enable UFW now? (yes/no)"

**Step 4** — Only after confirmation:
```bash
ssh "$DEVOPS_HOST" "sudo ufw --force enable"
ssh "$DEVOPS_HOST" "sudo ufw status verbose"
```
Show the final status output. Verify the SSH port is listed as ALLOW before reporting success.

## fail2ban

Check if fail2ban is installed:
```bash
ssh "$DEVOPS_HOST" "dpkg -l fail2ban 2>/dev/null | grep -q '^ii' && echo installed || echo missing"
```
If the output is `missing`, fail2ban is not installed — flag this as a finding.

Check overall status (lists all active jails):
```bash
ssh "$DEVOPS_HOST" "sudo fail2ban-client status"
```

Auto-detect the SSH jail name (Debian defaults to `ssh`, Ubuntu to `sshd`):
```bash
_f2b_jail=$(ssh "$DEVOPS_HOST" "sudo fail2ban-client status 2>/dev/null | grep -oi 'ssh[^ ,]*' | head -1")
if [ -z "$_f2b_jail" ]; then
  echo "No SSH jail found in fail2ban — check fail2ban config"
else
  echo "SSH jail detected: $_f2b_jail"
fi
```

**Before using `_f2b_jail` in any command, validate it** — the value comes from the server and must only contain safe characters:
```bash
if ! echo "$_f2b_jail" | grep -qE '^[a-z0-9_-]+$'; then
  echo "ERROR: unexpected jail name format '${_f2b_jail}' — aborting fail2ban checks"
  # stop
fi
```

Check the SSH jail specifically (using detected name):
```bash
ssh "$DEVOPS_HOST" "sudo fail2ban-client status '${_f2b_jail}'"
```

From the SSH jail output, verify:
- **Jail is active** (shown as `Status for the jail: <name>` with `Currently banned` count).
- **bantime**: retrieve from config — should be >= 3600 seconds (1 hour):
  ```bash
  ssh "$DEVOPS_HOST" "sudo fail2ban-client get '${_f2b_jail}' bantime"
  ```
- **maxretry**: should be <= 5:
  ```bash
  ssh "$DEVOPS_HOST" "sudo fail2ban-client get '${_f2b_jail}' maxretry"
  ```
- **findtime**: should be <= 600 seconds:
  ```bash
  ssh "$DEVOPS_HOST" "sudo fail2ban-client get '${_f2b_jail}' findtime"
  ```

If fail2ban is not installed, suggest: `sudo apt-get install -y fail2ban` and offer to configure it.

## Unattended Upgrades

Check if the package is installed and whether automatic upgrades are enabled:
```bash
ssh "$DEVOPS_HOST" "dpkg -l unattended-upgrades; cat /etc/apt/apt.conf.d/20auto-upgrades 2>/dev/null"
```

The `20auto-upgrades` file should contain:
```
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
```

If `APT::Periodic::Unattended-Upgrade` is `"0"` or the file is absent, flag as a finding. To enable:
```bash
ssh "$DEVOPS_HOST" "sudo dpkg-reconfigure -f noninteractive unattended-upgrades"
```

## Log Inspection

Check for suspicious activity in system logs:

```bash
# Recent failed SSH login attempts
ssh "$DEVOPS_HOST" "sudo grep 'Failed password\|Invalid user' /var/log/auth.log 2>/dev/null | tail -20"

# Recent successful SSH logins
ssh "$DEVOPS_HOST" "sudo grep 'Accepted' /var/log/auth.log 2>/dev/null | tail -10"

# System errors (last 24h)
ssh "$DEVOPS_HOST" "sudo journalctl -p err --since='24 hours ago' --no-pager | tail -30"

# Kernel messages (last 50 lines)
ssh "$DEVOPS_HOST" "sudo dmesg -T | tail -50"
```

Report any patterns that look suspicious: repeated failed logins from the same IP, unusual accepted logins at odd hours, OOM kills in dmesg, or disk I/O errors.

## Package CVE / Security Audit

```bash
# Verify apt-get is available (Debian/Ubuntu only)
if ! ssh "$DEVOPS_HOST" "dpkg -l apt 2>/dev/null | grep -q '^ii'" 2>/dev/null; then
  echo "apt-get not found — this server does not appear to be Debian/Ubuntu; package update checks are not supported"
  # skip remaining apt checks
fi
```

Check for security-specific pending updates:
```bash
ssh "$DEVOPS_HOST" "sudo apt-get --just-print upgrade 2>/dev/null | grep -c '^Inst.*security'"
```

Count all pending upgrades:
```bash
ssh "$DEVOPS_HOST" "apt list --upgradable 2>/dev/null | grep -c '/'"
```
This counts only actual package lines (which always contain `/`) — robust across apt versions and locales that may or may not emit a "Listing..." header.

For a full refresh and audit in one step:
```bash
ssh "$DEVOPS_HOST" "sudo apt-get update -qq && apt list --upgradable 2>/dev/null"
```

If security updates are pending, **do not apply them automatically**. Instead, follow the Production Safety Rules:

- If any service is tagged `production`: propose a targeted upgrade of only the affected packages. Ask the user which packages to update, then run:
  ```bash
  ssh "$DEVOPS_HOST" "sudo apt-get install --only-upgrade -y <package1> <package2>"
  ```
- If all services are tagged `experimental` (or no production services are running): show the pending list, confirm with the user, then run:
  ```bash
  ssh "$DEVOPS_HOST" "sudo apt-get upgrade -y"
  ```
- Never run `apt-get upgrade -y` without explicit user confirmation. The OS Package Updates Rule (audit by default, execute only on explicit request) always applies.

## Security Audit Report Format

When running a full security audit, collect all data first, then output a single structured report in this format:

```
🔒 Security Audit — <hostname> — <date>

SSH:
  ✓/✗ PermitRootLogin: <value>
  ✓/✗ PasswordAuthentication: <value>
  ✓/✗ MaxAuthTries: <value>
  ✓/✗ AllowUsers/AllowGroups: <value or "not set">
  Port: <value>

Firewall:
  ✓/✗ Status: active/inactive/not-installed/non-ufw-detected
  ✓/✗ Default inbound: deny/allow (n/a if not installed or non-UFW)
  Rules: <list of active UFW rules, "n/a", or "review manually (firewalld/nftables active)">
  Note: <omit if UFW; if firewalld/nftables detected: "firewalld/nftables active — UFW management not supported; review rules manually">

fail2ban:
  ✓/✗ Installed: yes/no
  ✓/✗ SSH jail: active/inactive
  bantime: <value>s  maxretry: <value>  findtime: <value>s
  Banned IPs: <count>

Updates:
  Pending: <count> packages (<N> security)
  Unattended upgrades: enabled/disabled

Findings:
  <numbered list of issues found, or "None — server is well hardened.">
```

Use ✓ for values that meet the threshold and ✗ for values that are findings. List all findings at the bottom with brief remediation advice.
