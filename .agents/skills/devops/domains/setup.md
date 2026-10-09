# Behavior: /devops setup (First-Run Wizard)

## Trigger

Run this wizard when:
- The user explicitly invokes `/devops setup`, or
- The skill loads `~/.agents/state/devops/server.md` and finds an empty or missing `services` block (first run).

---

## Step 1 — Show the Setup Plan

Before doing anything on the server, display the following plan to the user verbatim and wait for explicit approval. Do not proceed until the user says yes.

```
📋 DevOps Setup Plan

I'll configure your server for use with /devops. Here's what I'll check and potentially install:

Prerequisites (check only, no changes):
  □ SSH connectivity test
  □ Passwordless sudo verification
  □ /var/backups/devops/ directory
  □ UFW firewall (recommended prerequisite — flagged if missing)

Optional installs (I'll ask before each):
  □ ufw (firewall) — if not installed
  □ fail2ban — if not installed
  □ unattended-upgrades — if not installed
  □ certbot — if not installed and nginx is present

Discovery (read-only):
  □ Scan running systemd services
  □ Scan Docker containers
  □ Scan nginx vhosts and SSL certs
  □ Check system resources

Everything will be saved to ~/.agents/state/devops/server.md.

Proceed? (yes/no)
```

If the user says no, stop the setup wizard immediately. Do not run any further setup steps, but remain available for other /devops commands.

---

## Step 2 — Prerequisites Check

Run these checks in order. Stop and report clearly if any step fails.

**Step 2a — OS/distro check (run first, before anything else):**
```bash
ssh "$DEVOPS_HOST" "cat /etc/os-release 2>/dev/null || echo 'os-release not found'"
```
Parse the output for `ID=` and `ID_LIKE=` fields. This skill supports **Debian and Ubuntu only**.
- If `ID=debian`, `ID=ubuntu`, or `ID_LIKE` contains `debian` or `ubuntu` → proceed.
- For any other value (e.g. `ID=rhel`, `ID=fedora`, `ID=alpine`, `ID=amzn`) → **hard stop**:

  > "⛔ This skill supports Debian and Ubuntu only. Detected OS: <ID value>. Package management (`apt-get`), firewall (`ufw`), and service management (`systemd`) commands will not work on this system. Setup aborted."

  Do not run any further setup steps.

```bash
# Test SSH connectivity (see Connectivity check in SSH Connection section for fingerprint pre-check)
ssh "$DEVOPS_HOST" "echo ok"

# Test passwordless sudo (`sudo -n true` exits immediately with code 1 if a password is required, unlike `sudo echo` which may bypass the check)
ssh "$DEVOPS_HOST" "sudo -n true && echo ok || echo 'FAILED: passwordless sudo not configured'"

# Check whether backup directory exists
ssh "$DEVOPS_HOST" "[ -d /var/backups/devops ] && echo exists || echo missing"

# Check recommended prerequisites
ssh "$DEVOPS_HOST" "dpkg -l ufw 2>/dev/null | grep -q '^ii' && echo 'ufw: installed' || echo 'ufw: MISSING (recommended — install with sudo apt-get install ufw)'"
```

If the backup directory is missing, ask the user:
> "Create /var/backups/devops/ on the server? (yes/no)"

Only if they say yes:
```bash
ssh "$DEVOPS_HOST" "sudo mkdir -p /var/backups/devops && sudo chmod 750 /var/backups/devops"
```

---

## Step 3 — Optional Installs (ask before each)

For each tool in the list below, check whether it is installed, then ask the user before installing. Never install without explicit approval.

**Check pattern:**
```bash
ssh "$DEVOPS_HOST" "dpkg -l <package> 2>/dev/null | grep -q '^ii' && echo installed || echo missing"
```

**Install pattern (only after user says yes):**
```bash
ssh "$DEVOPS_HOST" "sudo apt-get install -y <package>"
```

**Tools to check in this order:**
1. `ufw` — package: `ufw`
2. `fail2ban` — package: `fail2ban`
3. `unattended-upgrades` — package: `unattended-upgrades`
4. `certbot` — only check if nginx is present on the server. Use a combined check (snap installs are not tracked by dpkg, and `-x` can fail on snap symlinks in non-interactive SSH):
   ```bash
   ssh "$DEVOPS_HOST" "( dpkg -l certbot 2>/dev/null | grep -q '^ii' || [ -L /usr/bin/certbot ] || [ -f /usr/bin/certbot ] ) && echo installed || echo missing"
   ```

For each missing tool, ask:
> "Install <tool>? (yes/no)"

Log every install to `~/.agents/state/devops/audit.log` using the INSTALL action.

---

## Step 4 — Auto-Discovery

Run all discovery commands below (read-only, no changes). Collect all output before presenting anything to the user.

```bash
# Running systemd services
ssh "$DEVOPS_HOST" "systemctl list-units --type=service --state=running --no-pager --plain | awk '{print \$1}'"

# Docker containers (skip silently if Docker is not installed)
ssh "$DEVOPS_HOST" "docker ps -a --format '{{.Names}}|{{.Image}}|{{.Status}}' 2>/dev/null"

# Nginx vhosts
ssh "$DEVOPS_HOST" "ls /etc/nginx/sites-enabled/ 2>/dev/null"
```

After collecting all discovery output, present **one combined list** of everything discovered — systemd services, Docker containers, and nginx vhosts — and ask the user to confirm production tagging. Do not split into separate sections; seeing everything together helps the user make consistent decisions.

Format the list clearly with recommended production items pre-marked (✓ suggested) and the rest as experimental. Assign each item a number. For example:

```
I found the following services and containers. I've pre-marked the ones I'd suggest tagging as production — please confirm or adjust.

Systemd services:
   1  ✓ nginx        — suggested: production
   2  ✓ ssh          — suggested: production
   3    docker       — suggested: experimental
   4    postfix      — suggested: experimental

Docker containers:
   5  ✓ app-web      — suggested: production
   6  ✓ app-db       — suggested: production
   7    changedetection — suggested: experimental
   8    portainer    — suggested: experimental

Type the numbers of any items to flip their tag (e.g. "3 7"), or type "ok" to accept all suggestions.
(Everything marked ✓ will be production; everything else will be experimental.)
```

Production suggestions: mark `nginx`, `ssh`, `sshd`, `fail2ban` as production. For containers: mark as production only if the name strongly suggests a primary user-facing service (e.g. contains a domain name, or ends/contains `-web`, `-app`, `-api`). Mark as experimental by default: `docker`, `containerd`, `db`, `redis`, `postgres`, `mysql` (backing services users restart frequently), monitoring/tooling (`portainer`, `watchtower`, `uptime-kuma`, `changedetection`), and any container whose name suggests a utility or admin UI (`phpmyadmin`, `redis-commander`, `adminer`). When in doubt, prefer experimental — users can promote to production, and a mistaken production tag causes friction on every routine restart.

Wait for the user's response. The user types numbers to flip items (e.g. "3 7" flips items 3 and 7 from their suggested tag). If the user types "ok" or says they're happy, accept all suggestions as-is. If the user flips some items, show the updated list and ask them to confirm with "ok" before proceeding. Any service not in the state file after this run defaults to STRICT at runtime until the next `/devops setup`.

---

## Step 4b — Confirm Before Writing

Before writing `~/.agents/state/devops/server.md`, show the user the exact YAML that will be written:

> "Here is what I'll write to `~/.agents/state/devops/server.md`:
> ```yaml
> host: "<host>"
> last_audit: "<today>"
> services:
>   - name: <service>
>     type: systemd
>     tag: <production|experimental>
>     status: <status>
> containers:
>   - name: <container>
>     image: <image>
>     tag: <production|experimental>
>     status: <status>
> vhosts:
>   - domain: <domain>
>     ssl: <true|false>
> ```
> Write this file? (yes/no)"

Wait for explicit "yes" before proceeding. If the user says no, allow them to adjust tags and re-run Step 4b.

---

## Step 5 — Save State

First, ensure the local state directory exists:
```bash
mkdir -p ~/.agents/state/devops
```

Write the full `~/.agents/state/devops/server.md` with all discovered and tagged data. Follow these rules:

- Write the complete file (this is a full rediscovery — the whole file is replaced).
- **Preserve existing `notes:` entries**: if a `server.md` already exists, read its `notes:` block first and include those entries in the new file. All other sections (services, containers, vhosts) are replaced with fresh discovery data.
- Populate: services list with names, images (for Docker), status, and the `production`/`experimental` tag assigned in Step 4.
- Record nginx vhosts under the appropriate section.
- Set `last_audit` to today's date (format: `YYYY-MM-DD`).

After writing, confirm to the user:
> "Setup complete. `~/.agents/state/devops/server.md` has been updated. Run `/devops` anytime to get a health check or security audit."
