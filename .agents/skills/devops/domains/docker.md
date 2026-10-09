# Domain: Docker Management

## Connection
All Docker commands run over SSH: `ssh "$DEVOPS_HOST" "docker <subcommand>"`
Never use a Docker management UI's API for writes — any Docker UI (Portainer, Dockge, Yacht, etc.) is read-only reference only.
If Docker is not installed, skip Docker sections silently in health checks and note it in setup.

## Auto-Discovery
During setup and on bare invocation, discover all containers:
```bash
ssh "$DEVOPS_HOST" "docker ps -a --format '{{.Names}}|{{.Image}}|{{.Status}}|{{.Ports}}'"
```
Store results in ~/.agents/state/devops/server.md under `containers:`. Tag each as production/experimental (ask user during setup).

## Container Status
```bash
# All containers
ssh "$DEVOPS_HOST" "docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Image}}\t{{.Ports}}'"

# Specific container
ssh "$DEVOPS_HOST" "docker inspect <container> --format '{{.State.Status}} | health: {{.State.Health.Status}}'"

# Resource usage live snapshot
ssh "$DEVOPS_HOST" "docker stats --no-stream --format 'table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}'"
```

## Container Logs
```bash
# Last 50 lines
ssh "$DEVOPS_HOST" "docker logs --tail=50 <container>"

# Last 100 lines (for more context)
ssh "$DEVOPS_HOST" "docker logs --tail=100 <container>"

# Since a timestamp
ssh "$DEVOPS_HOST" "docker logs --since=1h <container>"
```

**Note on follow mode (`-f`):** `docker logs -f` streams indefinitely and cannot be interrupted in a non-interactive agent context — it will hang the session. Do not run it unless the user explicitly asks for live log streaming and understands they may need to restart the session. For normal troubleshooting, `--tail=100` and `--since=1h` are the preferred variants.

## Container Lifecycle (with production safety)
Always apply production safety rules before touching a production-tagged container.

```bash
# Start
ssh "$DEVOPS_HOST" "docker start <container>"

# Stop (graceful, 30s timeout)
ssh "$DEVOPS_HOST" "docker stop --time=30 <container>"

# Restart
ssh "$DEVOPS_HOST" "docker restart --time=30 <container>"

# Pull latest image (do not auto-recreate — show user what changed first)
ssh "$DEVOPS_HOST" "docker pull <image>"
```

## Docker Compose
**Compose version:** This skill uses `docker compose` (V2 plugin syntax, available in Docker Engine 20.10+). If your server uses Compose V1 as a standalone binary, substitute `docker-compose` for `docker compose` in all commands below. Check which is available:
```bash
ssh "$DEVOPS_HOST" "docker compose version 2>/dev/null || docker-compose version 2>/dev/null"
```

If a container was started with Compose:

**Finding and validating the compose directory:**
```bash
# Get the compose directory from the container label
compose_dir=$(ssh "$DEVOPS_HOST" "docker inspect <container> --format '{{index .Config.Labels \"com.docker.compose.project.working_dir\"}}'")

# Check if the container was started with Compose
if [ -z "$compose_dir" ]; then
  echo "This container was not started with Compose — manage it directly with 'docker restart <container>' or 'docker stop/start <container>'"
  # stop here, do not proceed with compose operations
fi

# Validate: must be absolute path, no path traversal, safe characters only
if ! echo "$compose_dir" | grep -qE '^/[a-zA-Z0-9/_.-]+$'; then
  echo "ERROR: compose_dir must be an absolute path with safe characters: $compose_dir — aborting"
  exit 1
fi
if echo "$compose_dir" | grep -q '\.\.'; then
  echo "ERROR: compose_dir contains path traversal: $compose_dir — aborting"
  exit 1
fi
```

Once validated, use the variable quoted in all commands:
```bash
# Up/down via compose
ssh "$DEVOPS_HOST" "cd \"$compose_dir\" && docker compose up -d"
ssh "$DEVOPS_HOST" "cd \"$compose_dir\" && docker compose down"

# Pull and recreate
ssh "$DEVOPS_HOST" "cd \"$compose_dir\" && docker compose pull && docker compose up -d"
```
Always show the compose file to the user before running `up` after a `pull`:
```bash
ssh "$DEVOPS_HOST" "cat \"$compose_dir/docker-compose.yml\" 2>/dev/null || cat \"$compose_dir/docker-compose.yaml\" 2>/dev/null || cat \"$compose_dir/compose.yml\" 2>/dev/null || cat \"$compose_dir/compose.yaml\" 2>/dev/null"
```

## Docker Management UIs (Read-Only Reference)
Any Docker management UI (Portainer, Dockge, Yacht, etc.) is informational only — never use a UI's API to make changes. Always use the Docker CLI over SSH for all operations. If the user references a Docker UI, note: "I manage Docker directly via CLI over SSH. Your Docker UI shows the same containers and is a useful reference, but all changes go through the CLI."

## Log entry on container changes
Always log container lifecycle actions to ~/.agents/state/devops/audit.log:
`[timestamp] DOCKER | target: <container> | detail: <action>`
