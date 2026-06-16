---
name: model-creator
description: Create GORM models, SQL migrations, and model unit tests. **MANDATORY** ALWAYS read master-rules before implementation.
---

# Context
**CRITICAL**: You MUST read and strictly follow master-rules before starting any task. Every decision (naming, structure, errors, pagination) must align with these rules.

# Requirements
- **Required Fields**:
  - ID (BIGINT AUTO_INCREMENT PK)
  - `CreatedAt` and `UpdatedAt` timestamps
  - `OwnerAdminID` (uint) - for any model that belongs to an admin, to enforce ownership rules
- Tags: `json`, `gorm`, `validate`.
- Migrations: `core/migrations/` using `gormigrate` v2.
- Tests: Use `core/pkg/testutils` for in-memory SQLite if needed.
- **Ownership**: All models that belong to an admin must have `OwnerAdminID` field, and non-super admins can only access their own objects.


# Implementation
1. Add model to `internal/models/`.
2. Create migration file with prefix `00X_`.
3. Register migration in bootstrap.
4. Create model unit test in `internal/tests/models`.

# Model Unit Testing
## File Layout
- Path: `internal/tests/models/<model_name>_test.go`.
- Package: `package tests`.
- Helper: Use `pkg/testutils.SetupTestDB` for DB setup.
- Coverage: CRUD operations, status changes, and validation constraints.

## Example Test
```go
package tests

import (
	"testing"

	"api/internal/models"
	"api/pkg/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookingModel(t *testing.T) {
	db := testutils.SetupTestDB(t, &models.User{}, &models.Trip{}, &models.Booking{})

	user := &models.User{PhoneNumber: "09120000001", Status: models.UserStatusActive}
	require.NoError(t, db.Create(user).Error)

	t.Run("Create", func(t *testing.T) {
		b := &models.Booking{UserID: user.ID}
		assert.NoError(t, db.Create(b).Error)
		assert.NotZero(t, b.ID)
	})

	t.Run("Read with relations", func(t *testing.T) { ... })
	t.Run("Update status", func(t *testing.T) { ... })
	t.Run("Delete", func(t *testing.T) { ... })
	t.Run("Status constants", func(t *testing.T) { ... })
}
```
