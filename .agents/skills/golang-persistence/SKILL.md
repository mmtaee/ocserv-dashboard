---
name: golang-persistence
description: Work on PostgreSQL and GORM persistence in this Go repository, including models, repositories, migrations, transactions, UUID primary keys, schema changes, and database-related review.
---

# Go Persistence

Use this skill with `golang` for models, repositories, GORM queries, migrations, database schemas, transactions, or persistence behavior.

Keep persistence concerns in repositories, target PostgreSQL 18, use the project's existing GORM and UUIDv7 conventions, and make schema changes through new numbered go-gormigrate migrations.

## Detailed Guidance

Read [references/persistence.md](references/persistence.md) before changing models, queries, schemas, migrations, or transaction behavior. It contains model tags, PostgreSQL constraints, migration registration, rollback, and transaction rules.
