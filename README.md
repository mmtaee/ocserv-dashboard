# OCServ Dashboard Backend

This repository provides one backend application that runs the Admin API, Customer API, Worker, and optional Telegram Bot. The Docker images also run OCServ and PostgreSQL 18 in the same container.

## Runtime sequence

The Docker container starts services in this order:

```text
prepare OCServ configuration and networking
→ initialize and start PostgreSQL 18
→ wait for PostgreSQL readiness
→ run database migrations
→ start the unified backend
→ start OCServ
→ supervise all critical processes
```

If PostgreSQL, OCServ, or the backend exits unexpectedly, the remaining processes are stopped. SIGINT and SIGTERM trigger graceful shutdown.

## Requirements

- Linux host with Docker and Docker Compose
- `/dev/net/tun` available on the host
- Permission to add the `NET_ADMIN` capability
- Permission to read `/var/run/docker.sock`
- Ports `443/tcp`, `443/udp`, and `8080/tcp` available

For production, create persistent host directories:

```bash
sudo mkdir -p \
  /opt/ocserv_dashboard/docker_volumes/postgresql18 \
  /opt/ocserv_dashboard/docker_volumes/ocserv \
  /opt/ocserv_dashboard/docker_volumes/cron_journal \
  /opt/ocserv_dashboard/docker_volumes/telegram_receipts
```

The container must be named `ocserv`. The Worker reads OCServ logs through the Docker API using that container name.

## Production Docker deployment

Copy and configure the environment file:

```bash
cp .env.sample .env
```

Replace at least these values:

```env
SECRET_KEY="generate-a-random-value"
JWT_SECRET="generate-a-different-random-value"
POSTGRES_PASSWORD="set-a-strong-database-password"
SUPERADMIN_USERNAME="admin"
SUPERADMIN_PASSWORD="set-a-strong-superadmin-password"
```

Random secrets can be generated with:

```bash
openssl rand -hex 32
```

Use the following Compose configuration with the production `Dockerfile`:

```yaml
services:
  ocserv:
    container_name: ocserv
    image: ocserv-dashboard:latest
    build:
      context: .
      dockerfile: Dockerfile
      args:
        GO_VERSION: ${GO_VERSION:-1.26.0}
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
      - "8080:8080"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-fsS", "http://127.0.0.1:8080/health"]
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
  --build-arg GO_VERSION=1.26.0 \
  -f Dockerfile \
  -t ocserv-dashboard:latest \
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
  --publish 8080:8080 \
  ocserv-dashboard:latest
```

Follow its logs with:

```bash
sudo docker logs -f ocserv
```

## UI development Docker deployment

`Dockerfile-Dev` contains backend services only. It enables backend debug mode, permits local UI origins on ports `3000` and `5173`, disables Telegram by default, and supplies development-only database credentials.

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

The script validates Docker and TUN access, creates `.volume/` and its persistent directories, builds with detailed logs, replaces containers marked as development or using the configured development image, starts the stack, prints endpoint information, and follows container logs. Ctrl-C stops log viewing without stopping the container.

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
      dockerfile: Dockerfile-Dev
      args:
        GO_VERSION: ${GO_VERSION:-1.26.0}
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

The development PostgreSQL server is available locally at `127.0.0.1:5435`.

### Standalone development container

Build the UI-development backend image without Compose:

```bash
sudo docker build \
  --build-arg GO_VERSION=1.26.0 \
  -f Dockerfile-Dev \
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

Production receipts are persisted at:

```text
/opt/ocserv_dashboard/docker_volumes/telegram_receipts
```

Development receipts are persisted at `.volume/telegram_receipts` in the repository.

No separate Telegram executable or container is required.

## Native systemd installation

For a native Debian or Ubuntu deployment:

```bash
cp .env.sample .env
sudo ./scripts/install.sh
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
curl -fsS http://127.0.0.1:8080/health
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

If OCServ cannot initialize networking, confirm the TUN device and capability:

```bash
ls -l /dev/net/tun
docker inspect ocserv --format '{{json .HostConfig.CapAdd}}'
```

If the Worker stops because it cannot read OCServ logs, confirm:

- the container is named `ocserv`;
- `/var/run/docker.sock` is mounted;
- the Docker daemon socket is readable by the container process.

Development defaults to `OCSERV_DEBUG=999`, matching the legacy container and sending OCServ output to `docker logs`. Reduce the level when less output is desired:

```bash
OCSERV_DEBUG=3 ./scripts/dev.sh
```

If startup stops before the backend is launched, inspect the container logs. PostgreSQL readiness and migration failures are reported before OCServ and the backend are started.
