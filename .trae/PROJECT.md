# Project: Ocserv Dashboard
## Overview
A web-based dashboard to manage an OpenConnect VPN server (ocserv), including user/group management, monitoring, statistics, and an integrated Telegram bot for customer self-service.

## Tech Stack
### Backend
- **Language**: Go 1.25+
- **Framework**: Echo v5
- **ORM**: GORM
- **Database**: PostgreSQL
- **Migrations**: Gormigrate v2
- **Validation**: Validator v10
- **CLI**: Cobra

### Frontend
- **Framework**: Vue 3
- **Build Tool**: Vite
- **UI**: Custom components

### Infrastructure
- **Containerization**: Docker, Docker Compose
- **Deployment Options**: Docker-based or Systemd-based
- **Supported OS**: Debian 12+, Ubuntu 20.04+

## Project Structure
```
ocserv-dashboard/
├── .trae/                                    # TRAE AI configuration
│   ├── PROJECT.md                            # Project context file (MUST be updated after any change to file/directory structure)
│   └── skills/
│       └── backend/                          # Backend-specific TRAE skills
│           ├── api-creator/
│           ├── master-rules/
│           ├── middleware-creator/
│           ├── model-creator/
│           ├── service-creator/
│           └── test-creator/
├── core/                                     # Core shared code (formerly services/common)
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/                                  # Migration command
│   │   └── migrate.go
│   ├── migrations/                           # Gormigrate database migrations (001-013)
│   │   ├── 012_remove_uid_and_add_administrators.go  # Removed UID, added Administrators and OwnerAdminID
│   │   └── 013_add_system_client_profile_columns.go  # Added ClientProfileServerAddress/Port/ConnectionName to System
│   ├── models/                               # Shared GORM models
│   │   ├── admin.go                          # Administrator, AdministratorToken
│   │   ├── common.go
│   │   ├── occtl.go
│   │   ├── ocserv_group.go                   # Added OwnerAdminID
│   │   ├── ocserv_user.go                    # Removed UID, added OwnerAdminID
│   │   ├── telegram.go                       # Added OwnerAdminID
│   │   └── telegram_languages.go
│   ├── ocserv/                               # Ocserv-specific utilities
│   │   ├── group/
│   │   │   ├── group.go
│   │   │   └── types.go
│   │   ├── occtl/
│   │   │   └── occtl.go
│   │   └── user/
│   │       ├── types.go
│   │       ├── user.go
│   │       └── utils.go
│   └── pkg/                                  # Shared packages
│       ├── config/
│       │   └── config.go                     # Init() now gets all params from env vars (DEBUG, HOST, PORT)
│       ├── database/
│       │   └── database.go
│       ├── logger/
│       │   ├── service.go
│       │   └── types.go
│       ├── testutils/                        # Test utilities for model tests
│       │   └── testutils.go
│       └── utils/
│           └── utils.go
├── dashboard/
│   ├── api/                                  # Main API service
│   │   ├── cmd/                              # CLI commands
│   │   │   ├── root.go
│   │   │   └── serve.go
│   │   ├── config/
│   │   │   ├── config.go                     # Wrapper around core config
│   │   │   └── errors.json                   # Error codes and messages
│   │   ├── internal/
│   │   │   ├── repository/
│   │   │   │   └── admin.go                  # AdminRepository
│   │   │   ├── usecase/
│   │   │   │   └── admin.go                  # AdminUseCase (Login, GetProfile, ChangePassword)
│   │   │   ├── service/
│   │   │   │   └── auth/
│   │   │   │       ├── types.go               # Request/response types for auth service
│   │   │   │       ├── controller.go          # HTTP handlers for auth service
│   │   │   │       └── routes.go             # Route registration for auth service
│   │   │   └── providers/
│   │   │       └── routing/
│   │   │           └── routing.go          # Aggregate route registration
│   │   ├── pkg/
│   │   │   ├── auth/
│   │   │   │   └── jwt.go                    # Claims, CreateAdministratorToken, ValidateAdministratorToken
│   │   │   ├── bootstrap/
│   │   │   │   ├── migration.go              # Uses core migrations
│   │   │   │   └── serve.go                  # Initializes config, infra, migrations and runs server
│   │   │   ├── infra/
│   │   │   │   └── infra.go                  # No Redis, just DB
│   │   │   ├── middlewares/
│   │   │   │   ├── auth.go                   # Admin JWT auth middleware
│   │   │   │   ├── timeout.go                # Request timeout middleware
│   │   │   │   └── ratelimit.go              # In-memory rate limiter
│   │   │   ├── request/
│   │   │   │   ├── errors.go
│   │   │   │   ├── pagination.go
│   │   │   │   ├── response.go
│   │   │   │   └── validator.go
│   │   │   ├── routing/
│   │   │   │   ├── serve.go
│   │   │   │   └── utils.go
│   │   │   └── testutils/
│   │   │       └── db_loader.go
│   │   ├── main.go
│   │   └── go.mod
│   └── ui/                                   # Vue 3 frontend (empty for now)
├── docs/                                     # Project documentation & assets
│   ├── home.png
│   ├── home_stats.png
│   ├── home_sub.png
│   ├── logo.png
│   ├── menu.png
│   └── telegram-translations.md
├── .dockerignore
├── .env
├── .env.sample
├── .gitguardian.yaml
├── .gitignore
├── LICENSE
├── note
└── README.md
```

## Key Conventions
### Backend
- **No PUT methods**: Use POST/PATCH/DELETE instead
- **Error handling**: Use unique error codes from config/errors.json
- **Testing**: Model tests use in-memory SQLite; usecase tests use mocks/fakes; integration tests use Echo test harness

### Development Workflow
- **Start project**: Check README.md for instructions
- **API docs**: Generated via Swag

## Important Files
- **README.md**: General project documentation
- **TODO.md**: Roadmap & planned features
- **docker-compose.yml**: Docker deployment configuration (if present)
- **core/go.mod**: Go dependencies for core shared code
