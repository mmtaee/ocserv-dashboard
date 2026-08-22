# Persistence Rules

## Repositories

Repositories handle database reads and writes, GORM queries, persistence-specific filtering, and persistence-specific error translation when required. They must not generate HTTP responses, depend on Echo, contain controller logic, or hold unrelated business rules.

Use GORM unless a raw query is clearly required. Prefer readable, explicit queries over clever chaining.

## Models

Target PostgreSQL 18. Use UUIDv7 primary keys instead of new auto-incrementing integer IDs unless an existing schema explicitly requires otherwise. Reuse the project's UUIDv7 helper and do not introduce a second UUID library.

Persistent and request model fields should carry the tags required by their purpose. Use applicable `json`, `gorm`, and `validate` tags without inventing validation rules merely to make a tag non-empty.

```go
Name string `json:"name" gorm:"type:varchar(255);not null" validate:"required,max=255"`
```

Follow existing naming and nullable-field conventions.

## PostgreSQL and GORM

Use PostgreSQL-compatible SQL and behavior. Pay particular attention to UUIDs, JSON/JSONB, timestamps, indexes, partial indexes, generated columns, enums, `ON CONFLICT`, and identifier quoting. Do not use MariaDB/MySQL-specific SQL.

## Migrations

Migrations live in `internal/migrations/` and use `go-gormigrate/gormigrate/v2`. Follow the numeric filename convention, use the next unused prefix, and register each migration in the existing bootstrap, normally `pkg/bootstrap/migrate.go`.

Before creating a migration:

1. Inspect existing filenames and determine the next prefix.
2. Inspect registration order and nearby migration style.
3. Write PostgreSQL-compatible migration and rollback behavior.
4. Register the migration in order.

Use idempotent raw SQL where PostgreSQL safely supports it, such as `CREATE INDEX IF NOT EXISTS`, but do not use guards that hide schema errors.

Do not edit an already-applied historical migration for a new schema change. Create a new migration. Define rollback behavior when safely possible.

## Transactions

Use a transaction when multiple writes must succeed atomically. Transaction boundaries normally belong in the usecase/application layer, not controllers. Avoid nested or fragmented transactions, and ensure each participating repository operation uses the same transaction handle.

## Verification

Verify that each migration compiles, is registered in the correct order, uses valid PostgreSQL syntax, and has a practical rollback.
