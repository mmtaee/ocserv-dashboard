package units

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/services/worker/connectionexpiry"
	"github.com/stretchr/testify/require"
)

type firstConnectionRepository struct {
	mu   sync.Mutex
	user models.OcservUser
}

func (r *firstConnectionRepository) RecordFirstConnection(_ context.Context, username string, connectedAt time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if username != r.user.Username || r.user.ExpiryMode != models.ExpiryModeFirstConnection || r.user.FirstConnectedAt != nil {
		return false, nil
	}
	r.user.FirstConnectedAt = timePointer(connectedAt.UTC())
	r.user.ExpireAt = timePointer(connectedAt.UTC().AddDate(0, 0, *r.user.ExpireDaysAfterFirstConnection))
	return true, nil
}

func TestWorkerRecordsFirstConnectionAndCalculatesExpiryOnce(t *testing.T) {
	repository := &firstConnectionRepository{user: models.OcservUser{
		Username: "alice", ExpiryMode: models.ExpiryModeFirstConnection, ExpireDaysAfterFirstConnection: intPointer(30),
	}}
	observer := connectionexpiry.New(repository)
	first := time.Date(2026, time.September, 1, 12, 15, 0, 0, time.UTC)
	later := first.Add(48 * time.Hour)
	line := "ocserv[100]: main[alice]:192.0.2.10:443 new user session"

	require.NoError(t, observer.Observe(context.Background(), line, first))
	require.NotNil(t, repository.user.FirstConnectedAt)
	require.Equal(t, first, *repository.user.FirstConnectedAt)
	require.Equal(t, first.AddDate(0, 0, 30), *repository.user.ExpireAt)

	require.NoError(t, observer.Observe(context.Background(), line, later))
	require.Equal(t, first, *repository.user.FirstConnectedAt)
	require.Equal(t, first.AddDate(0, 0, 30), *repository.user.ExpireAt)
}

func timePointer(value time.Time) *time.Time {
	return &value
}
