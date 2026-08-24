package units

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/authz"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	ocservuser "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/users"
	"github.com/stretchr/testify/require"
)

type bulkUserRepository struct {
	users        map[uint]models.OcservUser
	failUpdateID uint
	commits      int
	rollbacks    int
}

func (r *bulkUserRepository) WithTransaction(ctx context.Context, operation func(repository.OcservUserBulkTx) error) error {
	working := cloneBulkUsers(r.users)
	tx := &bulkUserTx{users: working, failUpdateID: r.failUpdateID}
	if err := operation(tx); err != nil {
		r.rollbacks++
		return err
	}
	r.users = working
	r.commits++
	return nil
}

type bulkUserTx struct {
	users        map[uint]models.OcservUser
	failUpdateID uint
}

func (tx *bulkUserTx) GetByIDsForUpdate(_ context.Context, ids []uint) ([]models.OcservUser, error) {
	users := make([]models.OcservUser, 0, len(ids))
	for _, id := range ids {
		if user, ok := tx.users[id]; ok {
			users = append(users, user)
		}
	}
	sort.Slice(users, func(i, j int) bool { return users[i].ID < users[j].ID })
	return users, nil
}

func (tx *bulkUserTx) Update(_ context.Context, user *models.OcservUser) (*models.OcservUser, error) {
	if user.ID == tx.failUpdateID {
		return nil, errors.New("forced update failure")
	}
	tx.users[user.ID] = *user
	return user, nil
}

func (tx *bulkUserTx) Delete(_ context.Context, id uint) (*models.OcservUser, error) {
	user, ok := tx.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	delete(tx.users, id)
	return &user, nil
}

func (tx *bulkUserTx) Lock(_ context.Context, id uint) error {
	user := tx.users[id]
	user.IsLocked = true
	tx.users[id] = user
	return nil
}

func (tx *bulkUserTx) UnLock(_ context.Context, id uint) error {
	user := tx.users[id]
	user.IsLocked = false
	tx.users[id] = user
	return nil
}

type bulkAccountState struct {
	group    string
	password string
	locked   bool
}

type bulkAccountStore struct {
	accounts map[string]bulkAccountState
}

func (s *bulkAccountStore) Create(group, username, password string, _ *models.OcservUserConfig) error {
	state := s.accounts[username]
	state.group = group
	state.password = password
	s.accounts[username] = state
	return nil
}

func (s *bulkAccountStore) Lock(username string) (string, error) {
	state := s.accounts[username]
	state.locked = true
	s.accounts[username] = state
	return "", nil
}

func (s *bulkAccountStore) UnLock(username string) (string, error) {
	state := s.accounts[username]
	state.locked = false
	s.accounts[username] = state
	return "", nil
}

func (s *bulkAccountStore) Delete(username string) (string, error) {
	delete(s.accounts, username)
	return "", nil
}

func (s *bulkAccountStore) CertificateBackup(string) (*models.OcservUserCertificateBackup, error) {
	return nil, nil
}

func (s *bulkAccountStore) RestoreCertificateBackup(string, *models.OcservUserCertificateBackup) error {
	return nil
}

type bulkRuntime struct{}

func (bulkRuntime) Reload() (string, error)           { return "", nil }
func (bulkRuntime) Disconnect(string) (string, error) { return "", nil }
func (bulkRuntime) Terminate(string) (string, error)  { return "", nil }

func TestBulkSuperadminCanManageAnyUser(t *testing.T) {
	repo, accounts, usecase := newBulkFixture()
	principal := authz.Principal{UserID: 1, Username: "admin", Superadmin: true}
	description := "updated"

	result, err := usecase.Update(context.Background(), principal, ocservuser.BulkUpdateRequest{Users: []ocservuser.BulkUpdateItem{
		{ID: 10, Changes: ocservuser.UpdateOcservUserData{Description: &description}},
		{ID: 20, Changes: ocservuser.UpdateOcservUserData{Description: &description}},
	}})
	require.NoError(t, err)
	require.Equal(t, 2, result.Count)
	require.Equal(t, description, repo.users[20].Description)

	enabled := false
	_, err = usecase.SetEnabled(context.Background(), principal, ocservuser.BulkStatusRequest{IDs: []uint{20}, Enabled: &enabled})
	require.NoError(t, err)
	require.True(t, repo.users[20].IsLocked)
	require.True(t, accounts.accounts["bob"].locked)

	deleted, err := usecase.Delete(context.Background(), principal, ocservuser.BulkIDsRequest{IDs: []uint{20}})
	require.NoError(t, err)
	require.Equal(t, 1, deleted.Count)
	_, exists := repo.users[20]
	require.False(t, exists)
}

func TestBulkNormalUserCanAssignAndRemoveGroupFromOwnedUsers(t *testing.T) {
	repo, accounts, usecase := newBulkFixture()
	principal := authz.Principal{UserID: 7, Username: "staff"}

	result, err := usecase.SetGroup(context.Background(), principal, ocservuser.BulkGroupRequest{IDs: []uint{10}, Group: "premium"})
	require.NoError(t, err)
	require.Equal(t, "premium", result.Users[0].Group)
	require.Equal(t, "premium", accounts.accounts["alice"].group)

	result, err = usecase.SetGroup(context.Background(), principal, ocservuser.BulkGroupRequest{IDs: []uint{10}})
	require.NoError(t, err)
	require.Equal(t, "defaults", result.Users[0].Group)
	require.Equal(t, "defaults", repo.users[10].Group)
}

func TestBulkNormalUserCannotManageAnotherOwnersUser(t *testing.T) {
	repo, accounts, usecase := newBulkFixture()
	principal := authz.Principal{UserID: 7, Username: "staff"}

	_, err := usecase.SetGroup(context.Background(), principal, ocservuser.BulkGroupRequest{IDs: []uint{10, 20}, Group: "premium"})
	require.ErrorIs(t, err, authz.ErrForbidden)
	require.Equal(t, 1, repo.rollbacks)
	require.Equal(t, "defaults", repo.users[10].Group)
	require.Equal(t, "defaults", accounts.accounts["alice"].group)
}

func TestBulkRejectsInvalidIDsBeforeMutation(t *testing.T) {
	repo, accounts, usecase := newBulkFixture()
	principal := authz.Principal{UserID: 1, Username: "admin", Superadmin: true}

	_, err := usecase.SetGroup(context.Background(), principal, ocservuser.BulkGroupRequest{IDs: []uint{10, 99}, Group: "premium"})
	require.ErrorContains(t, err, "not found")
	require.Equal(t, 1, repo.rollbacks)
	require.Equal(t, "defaults", repo.users[10].Group)
	require.Equal(t, "defaults", accounts.accounts["alice"].group)

	_, err = usecase.Delete(context.Background(), principal, ocservuser.BulkIDsRequest{IDs: []uint{10, 10}})
	require.ErrorContains(t, err, "duplicate user id")
	require.Equal(t, 1, repo.rollbacks, "duplicate validation must happen before opening a transaction")
}

func TestBulkTransactionRollsBackEarlierUpdates(t *testing.T) {
	repo, accounts, usecase := newBulkFixture()
	repo.failUpdateID = 20
	principal := authz.Principal{UserID: 1, Username: "admin", Superadmin: true}
	group := "premium"

	_, err := usecase.Update(context.Background(), principal, ocservuser.BulkUpdateRequest{Users: []ocservuser.BulkUpdateItem{
		{ID: 10, Changes: ocservuser.UpdateOcservUserData{Group: &group}},
		{ID: 20, Changes: ocservuser.UpdateOcservUserData{Group: &group}},
	}})
	require.ErrorContains(t, err, "forced update failure")
	require.Equal(t, 1, repo.rollbacks)
	require.Equal(t, "defaults", repo.users[10].Group)
	require.Equal(t, "defaults", accounts.accounts["alice"].group)
}

func newBulkFixture() (*bulkUserRepository, *bulkAccountStore, *ocservuser.BulkUsecase) {
	repo := &bulkUserRepository{users: map[uint]models.OcservUser{
		10: {ID: 10, OwnerID: 7, Username: "alice", Password: "alice-pass", Group: "defaults"},
		20: {ID: 20, OwnerID: 8, Username: "bob", Password: "bob-pass", Group: "defaults"},
	}}
	accounts := &bulkAccountStore{accounts: map[string]bulkAccountState{
		"alice": {group: "defaults", password: "alice-pass"},
		"bob":   {group: "defaults", password: "bob-pass"},
	}}
	return repo, accounts, ocservuser.NewBulk(repo, accounts, bulkRuntime{})
}

func cloneBulkUsers(source map[uint]models.OcservUser) map[uint]models.OcservUser {
	result := make(map[uint]models.OcservUser, len(source))
	for id, user := range source {
		result[id] = user
	}
	return result
}
