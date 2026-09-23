# Ocserv Dashboard Backend

This repository provides one backend application that runs the Admin API, Customer API, Worker, and optional Telegram Bot. Its single configurable Docker image also runs Ocserv, PostgreSQL 18, nginx, and the enabled UI applications.

## Runtime sequence

The Docker container starts services in this order:

```text
prepare Ocserv configuration and networking
→ initialize and start PostgreSQL 18
→ wait for PostgreSQL readiness
→ run database migrations
→ start the unified backend
→ start nginx on full nodes
→ start Ocserv
→ supervise all critical processes
```

If PostgreSQL, Ocserv, or the backend exits unexpectedly, the remaining processes are stopped. SIGINT and SIGTERM trigger graceful shutdown.

## Requirements

- Linux host with Docker and Docker Compose
- `/dev/net/tun` available on the host
- Permission to add the `NET_ADMIN` capability
- Permission to read `/var/run/docker.sock`
- Ports `443/tcp`, `443/udp`, and `80/tcp` available for a full node (`8080/tcp` for an agent node)

## Guided installation

Download and run the installer, then choose Docker or systemd first and Master or Agent:

```bash
curl -fsSLO https://raw.githubusercontent.com/mmtaee/ocserv-dashboard/master/install.sh
bash install.sh
```

Or with wget:

```bash
wget -q https://raw.githubusercontent.com/mmtaee/ocserv-dashboard/master/install.sh -O install.sh
bash install.sh
```

The bootstrap installer downloads the latest tagged release into `./ocserv-dashboard`, records its tag in `.release`, then starts `setup.sh`.

From an existing repository checkout, run the guided installer directly:

```bash
./setup.sh
```

The terminal GUI lets you choose Install or Upgrade, go back between deployment and node choices, and edit the main `.env` settings before installation. Choose **Upgrade** to automatically detect the running Docker container or systemd service, fetch a newer release when available, and upgrade without deployment prompts. You can also run `./setup.sh --update`.

The installer creates `.env` from `.env.example` only when it is missing, generates initial secrets, and asks before changing an existing `AGENT_NODE` value. Non-interactive selection is also supported:

```bash
./setup.sh --deployment docker --node master
./setup.sh --deployment docker --node agent
./setup.sh --deployment systemd --node master
./setup.sh --deployment systemd --node agent
```

Master selects `AGENT_NODE=false`; Agent selects `AGENT_NODE=true`.

For production, create persistent host directories:

```bash
sudo mkdir -p \
  /opt/ocserv_dashboard/docker_volumes/postgresql18 \
  /opt/ocserv_dashboard/docker_volumes/ocserv \
  /opt/ocserv_dashboard/docker_volumes/cron_journal \
  /opt/ocserv_dashboard/docker_volumes/telegram_receipts
```

The container must be named `ocserv`. The Worker reads Ocserv logs through the Docker API using that container name.

## Production Docker deployment

Copy and configure the environment file:

```bash
cp .env.example .env
```

Replace at least these values:

```env
SECRET_KEY="generate-a-random-value"
AGENT_NODE=false
POSTGRES_PASSWORD="set-a-strong-database-password"
SUPERADMIN_USERNAME="admin"
SUPERADMIN_PASSWORD="set-a-strong-superadmin-password"
```

Random secrets can be generated with:

```bash
openssl rand -hex 32
```

`AGENT_NODE=false` runs the master-node agent-management APIs. Set
`AGENT_NODE=true` on an agent deployment to enable the agent-token migration and
local token CLI. Master and agent databases receive different node-specific
tables.

On an agent node, manage the local token from the backend directory (or use the
installed `/opt/ocserv-dashboard/backend` binary):

```bash
go run main.go agent-token create
go run main.go agent-token get
go run main.go agent-token renew
go run main.go agent-token remove
```

For installed deployments:

```bash
# Docker agent
docker exec ocserv backend agent-token get

# systemd agent
cd /opt/ocserv-dashboard
sudo ./backend agent-token get
```

These commands reject master mode. Agent-token management is not exposed over HTTP.

Use the following Compose configuration for a full node. Keep the feature values
in `.env` and `build.args` identical because they control both image contents and runtime services:

```yaml
services:
  ocserv:
    container_name: ocserv
    image: ocserv-dashboard:latest
    build:
      context: .
      dockerfile: deploy/docker/Dockerfile
      args:
        GO_VERSION: ${GO_VERSION:-1.27.1}
        NODE_VERSION: ${NODE_VERSION:-24}
        AGENT_NODE: ${AGENT_NODE:-false}
        CUSTOMER_API_ENABLED: ${CUSTOMER_API_ENABLED:-true}
        TELEGRAM_BOT_ENABLED: ${TELEGRAM_BOT_ENABLED:-false}
    env_file:
      - ./.env
    environment:
      HOST_PROC: /host/proc
      HOST_SYS: /host/sys
      POSTGRES_HOST: 127.0.0.1
      TELEGRAM_RECEIPTS_DIR: /opt/ocserv_dashboard/uploads/receipts
    cap_add:
      - NET_ADMIN
    sysctls:
      net.ipv4.ip_forward: "1"
    devices:
      - /dev/net/tun:/dev/net/tun
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /opt/ocserv_dashboard/docker_volumes/ocserv:/etc/ocserv
      - /opt/ocserv_dashboard/docker_volumes/telegram_receipts:/opt/ocserv_dashboard/uploads/receipts
      - /opt/ocserv_dashboard/docker_volumes/postgresql18:/var/lib/postgresql
      - /opt/ocserv_dashboard/docker_volumes/cron_journal:/app/cron_journal
    ports:
      - "${OCSERV_PORT:-443}:${OCSERV_PORT:-443}/tcp"
      - "${OCSERV_PORT:-443}:${OCSERV_PORT:-443}/udp"
      - "${HTTP_PORT:-80}:80"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-fsS", "http://127.0.0.1/health"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
```

Save it as `compose.production.yml`, then run:

```bash
sudo docker compose -f compose.production.yml up --build -d
```

Check startup and migration logs:

```bash
sudo docker compose -f compose.production.yml logs -f ocserv
```

### Standalone production container

Build the image without Compose:

```bash
sudo docker build \
  --build-arg GO_VERSION=1.27.1 \
  --build-arg NODE_VERSION=24 \
  --build-arg AGENT_NODE=false \
  --build-arg CUSTOMER_API_ENABLED=true \
  --build-arg TELEGRAM_BOT_ENABLED=false \
  -f deploy/docker/Dockerfile \
  -t ocserv-dashboard:master \
  .
```

Run the complete production backend stack:

```bash
sudo docker run -d \
  --name ocserv \
  --restart unless-stopped \
  --env-file ./.env \
  --env HOST_PROC=/host/proc \
  --env HOST_SYS=/host/sys \
  --env POSTGRES_HOST=127.0.0.1 \
  --env TELEGRAM_RECEIPTS_DIR=/opt/ocserv_dashboard/uploads/receipts \
  --cap-add NET_ADMIN \
  --sysctl net.ipv4.ip_forward=1 \
  --device /dev/net/tun:/dev/net/tun \
  --volume /var/run/docker.sock:/var/run/docker.sock:ro \
  --volume /proc:/host/proc:ro \
  --volume /sys:/host/sys:ro \
  --volume /opt/ocserv_dashboard/docker_volumes/ocserv:/etc/ocserv \
  --volume /opt/ocserv_dashboard/docker_volumes/telegram_receipts:/opt/ocserv_dashboard/uploads/receipts \
  --volume /opt/ocserv_dashboard/docker_volumes/postgresql18:/var/lib/postgresql \
  --volume /opt/ocserv_dashboard/docker_volumes/cron_journal:/app/cron_journal \
  --publish 443:443/tcp \
  --publish 443:443/udp \
  --publish 80:80 \
  ocserv-dashboard:master
```

Follow its logs with:

```bash
sudo docker logs -f ocserv
```

For an agent image, build the same Dockerfile with `AGENT_NODE=true` and
`CUSTOMER_API_ENABLED=false`; publish `8080:8080` instead of port 80. That
variant skips both UI builds and does not start nginx.

## Development Docker deployment

`deploy/docker/Dockerfile.dev` provides the backend-only development image used by `scripts/dev.sh`; the admin and customer UIs continue to run from their local development servers.

Development state is stored in the repository-local `.volume/` directory:

```text
.volume/
├── postgresql18/
├── ocserv/
├── cron_journal/
└── telegram_receipts/
```

The directory is excluded from Git and the Docker build context.

The quickest development startup is:

```bash
./scripts/dev.sh
```

The script validates Docker and TUN access, creates `.volume/` and its persistent directories, builds with detailed logs, replaces containers marked as development, starts the stack, prints endpoint information, and follows container logs. Ctrl-C stops log viewing without stopping the container.

Useful overrides:

```bash
FOLLOW_LOGS=false ./scripts/dev.sh
NO_CACHE=true ./scripts/dev.sh
DEV_API_PORT=9080 DEV_POSTGRES_PORT=55435 ./scripts/dev.sh
OCSERV_DEBUG=3 ./scripts/dev.sh
REPLACE_CONTAINER=true ./scripts/dev.sh
DEV_DATA_ROOT=/tmp/ocserv-dashboard-data ./scripts/dev.sh
```

`REPLACE_CONTAINER=true` may stop and remove an existing container named `ocserv`; persistent host data is not deleted.

When using Compose or `docker run` directly instead of the development script, create the local directories from the repository root first:

```bash
mkdir -p \
  .volume/postgresql18 \
  .volume/ocserv \
  .volume/cron_journal \
  .volume/telegram_receipts
```

Use this Compose configuration:

```yaml
services:
  ocserv:
    container_name: ocserv
    image: ocserv-dashboard-dev:latest
    build:
      context: .
      dockerfile: deploy/docker/Dockerfile.dev
      args:
        GO_VERSION: ${GO_VERSION:-1.27.1}
    environment:
      HOST_PROC: /host/proc
      HOST_SYS: /host/sys
      POSTGRES_HOST: 127.0.0.1
      TELEGRAM_RECEIPTS_DIR: /opt/ocserv_dashboard/uploads/receipts
    cap_add:
      - NET_ADMIN
    sysctls:
      net.ipv4.ip_forward: "1"
    devices:
      - /dev/net/tun:/dev/net/tun
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - ./.volume/ocserv:/etc/ocserv
      - ./.volume/telegram_receipts:/opt/ocserv_dashboard/uploads/receipts
      - ./.volume/postgresql18:/var/lib/postgresql
      - ./.volume/cron_journal:/app/cron_journal
    ports:
      - "443:443/tcp"
      - "443:443/udp"
      - "8080:8080"
      - "127.0.0.1:5435:5432"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-fsS", "http://127.0.0.1:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
```

Save it as `compose.development.yml`, then run:

```bash
sudo docker compose -f compose.development.yml up --build
```

The UI can access the API at:

```text
http://localhost:8080
```

Swagger UI is available at:

```text
http://localhost:8080/swagger/index.html
```

The development PostgreSQL server is available locally at `127.0.0.1:5435`.

### Standalone development container

Build the development image without Compose:

```bash
sudo docker build \
  --build-arg GO_VERSION=1.27.1 \
  -f deploy/docker/Dockerfile.dev \
  -t ocserv-dashboard-dev:latest \
  .
```

Run all development backend services:

```bash
sudo docker run -d \
  --name ocserv \
  --restart unless-stopped \
  --env HOST_PROC=/host/proc \
  --env HOST_SYS=/host/sys \
  --env POSTGRES_HOST=127.0.0.1 \
  --env TELEGRAM_RECEIPTS_DIR=/opt/ocserv_dashboard/uploads/receipts \
  --cap-add NET_ADMIN \
  --sysctl net.ipv4.ip_forward=1 \
  --device /dev/net/tun:/dev/net/tun \
  --volume /var/run/docker.sock:/var/run/docker.sock:ro \
  --volume /proc:/host/proc:ro \
  --volume /sys:/host/sys:ro \
  --volume "${PWD}/.volume/ocserv:/etc/ocserv" \
  --volume "${PWD}/.volume/telegram_receipts:/opt/ocserv_dashboard/uploads/receipts" \
  --volume "${PWD}/.volume/postgresql18:/var/lib/postgresql" \
  --volume "${PWD}/.volume/cron_journal:/app/cron_journal" \
  --publish 443:443/tcp \
  --publish 443:443/udp \
  --publish 8080:8080 \
  --publish 127.0.0.1:5435:5432 \
  ocserv-dashboard-dev:latest
```

Development data under `.volume/` is isolated from the production data under `/opt/ocserv_dashboard/docker_volumes`. The examples still use the same container name and default ports, so do not run them simultaneously without changing those values.

## Telegram Bot

Telegram settings and bot accounts are stored through the dashboard. Enable the in-process Telegram service with:

```env
TELEGRAM_BOT_ENABLED=true
```

The customer API is enabled by default and can be disabled independently:

```env
CUSTOMER_API_ENABLED=false
```

Production receipts are persisted at:

```text
/opt/ocserv_dashboard/docker_volumes/telegram_receipts
```

Development receipts are persisted at `.volume/telegram_receipts` in the repository.

No separate Telegram executable or container is required.

## Native systemd installation

For a native Debian or Ubuntu deployment, use the guided installer:

```bash
./setup.sh
```

The installer can provision local PostgreSQL when `INSTALL_POSTGRES=true`. It runs migrations and the idempotent `backend create-superadmin` command before startup using `SUPERADMIN_USERNAME` and `SUPERADMIN_PASSWORD`. Set `INSTALL_POSTGRES=false` and configure the PostgreSQL connection variables when using an existing server.

Check native service status:

```bash
systemctl status ocserv ocserv-dashboard
journalctl -fu ocserv-dashboard
```

## Common operations

Health check:

```bash
curl -fsS http://127.0.0.1/health
```

Open API documentation:

```text
http://127.0.0.1/swagger/index.html
```

Stop the production stack:

```bash
docker compose -f compose.production.yml down
```

Rebuild after backend changes:

```bash
docker compose -f compose.development.yml up --build
```

Reset development data only when it is safe to permanently remove the selected host directories. PostgreSQL 18 data must be mounted at `/var/lib/postgresql`; `/var/lib/postgresql/data` is the pre-18 layout and must not be used for this image.

Legacy deployments stored PostgreSQL 17 data under `/opt/ocserv_dashboard/docker_volumes/pg_db`. That directory is deliberately not mounted by the PostgreSQL 18 examples. It remains untouched and must be migrated explicitly with `pg_upgrade` if its data is needed. Do not copy PostgreSQL 17 files directly into `postgresql18`.

## Troubleshooting

If Ocserv cannot initialize networking, confirm the TUN device and capability:

```bash
ls -l /dev/net/tun
docker inspect ocserv --format '{{json .HostConfig.CapAdd}}'
```

If the Worker stops because it cannot read Ocserv logs, confirm:

- the container is named `ocserv`;
- `/var/run/docker.sock` is mounted;
- the Docker daemon socket is readable by the container process.

Development defaults to `OCSERV_DEBUG=999`, matching the legacy container and sending Ocserv output to `docker logs`. Reduce the level when less output is desired:

```bash
OCSERV_DEBUG=3 ./scripts/dev.sh
```

If startup stops before the backend is launched, inspect the container logs. PostgreSQL readiness and migration failures are reported before Ocserv and the backend are started.
