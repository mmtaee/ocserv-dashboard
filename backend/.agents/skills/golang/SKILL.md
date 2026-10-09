---
name: golang
description: Implement, modify, review, and test Go code in this repository using its established stack and controller-usecase-repository architecture. Use for general Go work and pair with the specialized Go skills when applicable.
---

# Go Project Rules

Preserve the repository's existing behavior, architecture, and conventions. Keep changes scoped to the user's request and inspect nearby implementations before introducing new patterns.

Use the established stack: Go 1.27+, PostgreSQL 18, Echo v5, GORM, go-gormigrate v2, Cobra, and validator v10. Do not replace these technologies or add major dependencies unless explicitly requested.

Maintain this dependency flow:

```text
Controller -> Usecase -> Repository -> Database
```

Pass request context through the full flow, use existing project helpers, format modified Go files, and run relevant tests.

## Detailed Guidance

Read [references/project-rules.md](references/project-rules.md) before implementing or reviewing Go changes that require project layout, architecture, context, dependency, or workflow decisions.

Use the specialized skill alongside this one when relevant:

- `golang-api` for controllers, routes, validation, Swagger, pagination, middleware, and HTTP errors.
- `golang-persistence` for models, repositories, GORM, migrations, transactions, and UUIDs.
- `golang-runtime` for Cobra commands, bootstrap, dependency wiring, route providers, and shutdown.
- `golang-testing` for tests, formatting, vetting, and completion checks.
