# Go Project Rules

## Priority

When instructions conflict, use this precedence:

1. Existing project behavior and explicit user requirements.
2. Rules in applicable project skills.
3. Existing conventions visible in nearby files.
4. Idiomatic Go conventions.
5. Personal implementation preference.

Do not invent project conventions when the repository can be inspected.

## Architecture

Controllers handle HTTP input/output, binding, request validation, application error mapping, and response formatting. They must not perform database queries, business logic, repository behavior, or complex usecase transformations.

Usecases handle business rules, application orchestration, transactions when required, repository calls, and domain validation beyond request shape. They must not depend directly on Echo request/response behavior.

Repositories handle database reads and writes, GORM queries, persistence filtering, and persistence-specific error translation. They must not generate HTTP responses or contain controller logic.

Use the project's `*context.Context` convention for service, usecase, and repository operations where applicable. Do not replace it with plain `context.Context` unless explicitly requested.

## Workflow

For coding tasks:

1. Inspect relevant existing files and identify the controller-usecase-repository flow.
2. Search for reusable helpers before creating new ones.
3. Search `config/errors.json` before assigning an application error code.
4. Implement the smallest coherent change.
5. Add or update focused tests.
6. Format modified Go files.
7. Run relevant tests and broader checks when practical.
8. Report modified files and unresolved issues.

Do not copy shared behavior into multiple packages or over-generalize code used once.

## Source Organization

Within Go files, keep declarations ordered as closely as practical to imports, constants, variables, structs/types, interfaces, public functions, and private functions. Do not reorganize unrelated code solely for ordering.

Run `gofmt` on every modified Go file. Use `goimports` when the repository already uses it.

Add comments for business rules, important constraints, and non-obvious behavior. Avoid comments that restate obvious code.

## Project Layout

Use the existing repository layout first. Established locations include:

```text
cmd/
config/
internal/migrations/
internal/models/
internal/repository/{entity-name}/
internal/service/{service-name}/
internal/usecase/
internal/provider/routing/
pkg/bootstrap/
pkg/middlewares/
pkg/request/
pkg/routing/
tests/model/
tests/units/
tests/integration/
```

Do not create new top-level architectural layers without a clear need.

## Context

Pass context through `Controller -> Usecase -> Repository`. Do not replace an incoming request context with `context.Background()` or `context.TODO()` inside a request flow without an explicit reason.

When using GORM, propagate context where supported:

```go
db.WithContext(*ctx)
```

## Change Discipline

Do not refactor unrelated packages, rename unrelated symbols, reformat untouched files, upgrade dependencies, or change public API/database behavior unless required.

Before generating code, inspect existing examples of controllers, usecases, repositories, models, pagination, error responses, authentication, middleware, migrations, UUID generation, and tests. Inspect helper implementations or usages instead of assuming their API from their name.

If repository evidence is contradictory, follow the most recent and widely used convention and report the discrepancy.
