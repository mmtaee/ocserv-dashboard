package units

import (
	"encoding/json"
	"testing"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/bootstrap"
	"github.com/stretchr/testify/require"
)

func TestFreshMigrationSetIsCompleteAndUnique(t *testing.T) {
	master := bootstrap.MigrationsFor(false)
	agent := bootstrap.MigrationsFor(true)
	require.Len(t, master, 8)
	require.Len(t, agent, 8)
	require.Equal(t, "008_create_ocserv_agents", master[7].ID)
	require.Equal(t, "009_create_agent_tokens", agent[7].ID)

	seen := make(map[string]struct{}, len(master))
	for _, migration := range master {
		require.NotNil(t, migration)
		require.NotEmpty(t, migration.ID)
		require.NotNil(t, migration.Migrate)
		require.NotNil(t, migration.Rollback)
		_, exists := seen[migration.ID]
		require.False(t, exists, "duplicate migration ID %q", migration.ID)
		seen[migration.ID] = struct{}{}
	}
	require.NotContains(t, seen, "009_create_agent_tokens")

	seen = make(map[string]struct{}, len(agent))
	for _, migration := range agent {
		_, exists := seen[migration.ID]
		require.False(t, exists, "duplicate migration ID %q", migration.ID)
		seen[migration.ID] = struct{}{}
	}
	require.NotContains(t, seen, "008_create_ocserv_agents")
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
