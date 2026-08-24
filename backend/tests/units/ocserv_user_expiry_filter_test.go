package units

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/authz"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	usercontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/ocserv_user"
	userusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/users"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
	"github.com/stretchr/testify/require"
)

type expiryFilterRepository struct {
	userusecase.Repository
	users         []models.OcservUser
	receivedOwner uint
	receivedQuery string
	receivedGroup string
	receivedPage  *request.Pagination
}

func (r *expiryFilterRepository) Users(
	_ context.Context,
	pagination *request.Pagination,
	ownerID uint,
	q string,
	_ string,
	group string,
	window *repository.OcservUserExpiryWindow,
) ([]models.OcservUser, int64, error) {
	r.receivedOwner, r.receivedQuery, r.receivedGroup, r.receivedPage = ownerID, q, group, pagination
	return r.filtered(ownerID, window), int64(len(r.filtered(ownerID, window))), nil
}

func (r *expiryFilterRepository) UsersByUsername(
	ctx context.Context,
	pagination *request.Pagination,
	ownerID uint,
	_ []string,
	q string,
	group string,
	window *repository.OcservUserExpiryWindow,
) ([]models.OcservUser, int64, error) {
	return r.Users(ctx, pagination, ownerID, q, "", group, window)
}

func (r *expiryFilterRepository) filtered(ownerID uint, window *repository.OcservUserExpiryWindow) []models.OcservUser {
	result := make([]models.OcservUser, 0)
	for _, user := range r.users {
		if ownerID != 0 && user.OwnerID != ownerID {
			continue
		}
		if window != nil && (user.ExpireAt == nil || user.ExpireAt.Before(window.StartsAt) || user.ExpireAt.After(window.EndsAt)) {
			continue
		}
		result = append(result, user)
	}
	return result
}

type expiryFilterOCCTL struct{}

func (expiryFilterOCCTL) OnlineSessions() ([]models.OnlineUserSession, error) { return nil, nil }
func (expiryFilterOCCTL) ShowUserByID(string) (models.OnlineUserSession, error) {
	return models.OnlineUserSession{}, nil
}
func (expiryFilterOCCTL) Terminate(string) (string, error)         { return "", nil }
func (expiryFilterOCCTL) Disconnect(string) (string, error)        { return "", nil }
func (expiryFilterOCCTL) DisconnectSession(string) (string, error) { return "", nil }
func (expiryFilterOCCTL) TerminateSession(string) (string, error)  { return "", nil }
func (expiryFilterOCCTL) Reload() (string, error)                  { return "", nil }

func TestOcservUserListFiltersByEffectiveExpiryAndOwner(t *testing.T) {
	now := time.Now().UTC()
	fixedExpiry := now.Add(24 * time.Hour)
	firstConnectedAt := now.Add(-24 * time.Hour)
	firstConnectionExpiry := now.Add(48 * time.Hour)
	outsideExpiry := now.Add(4 * 24 * time.Hour)
	otherOwnerExpiry := now.Add(12 * time.Hour)
	repo := &expiryFilterRepository{users: []models.OcservUser{
		{ID: 1, OwnerID: 7, Username: "fixed", ExpiryMode: models.ExpiryModeFixed, ExpireAt: &fixedExpiry},
		{ID: 2, OwnerID: 7, Username: "first", ExpiryMode: models.ExpiryModeFirstConnection, FirstConnectedAt: &firstConnectedAt, ExpireAt: &firstConnectionExpiry},
		{ID: 3, OwnerID: 7, Username: "outside", ExpiryMode: models.ExpiryModeFixed, ExpireAt: &outsideExpiry},
		{ID: 4, OwnerID: 7, Username: "uncalculated", ExpiryMode: models.ExpiryModeFirstConnection},
		{ID: 5, OwnerID: 8, Username: "other-owner", ExpiryMode: models.ExpiryModeFixed, ExpireAt: &otherOwnerExpiry},
	}}
	usecase := userusecase.New(repo, nil, expiryFilterOCCTL{}, nil)
	days := 3
	pagination := &request.Pagination{Page: 2, PageSize: 25, Order: "id", Sort: "ASC"}

	normalResult, err := usecase.List(context.Background(), userusecase.ListOptions{
		Pagination: pagination,
		Principal:  authz.Principal{UserID: 7},
		Query:      "fi", Group: "defaults", ExpireInDays: &days,
	})
	require.NoError(t, err)
	require.Equal(t, []uint{1, 2}, ocservUserIDs(normalResult.Users))
	require.Equal(t, int64(2), normalResult.Total)
	require.Equal(t, uint(7), repo.receivedOwner)
	require.Equal(t, "fi", repo.receivedQuery)
	require.Equal(t, "defaults", repo.receivedGroup)
	require.Same(t, pagination, repo.receivedPage)

	superadminResult, err := usecase.List(context.Background(), userusecase.ListOptions{
		Pagination:   pagination,
		Principal:    authz.Principal{UserID: 1, Superadmin: true},
		ExpireInDays: &days,
	})
	require.NoError(t, err)
	require.Equal(t, []uint{1, 2, 5}, ocservUserIDs(superadminResult.Users))
	require.Zero(t, repo.receivedOwner)
}

func TestOcservUserListRejectsInvalidExpireInDaysInUsecase(t *testing.T) {
	repo := &expiryFilterRepository{}
	usecase := userusecase.New(repo, nil, expiryFilterOCCTL{}, nil)
	zero := 0

	_, err := usecase.List(context.Background(), userusecase.ListOptions{
		Pagination:   &request.Pagination{Page: 1, PageSize: 50},
		Principal:    authz.Principal{UserID: 1, Superadmin: true},
		ExpireInDays: &zero,
	})
	require.ErrorIs(t, err, userusecase.ErrInvalidExpireInDays)
}

func TestOcservUserListRejectsInvalidExpireInDaysQuery(t *testing.T) {
	config.Init(false, "", 0)
	token, err := crypto.GenerateAccessToken(1, "admin", time.Now().Add(time.Hour).Unix(), true)
	require.NoError(t, err)
	controller := usercontroller.New(userusecase.New(nil, nil, nil, nil))

	for _, value := range []string{"abc", "0", "-1", ""} {
		t.Run("value="+value, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/ocserv/users?expire_in_days="+value, nil)
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
			recorder := httptest.NewRecorder()
			ctx := e.NewContext(req, recorder)
			handler := middlewares.AuthMiddleware()(controller.Users)

			require.NoError(t, handler(ctx))
			require.Equal(t, http.StatusBadRequest, recorder.Code)
		})
	}
}

func ocservUserIDs(users []models.OcservUser) []uint {
	ids := make([]uint, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids
}
