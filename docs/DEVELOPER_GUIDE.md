# Developer Guide

Welcome to the Ocserv Users Management project! This guide will help you get started with developing services and web features.

## Table of Contents

1. [Project Structure](#project-structure)
2. [Backend Development](#backend-development)
3. [Frontend Development](#frontend-development)
4. [Development Workflow](#development-workflow)
5. [Coding Standards](#coding-standards)
6. [Important Guidelines](#important-guidelines)

---

## Project Structure

```
ocserv-users-management/
├── configs/              # Configuration files
├── docker/               # Docker files
├── docs/                 # Documentation
├── scripts/              # Deployment and setup scripts
├── services/             # Backend Go services
│   ├── api/              # Main API service
│   ├── common/           # Shared packages
│   ├── log_stream/       # Log streaming service
│   ├── telegram_bot/     # Telegram bot service
│   ├── user_expiry/      # User expiry service
│   └── webhook/          # Webhook service
└── web/                  # Frontend Vue.js application
```

---

## Backend Development

### Prerequisites

- Go 1.25 or later
- Docker and Docker Compose (for development)

### API Service (services/api)

#### Generate Swagger Documentation

After making changes to API endpoints, regenerate the Swagger documentation:

```bash
cd services/api
swag init --pd
```

This generates the Swagger JSON/YAML files in `services/api/docs/`.

#### Export OpenAPI Schema

To generate the OpenAPI schema file for frontend code generation:

```bash
cd services/api
go run main.go docs
```

### Common Packages (services/core)

Shared packages used by all backend services:
- `models/`: Data models
- `pkg/config/`: Configuration management
- `pkg/database/`: Database connections (PostgreSQL only)
- `pkg/logger/`: Logging utilities

---

## Frontend Development

### Prerequisites

- Node.js 24 or later
- Yarn package manager

### Setup

1. Navigate to the web directory:
```bash
cd web
```

2. Install dependencies:
```bash
yarn install
```

### Development Server

Start the development server:
```bash
yarn dev
```

The application will be available at `http://localhost:5173`.

### ⚠️ IMPORTANT: API Client Usage

**Never use direct axios calls!** Always use the generated API client from `@/api`.

#### Step 1: Generate OpenAPI Schema

First, make sure the backend is running or generate the schema:
```bash
cd services/api
go run main.go docs
```

#### Step 2: Generate TypeScript Client

Then, in the web directory, regenerate the API client:
```bash
cd web
yarn codegen
```

This creates TypeScript types and API client classes in `web/src/api/`.

#### Step 3: Use the API Client

Import and use the generated API classes:

```typescript
import { OcservUsersApi, type OcservUserCreateOcservUserData } from '@/api';
import { getAuthorization } from '@/utils/request';

// Create API instance
const api = new OcservUsersApi();

// Call API methods
const createUser = (data: OcservUserCreateOcservUserData) => {
    api.ocservUsersPost({
        ...getAuthorization(),
        request: data
    })
    .then((res) => {
        // Handle success
    });
};
```

### Type Checking

Run TypeScript type checking:
```bash
yarn typecheck
```

### Code Formatting

**Always run this after completing any web UI code changes!**

```bash
yarn format
```

### Build for Production

Create a production build:
```bash
yarn build
```

---

## Development Workflow

### Using Docker for Development

1. Copy the sample environment file:
```bash
cp .env.sample .env
```

2. Edit `.env` and configure your settings (including development mirrors if needed)

3. Start the development stack:
```bash
sudo docker compose up --build
```

### Database Migrations

Migrations are run automatically when the API service starts.

---

## Coding Standards

### Backend (Go)

- Follow Go standard conventions
- Use `gofmt` for code formatting
- Write tests for new features
- Keep functions small and focused

### Frontend (Vue.js / TypeScript)

- Use TypeScript for all new code
- Follow Vue 3 Composition API style
- Use Prettier for code formatting (`yarn format`)
- Keep components focused and reusable
- Use i18n for all user-facing text

---

## Important Guidelines

1. **Swagger Documentation**: Always run `swag init --pd` after modifying API endpoints
2. **OpenAPI Schema**: Run `go run main.go docs` in services/api to export schema
3. **Code Generation**: Always run `yarn codegen` in the web directory after API changes
4. **API Client Usage**: ⚠️ **CRITICAL** - Never use direct axios! Always use the generated client from `@/api`
5. **Formatting**: Always format code before committing
6. **Environment Variables**: Never commit `.env` file - use `.env.sample` as template
7. **SQLite Support**: SQLite has been removed - use PostgreSQL exclusively

---

## Getting Help

If you encounter any issues or have questions, please check the project's issue tracker or contact the maintainers.
