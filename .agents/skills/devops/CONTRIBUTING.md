# Contributing to /devops

Thank you for improving this skill. Contributions that fix bugs, improve safety, or add well-scoped domains are welcome.

---

## What belongs here

✅ Good contributions:

- Bug fixes in existing domains
- Graceful degradation for edge cases (missing tools, non-standard configs)
- New domains for commonly self-hosted services (PostgreSQL, Redis, Caddy, Prometheus, etc.)
- Safety improvements (injection safety, validation, rollback handling)
- Documentation improvements

❌ Out of scope:

- Multi-distro support (RHEL, Alpine, etc.) — the skill targets Debian/Ubuntu intentionally
- Multi-server orchestration
- GUI/API-based management (all operations go through SSH CLI)
- Features that significantly increase the core SKILL.md size

---

## File structure

```
~/.agents/skills/devops/
  SKILL.md              ← always loaded: SSH, backup, safety rules, health check, core domains
  domains/
    security.md         ← loaded for: audit, harden, ssh, firewall, fail2ban
    nginx.md            ← loaded for: nginx, ssl, vhost, certbot
    docker.md           ← loaded for: docker, container, compose
    setup.md            ← loaded for: /devops setup, first run
    cron.md             ← loaded for: cron, timer, scheduled tasks
  README.md
  CHANGELOG.md
  CONTRIBUTING.md
  LICENSE
```

**Safety-critical code lives in `SKILL.md`** (backup system, config edit workflow, injection safety, production gates, audit log). This is always in context. Domain files are loaded on demand.

---

## Adding a new domain

1. **Create `domains/<name>.md`** — follow the structure of an existing domain file:
   - Start with a single `# Domain: <Name>` heading
   - Use `## Section` headings for each capability
   - Every command runs over SSH (`ssh "$DEVOPS_HOST" "..."`)
   - Every command that modifies the server must log to `~/.agents/state/devops/audit.log`
   - Add graceful degradation: check if the tool is installed before using it

2. **Add a row to the Domain Detection table in `SKILL.md`** — map trigger keywords to your file path:

   ```
   | User intent | Trigger keywords | Load |
   | Your domain | keyword1, keyword2 | `./domains/<name>.md` |
   ```

3. **Add a row to the Domain Coverage table in `README.md`**

4. **Update `CHANGELOG.md`** with a new version entry

### Minimal domain template

````markdown
# Domain: <Name>

Use this domain when the user asks about <topic>.

## Check Status

```bash
# Check if <tool> is installed (use dpkg — `which` is unreliable in non-interactive SSH PATH)
# For snap-installable tools (e.g. certbot), also check for the symlink: || [ -L /usr/bin/<tool> ]
# Note: use -L (symlink exists), not -x (executable) — snap symlink targets aren't accessible in non-interactive SSH
ssh "$DEVOPS_HOST" "dpkg -l <package> 2>/dev/null | grep -q '^ii' && echo installed || echo missing"
```
````

If not installed, report: "`<tool>` is not installed — install with: `sudo apt-get install -y <package>`"

## Common Operation

```bash
ssh "$DEVOPS_HOST" "<command>"
```

Always log changes:

```bash
echo "[$(date '+%Y-%m-%d %H:%M:%S')] ACTION | target: <target> | detail: <what changed>" >> ~/.agents/state/devops/audit.log
```

````

---

## Coding conventions

### Injection safety

All user-supplied values (domain names, container names, file paths) must be single-quoted when interpolated into remote shell commands:

```bash
# Safe
ssh "$DEVOPS_HOST" "sudo certbot --nginx -d '${domain}'"

# Unsafe — do not do this
ssh "$DEVOPS_HOST" "sudo certbot --nginx -d ${domain}"
````

Audit log entries must be sanitized before writing:

```bash
safe_target=$(printf '%s' "$target" | LC_ALL=C tr -cd '[:print:]' | tr -d ';\|&$`()' | cut -c1-200)
```

### Graceful degradation

Every optional tool must have a skip-with-note path:

```bash
if ! ssh "$DEVOPS_HOST" "dpkg -l <package> 2>/dev/null | grep -q '^ii'" 2>/dev/null; then
  echo "<tool> not installed — skipping <section>"
  # stop here, do not error
fi
```

### Backup before edit

Any command that modifies a file on the server must follow the **Backup System** in `SKILL.md` — create a timestamped backup first, then apply the change, then run cleanup.

### Audit log

Every action that modifies the server must append a line to `~/.agents/state/devops/audit.log`. See the **Audit Log** section in `SKILL.md` for format and sanitization requirements.

---

## Submitting a PR

1. Fork the repository and create a branch: `feature/<domain-name>` or `fix/<description>`
2. Test against a real Debian/Ubuntu server (Ubuntu 22.04 LTS or 24.04 LTS recommended)
3. Update `CHANGELOG.md` under a new version heading
4. Open a PR with a description of what changed and why

---

## Questions

Open a GitHub issue for discussion before starting large changes.
