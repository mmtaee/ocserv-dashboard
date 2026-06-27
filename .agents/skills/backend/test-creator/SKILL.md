---
name: test-creator
description: Create model unit, usecase unit, and HTTP integration tests for this Go backend project.
---

# Context
This is a Go backend project (module: `api`) using:
- Echo v5 for HTTP
- GORM + MariaDB for persistence
- Clean Architecture (Service → Usecase → Repository)

# Test Types Covered
1. **Model Tests**: SQLite in-memory via `pkg/testutils`
2. **Usecase Tests**: Fake/mock repositories, no real DB
3. **Integration Tests**: Echo test harness, mock usecase layer

---

# 1. Model Unit Tests
Test GORM model behavior.

## File Layout
```
internal/tests/models/<model_name>_test.go
```
Package: `package tests`

## Example
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
		assert.NotEmpty(t, b.UID)
	})

	t.Run("Read with relations", func(t *testing.T) { ... })
	t.Run("Update status", func(t *testing.T) { ... })
	t.Run("Delete", func(t *testing.T) { ... })
}
```

---

# 2. Usecase Unit Tests
Test business logic with fake or mock repositories.

## File Layout
```
internal/tests/usecase/<feature>_test.go
```
Package: `package tests`

## Example with Fake Repo
```go
package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"api/config"
	"api/internal/models"
	clientuc "api/internal/usecase/client"
	"api/pkg/request"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{}
	}
	config.AppConfig.JWTSecret = "test-secret"
}

type fakeRuleRepo struct {
	getLatestFn func() (*models.Rule, error)
}

func (f *fakeRuleRepo) GetLatest(_ context.Context) (*models.Rule, error) {
	return f.getLatestFn()
}

func TestRuleUsecase_GetLatest(t *testing.T) {
	t.Run("returns latest rule", func(t *testing.T) {
		testRule := &models.Rule{
			ID:        1,
			Text:      "Test rules",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo := &fakeRuleRepo{
			getLatestFn: func() (*models.Rule, error) { return testRule, nil },
		}
		uc := clientuc.NewRuleUsecase(repo)
		rule, err := uc.GetLatestRule(context.Background())
		require.NoError(t, err)
		assert.Equal(t, uint(1), rule.ID)
	})

	t.Run("returns error when repo fails", func(t *testing.T) {
		repo := &fakeRuleRepo{
			getLatestFn: func() (*models.Rule, error) { return nil, errors.New("not found") },
		}
		uc := clientuc.NewRuleUsecase(repo)
		_, err := uc.GetLatestRule(context.Background())
		require.Error(t, err)
	})
}
```

---

# 3. HTTP Integration Tests
Test HTTP endpoints with Echo test harness.

## File Layout
- Client endpoints: `internal/tests/<feature>/integration_test.go`
- Admin endpoints: `internal/tests/<feature>/integration_test.go`
Package: `<feature>_test`

## Example
```go
package rule_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api/config"
	"api/internal/models"
	"api/internal/service/client/rule"
	clientuc "api/internal/usecase/client"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRuleUsecase struct {
	clientuc.RuleUsecase
	getLatestFn func(ctx context.Context) (*models.Rule, error)
}

func (m *mockRuleUsecase) GetLatestRule(ctx context.Context) (*models.Rule, error) {
	return m.getLatestFn(ctx)
}

func TestGetLatestRuleIntegration(t *testing.T) {
	t.Run("returns 200 with latest rule", func(t *testing.T) {
		e := echo.New()
		testRule := &models.Rule{
			ID:        1,
			Text:      "Test rules",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockUC := &mockRuleUsecase{
			getLatestFn: func(_ context.Context) (*models.Rule, error) { return testRule, nil },
		}
		ctrl := rule.NewController(mockUC)

		req := httptest.NewRequest(http.MethodGet, "/rules/latest", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ctrl.GetLatestRule(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
```

---

# General Rules
- **No Real DB**: Except model tests using SQLite in-memory
- **No Real External Calls**: Mock/fake everything
- **Use `t.Run`**: For subtests and clarity
- **Module Name**: `api`, imports start with `api/internal/...` or `api/pkg/...`
- **Run Tests**: Use `go test ./...` to verify

# Run Commands
```bash
# All tests
go test ./...

# Specific test categories
go test ./internal/models/tests/...
go test ./internal/usecase/...
go test ./internal/service/...

# Single package
go test ./internal/usecase/client/tests/... -v

# Single test by name
go test ./internal/models/tests/... -run TestBookingModel -v
```
