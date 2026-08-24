package units

import (
	"encoding/json"
	"testing"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/bootstrap"
	"github.com/stretchr/testify/require"
)

func TestFreshMigrationSetIsCompleteAndUnique(t *testing.T) {
	require.Len(t, bootstrap.Migrations, 5)
	require.Equal(t, "005_add_ocserv_user_expiry_modes", bootstrap.Migrations[4].ID)

	seen := make(map[string]struct{}, len(bootstrap.Migrations))
	for _, migration := range bootstrap.Migrations {
		require.NotNil(t, migration)
		require.NotEmpty(t, migration.ID)
		require.NotNil(t, migration.Migrate)
		require.NotNil(t, migration.Rollback)
		_, exists := seen[migration.ID]
		require.False(t, exists, "duplicate migration ID %q", migration.ID)
		seen[migration.ID] = struct{}{}
	}
}

func TestEntityResponsesExposeIDOnly(t *testing.T) {
	for _, entity := range []any{
		models.User{ID: 7, Username: "admin"},
		models.OcservUser{ID: 9, Username: "vpn-user"},
	} {
		data, err := json.Marshal(entity)
		require.NoError(t, err)
		require.Contains(t, string(data), `"id":`)
	}
}
