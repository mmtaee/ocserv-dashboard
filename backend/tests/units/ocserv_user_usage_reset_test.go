package units

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/authz"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	logger "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	ocservuser "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/users"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/worker/logprocessor"
	"github.com/stretchr/testify/require"
)

type usageResetRepository struct {
	ocservuser.Repository
	users map[uint]models.OcservUser
}

func (r *usageResetRepository) GetByID(_ context.Context, id uint) (*models.OcservUser, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, ocservuser.ErrUserNotFound
	}
	return &user, nil
}

func (r *usageResetRepository) ResetUsage(_ context.Context, user *models.OcservUser) (*models.OcservUser, error) {
	r.users[user.ID] = *user
	return user, nil
}

func TestResetUsagePermissionsAndCounters(t *testing.T) {
	tests := []struct {
		name      string
		principal authz.Principal
		userID    uint
		wantError error
	}{
		{"normal user resets own usage", authz.Principal{UserID: 7, Username: "alice"}, 10, nil},
		{"normal user cannot reset another owner", authz.Principal{UserID: 7, Username: "alice"}, 20, authz.ErrForbidden},
		{"superadmin resets any usage", authz.Principal{UserID: 1, Username: "admin", Superadmin: true}, 20, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newUsageResetRepository()
			usecase := ocservuser.New(repo, nil, nil, nil)

			result, err := usecase.ResetUsage(context.Background(), test.principal, test.userID)
			if test.wantError != nil {
				require.ErrorIs(t, err, test.wantError)
				require.Equal(t, 300, repo.users[test.userID].RunningRx)
				require.Equal(t, 500, repo.users[test.userID].RunningTx)
				return
			}

			require.NoError(t, err)
			require.Zero(t, result.RunningRx)
			require.Zero(t, result.RunningTx)
			require.NotNil(t, result.UsageResetAt)
			require.Zero(t, repo.users[test.userID].RunningRx)
			require.Zero(t, repo.users[test.userID].RunningTx)
		})
	}
}

func TestWorkerAccumulatesRunningUsageAfterReset(t *testing.T) {
	repo := newUsageResetRepository()
	usecase := ocservuser.New(repo, nil, nil, nil)
	_, err := usecase.ResetUsage(context.Background(), authz.Principal{UserID: 7}, 10)
	require.NoError(t, err)

	stream := make(chan logger.StreamEntry, 2)
	stream <- logger.StreamEntry{Message: "worker[alice]: 10.0.0.1 sent periodic stats (in: 7, out: 11)", Timestamp: time.Now().UTC()}
	stream <- logger.StreamEntry{Message: "worker[alice]: 10.0.0.1 sent periodic stats (in: 10, out: 20)", Timestamp: time.Now().UTC()}
	close(stream)

	workerRepo := &usageWorkerRepository{state: repo, userID: 10}
	worker := logprocessor.New(context.Background(), stream, workerRepo, usageWorkerAccess{}, usageConnectionObserver{})
	require.ErrorContains(t, worker.CalculateUserStats(), "stream closed")

	updated := repo.users[10]
	require.Equal(t, 10, updated.RunningRx)
	require.Equal(t, 20, updated.RunningTx)
	require.Len(t, workerRepo.traffic, 2)
}

func TestOcservUserResponseUsesRunningUsageFields(t *testing.T) {
	data, err := json.Marshal(models.OcservUser{RunningRx: 7, RunningTx: 11})
	require.NoError(t, err)
	require.Contains(t, string(data), `"running_rx":7`)
	require.Contains(t, string(data), `"running_tx":11`)
	require.NotContains(t, string(data), `"rx":`)
	require.NotContains(t, string(data), `"tx":`)
}

func newUsageResetRepository() *usageResetRepository {
	return &usageResetRepository{users: map[uint]models.OcservUser{
		10: {ID: 10, OwnerID: 7, Username: "alice", TrafficType: models.Free, RunningRx: 300, RunningTx: 500},
		20: {ID: 20, OwnerID: 8, Username: "bob", TrafficType: models.Free, RunningRx: 300, RunningTx: 500},
	}}
}

type usageWorkerRepository struct {
	state   *usageResetRepository
	userID  uint
	traffic []models.OcservUserTrafficStatistics
}

func (r *usageWorkerRepository) FindUser(_ context.Context, username string) (*models.OcservUser, error) {
	user := r.state.users[r.userID]
	return &user, nil
}

func (r *usageWorkerRepository) RecordUsage(_ context.Context, traffic *models.OcservUserTrafficStatistics) (*models.OcservUser, error) {
	user := r.state.users[r.userID]
	user.RunningRx += traffic.Rx
	user.RunningTx += traffic.Tx
	r.state.users[r.userID] = user
	r.traffic = append(r.traffic, *traffic)
	return &user, nil
}

func (r *usageWorkerRepository) CurrentMonthTotals(context.Context, uint, *time.Time) (int, int, error) {
	user := r.state.users[r.userID]
	return user.RunningRx, user.RunningTx, nil
}

func (r *usageWorkerRepository) UpdateAccessState(context.Context, uint, bool, *time.Time) error {
	return nil
}

func (r *usageWorkerRepository) SaveSessionLog(context.Context, *models.OcservUserSessionLog) error {
	return nil
}

type usageWorkerAccess struct{}

func (usageWorkerAccess) Disconnect(string) error { return nil }
func (usageWorkerAccess) Lock(string) error       { return nil }

type usageConnectionObserver struct{}

func (usageConnectionObserver) Observe(context.Context, string, time.Time) error { return nil }
