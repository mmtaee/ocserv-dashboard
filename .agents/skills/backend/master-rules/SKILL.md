---
name: master-rules
description: Unified global rules for this Go backend project, covering architecture, implementation, testing, and best practices.
---

## Core Philosophy

- **Clean Architecture**: Follow Clean Architecture principles: Presentation → Service → Usecase → Repository → Infrastructure.
- **Consistency**: Maintain consistency with existing codebase patterns before introducing new approaches.
- **Test-First**: Write tests alongside feature implementation.
- **Education**: Explain architectural choices and implementation decisions clearly.

## Preflight Checklist

1. **Read Project Structure**: Explore the codebase to understand existing patterns.
2. **Check Dependencies**: Review `go.mod` for existing dependencies before adding new ones.
3. **Follow Conventions**: Use existing naming, directory structure, and error handling patterns.
4. **Update PROJECT.json**: After ANY change (create/delete/move/rename files/directories), update `/.agents/PROJECT.json`.
5. **Check Role Rules**: Review `/.agents/role-rules.json` before implementing any new features, and update it if new roles or permissions are added.
6. **Enforce Ownership**: All models that belong to an admin must have an `OwnerAdminID` field, and non-super admins can only access their own objects.

## Role-Based Access Control (RBAC)

### Core Rules
- **Super Admin**: Has full access to all objects and operations (permissions: `*`).
- **Regular Admin**: Can only access, modify, and delete objects they own (where `OwnerAdminID` matches their admin ID).

### Key Requirements for New Features
1. **Ownership Field**: Every new domain model that belongs to an admin must include an `OwnerAdminID` field.
2. **Repository Layer**: All repository methods must accept an `adminID` and `role` parameter.
3. **Ownership Checks**: In repository methods, for non-super admins, always add a `WHERE owner_admin_id = ?` condition to queries.
4. **Role Propagation**: Pass the admin ID and role from the controller → use case → repository layer.
5. **Update Role Rules**: If you add new roles or permissions, update `/.agents/role-rules.json` immediately.


## Project Stack & Architecture

This is a Go backend project using:
- **Framework**: Echo v5
- **Database**: GORM (MariaDB)
- **Migrations**: Gormigrate v2
- **CLI**: Cobra
- **Validation**: Validator v10

### Directory Structure
```
cmd/                → CLI commands (root, serve, docs)
internal/           → Business logic
  service/          → HTTP handlers (controllers) and routes
  usecase/          → Business logic orchestration
  repository/       → Data access layer
  provider/         → Dependency injection and route aggregation
config/             → Configuration and errors.json
pkg/                → Shared modules (bootstrap, middlewares, request, testutils)
```

## Code Style & Naming

- **Naming**: Use PascalCase for exported types/functions, camelCase for unexported.
- **Imports**: Group imports as: standard library → third-party → project internal.
- **Errors**: Use `config/errors.json` for unique error codes; no code reuse across different errors.
- **HTTP Methods**: Do NOT use PUT; use POST/PATCH/DELETE as appropriate.
- **Echo v5**: 
  - Use `c *echo.Context` (not `c echo.Context`)
  - Retrieve values from context with: `value := c.Get("key").(Type)`
  - Example: `userID := c.Get("user_id").(uint)`

## Testing Guidelines

- **Model Tests**: SQLite in-memory via `pkg/testutils.SetupTestDB`; cover CRUD and constraints.
- **Usecase Tests**: Fake or mock repository implementations; no real DB.
- **Integration Tests**: Echo test harness; mock usecase layer; no full server.
- **Run Tests**: Use `go test ./...` to verify everything passes.

## Documentation
- **Swagger/OpenAPI**: Document endpoints with Swagger comments; run `swag init --pd` to regenerate docs. **CRITICAL**: Swagger @Router comment MUST NOT include /api/v1/ prefix. For example, use // @Router /auth/login [post], NOT /api/v1/auth/login.
- **Comments**: Add comments for non-trivial logic explaining "why" not just "what".
- **Pagination Responses**: For paginated list responses, use the structure with `meta` (containing pagination) and `result` (containing the items):
```go
type <Name>Response struct {
	Meta   request.Pagination `json:"meta"`
	Result []<ModelType>      `json:"result"`
}
```
- **CRITICAL**: NEVER use pointers to arrays/slices (like `*[]<Type>`) anywhere in the code. Always use direct slices (like `[]<Type>`).

## Finalization Protocol

Before completing a task:
1. **Format**: Ensure code is properly formatted (`gofmt` or equivalent).
2. **Test**: Run `go test ./...` to confirm all tests pass.
3. **Verify**: Ensure the implementation aligns with existing patterns and requirements.
4. **Update project-structure.json**: Update `/.agents/project-structure.json` to reflect any changes to the project structure or files.
5. **Update role-rules.json**: If new roles or permissions were added, update `/.agents/role-rules.json` accordingly.
6. **Check Ownership**: Double-check that all new models have an `OwnerAdminID` field and that ownership checks are implemented in the repository layer.
