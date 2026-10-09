---
name: docker-development
description: "Use when working with ANY Docker task: writing Dockerfiles, configuring docker-compose/compose.yml, multi-stage builds, docker-bake.hcl, container security audits, .dockerignore optimization, or CI/CD container testing. Triggers on: Dockerfile, docker-compose, container, image build, multi-stage, docker bake, compose."
license: "(MIT AND CC-BY-SA-4.0)"
compatibility: "Requires docker, docker compose."
metadata:
  version: "1.15.2"
  repository: "https://github.com/netresearch/docker-development-skill"
  author: "Netresearch DTT GmbH"
allowed-tools:
  - "Bash(docker:*)"
  - "Bash(grep:*)"
  - "Read"
  - "Write"
  - "Glob"
  - "Grep"
---

# Docker Development

## Core Principles

1. **Minimal** -- Alpine/distroless, multi-stage
2. **Secure** -- Non-root USER, no layer secrets, pin versions
3. **Testable** -- entrypoint bypass, DNS mocking
4. **Cache-efficient** -- deps first, clean in-layer

## Quick Reference

### Multi-Stage Build (Node.js)

```dockerfile
FROM node:24-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .

FROM node:24-alpine
RUN addgroup -g 1001 app && adduser -u 1001 -G app -D app
USER app
COPY --from=builder /app .
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:3000/health || exit 1
CMD ["node", "server.js"]
```

### Multi-Stage Build (Go -- scratch/distroless)

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server .

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/server /server
CMD ["/server"]
```

### Layer Optimization

```dockerfile
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl && \
    rm -rf /var/lib/apt/lists/*
```

### Build Cache: Copy Dependency Files First

```dockerfile
COPY package*.json ./
RUN npm ci
COPY . .
```

Manifests before source keeps install layers cached.

### BuildKit Secrets

```dockerfile
RUN --mount=type=secret,id=ssh_key,dst=/root/.ssh/id_rsa git clone git@github.com:org/repo.git
```

`ARG` leaks via `docker history` **and** SLSA provenance --
`references/build-secret-leaks.md`

### Docker Bake (Multi-Platform)

```hcl
target "app" {
  platforms = ["linux/amd64", "linux/arm64"]
  cache-from = ["type=gha"]
  cache-to = ["type=gha,mode=max"]
}
```

## Security Anti-Patterns

| Anti-pattern | Fix |
|---|---|
| `FROM image:latest` | Pin version: `image:1.2.3-alpine` |
| No `USER` directive | `adduser` + `USER appuser` |
| `chmod 777` | Use specific permissions: `chmod 550` |
| `privileged: true` in compose | Remove or use specific `cap_add` |
| `volumes: [/:/host]` | Mount only needed paths |
| `ports: ["0.0.0.0:3000:3000"]` | Bind to `127.0.0.1:3000:3000` |
| `ENV DB_PASSWORD=secret` | Use `--mount=type=secret` or compose secrets |

## CI Testing Gotchas

1. **Bypass entrypoint**: `docker run --rm --entrypoint php myimage -v`
2. **Mock upstream DNS**: `docker run --rm --add-host backend:127.0.0.1 nginx-image nginx -t`
3. **Compose validation**: `cp .env.example .env` before `docker compose config`
4. **Secret scanning**: exclude `.env.example`, README, docs
5. **Root-owned artifacts**: bind-mount dirs (`EACCES`) -- `references/bind-mount-ownership.md`

## .dockerignore

Exclude: `.git`, `node_modules`/`vendor`, `.env*`, `*.pem`, `*.key`

## Compose Essentials

- startup ordering: `depends_on.condition: service_healthy` + `healthcheck` `start_period`
- `networks.internal: true` isolates databases
- `profiles: [debug]`: start only with `--profile debug`
- shared image ref: define ONCE per file as top-level extension field + anchor -- `x-app-image: &app-image registry/app:${APP_IMAGE_VERSION:-85}`, services use `image: *app-image`. The field must sit ABOVE `services:` — an alias is only valid after its anchor in document order. `:-` defaults cover unset AND empty vars (a bare omitted tag silently resolves `:latest`). Anchors are file-local: every overlay file needs its own. Verify both paths: `APP_IMAGE_VERSION= docker compose config` and with an override

## References

- `references/ci-testing.md` -- CI testing patterns for Docker images
- `references/dind-testing-patterns.md` -- Docker-in-Docker testing patterns
- `references/bind-mount-ownership.md` -- root-owned bind-mount artifacts
- `references/gpg-verification.md` -- gpgv patterns; stale keybox locks
- `references/registry-catalogue-and-pin-rot.md` -- catalogue probes; pin rot
- `references/build-secret-leaks.md` -- `ARG` in provenance
- `references/php-fpm-worker-starvation.md` -- keepalive pins php-fpm
