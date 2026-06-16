# Ocserv Dashboard API

## Overview

This is the backend API for the Ocserv Dashboard project.

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL 12+

### Environment Setup

1. Copy the `.env.sample` file from the project root:
   ```bash
   cp ../../.env.sample .env
   ```

2. Edit the `.env` file with your configuration:

| Variable | Description | Default |
|----------|-------------|---------|
| DEBUG | Enable debug mode | false |
| HOST | API host | 0.0.0.0 |
| PORT | API port | 8080 |
| SECRET_KEY | Secret key for internal use | SECRET_KEY122456 |
| JWT_SECRET | JWT signing secret | secret1234 |
| ALLOW_ORIGINS | CORS allowed origins (comma-separated) | - |
| POSTGRES_HOST | PostgreSQL host | localhost |
| POSTGRES_PORT | PostgreSQL port | 5432 |
| POSTGRES_USER | PostgreSQL user | ocserv |
| POSTGRES_PASSWORD | PostgreSQL password | ocserv-passwd |
| POSTGRES_DB | PostgreSQL database name | ocserv_db |
| POSTGRES_SSLMODE | PostgreSQL SSL mode | disable |

### Commands

#### Create Super Admin

Creates a new super admin user. Only one super admin can exist.

```bash
go run main.go create-super-admin -u <username> -p <password>
```

Example:
```bash
go run main.go create-super-admin -u admin -p myStrongPassword123
```

#### Serve API

Starts the API server.

```bash
go run main.go serve
```

### API Documentation

Swagger documentation is available at `/swagger/index.html` when the server is running.

To regenerate Swagger docs:
```bash
swag init --pd
```

## Development

### Run Tests

```bash
go test ./...
```

### Build

```bash
go build -o api
```
