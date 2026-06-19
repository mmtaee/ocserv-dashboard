# ====================================================
# Builder Stage
# ====================================================
FROM golang:1.26 AS builder

ARG GO_PROXY

# Configure Go environment
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Set Go proxy if provided
RUN if [ -n "${GO_PROXY}" ]; then \
        go env -w GOPROXY="${GO_PROXY}" && \
        go env -w GOSUMDB=off; \
    fi

# Create app directory
WORKDIR /app

# Copy go.mod and go.sum from root and core first for caching
COPY go.mod .
COPY go.sum .
COPY core/go.mod core/
COPY core/go.sum core/

# Download core dependencies
WORKDIR /app/core
RUN go mod download

# Now copy core source
COPY core /app/core

# Build each Go service
# 1. Admin Dashboard API
WORKDIR /app/admin_dashboard/api
COPY admin_dashboard/api/go.mod .
COPY admin_dashboard/api/go.sum .
RUN go mod download
COPY admin_dashboard/api .
RUN go build -ldflags="-s -w" -o /app/bin/admin_api main.go

# 2. Customer Dashboard API
WORKDIR /app/customer_dashboard/api
COPY customer_dashboard/api/go.mod .
COPY customer_dashboard/api/go.sum .
RUN go mod download
COPY customer_dashboard/api .
RUN go build -ldflags="-s -w" -o /app/bin/customer_api main.go

# 3. Ocserv User Manager
WORKDIR /app/ocserv_user_manager
COPY ocserv_user_manager/go.mod .
COPY ocserv_user_manager/go.sum .
RUN go mod download
COPY ocserv_user_manager .
RUN go build -ldflags="-s -w" -o /app/bin/ocserv_user_manager main.go

# 4. Ocserv Log Parser
WORKDIR /app/ocserv_log_parser
COPY ocserv_log_parser/go.mod .
COPY ocserv_log_parser/go.sum .
RUN go mod download
COPY ocserv_log_parser .
RUN go build -ldflags="-s -w" -o /app/bin/ocserv_log_parser main.go

# 5. Ocserv Telegram Bot
WORKDIR /app/ocserv_telegram_bot
COPY ocserv_telegram_bot/go.mod .
COPY ocserv_telegram_bot/go.sum .
RUN go mod download
COPY ocserv_telegram_bot .
RUN go build -ldflags="-s -w" -o /app/bin/ocserv_telegram_bot main.go


# ====================================================
# Final Stage - Common Runtime Base
# ====================================================
FROM debian:trixie-slim AS base

ARG DEBIAN_MIRROR
ARG DEBIAN_SECURITY_MIRROR

# Configure apt mirrors if provided
RUN if [ -n "${DEBIAN_MIRROR}" ] || [ -n "${DEBIAN_SECURITY_MIRROR}" ]; then \
        debian_mirror="${DEBIAN_MIRROR:-http://deb.debian.org/debian}" && \
        debian_security_mirror="${DEBIAN_SECURITY_MIRROR:-http://deb.debian.org/debian-security}" && \
        echo "deb ${debian_mirror} trixie main contrib non-free non-free-firmware" > /etc/apt/sources.list && \
        echo "deb ${debian_mirror} trixie-updates main contrib non-free non-free-firmware" >> /etc/apt/sources.list && \
        echo "deb ${debian_security_mirror} trixie-security main non-free-firmware" >> /etc/apt/sources.list \
    ; fi

# Install runtime dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        && \
    rm -rf /var/lib/apt/lists/*

# ====================================================
# Individual Service Images
# ====================================================
FROM base AS admin_api
COPY --from=builder --chmod=755 /app/bin/admin_api /usr/local/bin/
CMD ["admin_api", "serve"]

FROM base AS customer_api
COPY --from=builder --chmod=755 /app/bin/customer_api /usr/local/bin/
CMD ["customer_api", "serve"]

FROM base AS ocserv_user_manager
COPY --from=builder --chmod=755 /app/bin/ocserv_user_manager /usr/local/bin/
# Volume for cron journal
VOLUME ["/app/cron_journal"]
CMD ["ocserv_user_manager", "serve"]

FROM base AS ocserv_log_parser
COPY --from=builder --chmod=755 /app/bin/ocserv_log_parser /usr/local/bin/
# Docker socket needed for reading container logs (if in Docker mode)
VOLUME ["/var/run/docker.sock:/var/run/docker.sock:ro"]
CMD ["ocserv_log_parser", "serve"]

FROM base AS ocserv_telegram_bot
COPY --from=builder --chmod=755 /app/bin/ocserv_telegram_bot /usr/local/bin/
CMD ["ocserv_telegram_bot", "serve"]
