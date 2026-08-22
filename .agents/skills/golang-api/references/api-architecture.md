# API Architecture

## Controllers

Controller flow should generally be:

```text
Bind request -> Validate request -> Call usecase -> Handle application error -> Return response
```

Controllers read path, query, header, or body input; bind and validate requests; call usecases; map known application errors to HTTP responses; and return API responses. They must not execute database queries, contain business logic, implement repository behavior, or perform complex usecase transformations.

## HTTP Methods

Use `GET` for reads, `POST` for creation or actions, `PATCH` for partial updates, and `DELETE` for deletion. Never create a new endpoint using `PUT`.

## Request Validation

Use the project's request helpers:

```go
request.Validator(...)
request.BadRequest(...)
```

Use validator v10 tags on request DTOs. Request-shape validation belongs at the controller boundary; business validation belongs in the usecase.

## Error Codes

Application error codes are defined in `config/errors.json`. Before referencing a new code:

1. Confirm that an equivalent semantic error does not exist.
2. Add a unique code to `config/errors.json`.
3. Reference it from Go code using `ResponseWithCode` according to project conventions.

Never reuse an unrelated code, invent an unregistered code, duplicate a code, or change an existing code's meaning.

## Swagger

Do not use `// @Security ApiKeyAuth`. Use:

```go
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 401 {object} request.ErrorResponse
```

Annotations must match the actual method, route, body, query parameters, response types, and status codes.

## Pagination and Echo

Use `pkg/request/pagination` for paginated APIs. Do not implement duplicate pagination parsing.

Use Echo v5 imports and APIs:

```go
import "github.com/labstack/echo/v5"
```

Do not introduce Echo v4 imports.

## Middleware

Middleware belongs in `pkg/middlewares/`. Every middleware must include a usage comment showing registration, for example:

```go
// Usage: e.Use(middlewares.RequestID())
```

Follow the repository's Echo v5 context convention, including `*echo.Context` where required. Keep middleware concurrency-safe, reusable, stateless when practical, and free of endpoint-specific business logic.

## Routing

Service-specific routes belong in `internal/service/{service-name}/router.go`. Global registration belongs in the established routing provider, normally `internal/provider/routing/routes.go` or `pkg/routing/router.go`.

Do not register a route twice. Inspect route groups, middleware, version prefixes, authentication requirements, and naming conventions before adding one.

## New Endpoint Checklist

- Route added with no new `PUT` usage.
- Swagger and authorization annotations match behavior.
- Request DTO is added or reused and validated with project helpers.
- Controller contains only HTTP concerns.
- Business logic is in a usecase and persistence logic is in a repository.
- Context is propagated.
- Shared pagination is used when applicable.
- New application errors are uniquely registered.
- Schema changes use a numbered and registered migration.
- PostgreSQL syntax is used.
- Tests are added, code is formatted, and relevant checks are run.
