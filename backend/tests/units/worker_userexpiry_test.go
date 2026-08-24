package units

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	userexpiry "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/worker/userexpiry"
	"github.com/stretchr/testify/require"
)

type expiryRepository struct {
	expired     []models.OcservUser
	monthly     []models.OcservUser
	settings    models.System
	deactivated []uint
	reactivated []uint
	deletedAt   *time.Time
	mu          sync.Mutex
}

func (r *expiryRepository) ExpiredUsers(_ context.Context, at time.Time) ([]models.OcservUser, error) {
	result := make([]models.OcservUser, 0, len(r.expired))
	for _, user := range r.expired {
		if user.ExpireAt != nil && user.DeactivatedAt == nil && !user.ExpireAt.After(at) {
			result = append(result, user)
		}
	}
	return result, nil
}

func (r *expiryRepository) Deactivate(_ context.Context, id uint, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deactivated = append(r.deactivated, id)
	return nil
}

func (r *expiryRepository) MonthlyUsers(context.Context, time.Time) ([]models.OcservUser, error) {
	return r.monthly, nil
}

func (r *expiryRepository) Reactivate(_ context.Context, id uint, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reactivated = append(r.reactivated, id)
	return nil
}

func (r *expiryRepository) SystemSettings(context.Context) (*models.System, error) {
	return &r.settings, nil
}

func (r *expiryRepository) DeleteExpired(_ context.Context, cutoff time.Time) (int64, error) {
	r.deletedAt = &cutoff
	return 2, nil
}

type expiryAccess struct {
	mu           sync.Mutex
	disconnected []string
	locked       []string
	unlocked     []string
}

func (a *expiryAccess) Disconnect(username string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.disconnected = append(a.disconnected, username)
	return nil
}

func (a *expiryAccess) Lock(username string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.locked = append(a.locked, username)
	return nil
}

func (a *expiryAccess) Unlock(username string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.unlocked = append(a.unlocked, username)
	return nil
}

func TestWorkerUserExpiryCoordinatesPersistenceAndAccess(t *testing.T) {
	expiresAt := time.Now().Add(-time.Minute)
	repo := &expiryRepository{expired: []models.OcservUser{{ID: 7, Username: "alice", ExpiryMode: models.ExpiryModeFixed, ExpireAt: &expiresAt}}}
	access := &expiryAccess{}
	usecase := userexpiry.New(repo, access)

	require.NoError(t, usecase.Expire(context.Background(), time.Now()))
	require.Equal(t, []uint{7}, repo.deactivated)
	require.Equal(t, []string{"alice"}, access.disconnected)
	require.Equal(t, []string{"alice"}, access.locked)
}

func TestWorkerDoesNotExpireFirstConnectionUserBeforeConnection(t *testing.T) {
	repo := &expiryRepository{expired: []models.OcservUser{{
		ID: 8, Username: "waiting", ExpiryMode: models.ExpiryModeFirstConnection,
		ExpireDaysAfterFirstConnection: intPointer(30),
	}}}
	access := &expiryAccess{}

	require.NoError(t, userexpiry.New(repo, access).Expire(context.Background(), time.Now().UTC()))
	require.Empty(t, repo.deactivated)
	require.Empty(t, access.locked)
}

func TestWorkerExpiresFirstConnectionUserAtEffectiveExpiry(t *testing.T) {
	firstConnectedAt := time.Date(2026, time.September, 1, 10, 30, 0, 0, time.UTC)
	expiresAt := firstConnectedAt.AddDate(0, 0, 30)
	repo := &expiryRepository{expired: []models.OcservUser{{
		ID: 9, Username: "connected", ExpiryMode: models.ExpiryModeFirstConnection,
		ExpireDaysAfterFirstConnection: intPointer(30), FirstConnectedAt: &firstConnectedAt, ExpireAt: &expiresAt,
	}}}
	access := &expiryAccess{}
	usecase := userexpiry.New(repo, access)

	require.NoError(t, usecase.Expire(context.Background(), expiresAt.Add(-time.Nanosecond)))
	require.Empty(t, repo.deactivated)
	require.NoError(t, usecase.Expire(context.Background(), expiresAt))
	require.Equal(t, []uint{9}, repo.deactivated)
	require.Equal(t, []string{"connected"}, access.locked)
}

func TestWorkerUserExpiryHonorsDeleteSettings(t *testing.T) {
	now := time.Date(2026, time.August, 23, 0, 0, 0, 0, time.UTC)
	repo := &expiryRepository{settings: models.System{AutoDeleteInactiveUsers: true, KeepInactiveUserDays: 30}}
	usecase := userexpiry.New(repo, &expiryAccess{})

	deleted, err := usecase.DeleteInactive(context.Background(), now)
	require.NoError(t, err)
	require.EqualValues(t, 2, deleted)
	require.NotNil(t, repo.deletedAt)
	require.Equal(t, now.AddDate(0, 0, -30), *repo.deletedAt)
}

func intPointer(value int) *int {
	return &value
}
