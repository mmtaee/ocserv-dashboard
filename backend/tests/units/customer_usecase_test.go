package units

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	customerapi "github.com/mmtaee/ocserv-dashboard/backend/internal/services/customer_api"
	customerusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/customer_api"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
	"github.com/stretchr/testify/require"
)

type customerSystemRepository struct{}

func (customerSystemRepository) System(context.Context) (*models.System, error) {
	return &models.System{ClientProfileConnectionName: "VPN", ClientProfileServerAddress: "vpn.example.com", ClientProfileServerPort: 443}, nil
}

type customerUserRepository struct {
	users map[string]*models.OcservUser
	logs  []models.OcservUserSessionLog
}

func (r *customerUserRepository) GetByUsername(_ context.Context, username string) (*models.OcservUser, error) {
	item, ok := r.users[username]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *item
	return &copy, nil
}

func (r *customerUserRepository) Update(_ context.Context, value *models.OcservUser) (*models.OcservUser, error) {
	copy := *value
	r.users[value.Username] = &copy
	return value, nil
}

func (*customerUserRepository) UserStatistics(context.Context, uint, *time.Time, *time.Time) ([]models.DailyTraffic, error) {
	return []models.DailyTraffic{{Date: "2026-09-11", Rx: 1, Tx: 2}}, nil
}

func (*customerUserRepository) TotalBandwidthUserDateRange(context.Context, uint, *time.Time, *time.Time) (repository.TotalBandwidths, error) {
	return repository.TotalBandwidths{RX: 10, TX: 20}, nil
}

func (r *customerUserRepository) UserSessionLogs(_ context.Context, pagination *request.Pagination, username string, _, _ *time.Time) (*[]models.OcservUserSessionLog, int64, error) {
	filtered := make([]models.OcservUserSessionLog, 0)
	for _, log := range r.logs {
		if log.Username == username {
			filtered = append(filtered, log)
		}
	}
	total := int64(len(filtered))
	start := min((pagination.Page-1)*pagination.PageSize, len(filtered))
	end := min(start+pagination.PageSize, len(filtered))
	filtered = filtered[start:end]
	return &filtered, total, nil
}

type customerRuntime struct {
	sessions   []models.OnlineUserSession
	disconnect string
	terminate  string
	password   string
}

func (r *customerRuntime) OnlineSessions() ([]models.OnlineUserSession, error) {
	return r.sessions, nil
}
func (r *customerRuntime) Disconnect(username string) (string, error) {
	r.disconnect = username
	return "", nil
}
func (r *customerRuntime) Terminate(username string) (string, error) {
	r.terminate = username
	return "", nil
}
func (r *customerRuntime) Create(_, _, password string, _ *models.OcservUserConfig) error {
	r.password = password
	return nil
}
func (*customerRuntime) CertificatePath(username string) (string, error) {
	return "/tmp/" + username + ".p12", nil
}
func (*customerRuntime) CreateCertificate(string, string) error { return nil }
func (*customerRuntime) CertificateStatus(string) user.CertificateStatus {
	return user.CertificateStatus{Enabled: true, Available: true}
}

func newCustomerUsecase() (*customerusecase.Usecase, *customerUserRepository, *customerRuntime) {
	users := &customerUserRepository{users: map[string]*models.OcservUser{
		"alice": {ID: 1, Username: "alice", Password: "correct", OwnerID: 7, Owner: models.User{ID: 7, Username: "admin"}},
		"bob":   {ID: 2, Username: "bob", Password: "other", OwnerID: 7},
	}}
	runtime := &customerRuntime{sessions: []models.OnlineUserSession{{ID: 1, Username: "alice"}, {ID: 2, Username: "bob"}}}
	return customerusecase.New(customerSystemRepository{}, users, runtime, "secret", runtime), users, runtime
}

func TestCustomerLoginIssuesSevenHourJWT(t *testing.T) {
	usecase, _, _ := newCustomerUsecase()
	before := time.Now()
	result, err := usecase.Login(context.Background(), customerusecase.Credentials{Username: "alice", Password: "correct"})
	require.NoError(t, err)
	require.WithinDuration(t, before.Add(7*time.Hour), result.ExpiresAt, time.Second)
	require.Len(t, strings.Split(result.Token, "."), 3)
	principal, err := usecase.AuthenticateToken(context.Background(), result.Token)
	require.NoError(t, err)
	require.Equal(t, uint(1), principal.UserID)
	require.Equal(t, "alice", principal.Username)
}

func TestCustomerLoginRejectsInvalidPassword(t *testing.T) {
	usecase, _, _ := newCustomerUsecase()
	_, err := usecase.Login(context.Background(), customerusecase.Credentials{Username: "alice", Password: "wrong"})
	require.ErrorIs(t, err, customerusecase.ErrInvalidCredentials)
}

func TestCustomerTokenRejectsExpiredAndInvalidTokens(t *testing.T) {
	usecase, _, _ := newCustomerUsecase()
	expired := signedCustomerToken(t, "secret", map[string]any{"uid": 1, "sub": "alice", "iat": 1, "exp": 2})
	_, err := usecase.AuthenticateToken(context.Background(), expired)
	require.ErrorIs(t, err, customerusecase.ErrInvalidToken)
	_, err = usecase.AuthenticateToken(context.Background(), "not-a-jwt")
	require.ErrorIs(t, err, customerusecase.ErrInvalidToken)
	mismatched := signedCustomerToken(t, "secret", map[string]any{"uid": 2, "sub": "alice", "iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix()})
	_, err = usecase.AuthenticateToken(context.Background(), mismatched)
	require.ErrorIs(t, err, customerusecase.ErrInvalidToken)
}

func TestCustomerActionsUseOnlyAuthenticatedIdentity(t *testing.T) {
	usecase, users, runtime := newCustomerUsecase()
	result, err := usecase.Summary(context.Background(), "alice")
	require.NoError(t, err)
	require.Equal(t, "alice", result.OcservUser.Username)
	require.Equal(t, "admin", result.OcservUser.Owner)

	sessions, err := usecase.Sessions(context.Background(), "alice")
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	require.Equal(t, "alice", sessions[0].Username)

	require.NoError(t, usecase.Disconnect(context.Background(), "alice"))
	require.NoError(t, usecase.Terminate(context.Background(), "alice"))
	require.NoError(t, usecase.ChangePassword(context.Background(), "alice", "new-password"))
	require.Equal(t, "alice", runtime.disconnect)
	require.Equal(t, "alice", runtime.terminate)
	require.Equal(t, "new-password", runtime.password)
	require.Equal(t, "new-password", users.users["alice"].Password)

	path, username, err := usecase.CertificatePath(context.Background(), "alice")
	require.NoError(t, err)
	require.Equal(t, "alice", username)
	require.Equal(t, "/tmp/alice.p12", path)
	setup, err := usecase.CiscoSetup(context.Background(), "alice", "https://panel.example.com")
	require.NoError(t, err)
	require.Contains(t, setup.CertificateImportURI, "customers%2Fsetup%2Fcisco%2Fcertificate")
}

func TestCustomerStatsActivitiesAndBandwidthAreOwnerScoped(t *testing.T) {
	usecase, users, _ := newCustomerUsecase()
	users.logs = []models.OcservUserSessionLog{{Username: "alice"}, {Username: "bob"}}
	stats, err := usecase.Statistics(context.Background(), "alice", customerusecase.DateRange{})
	require.NoError(t, err)
	require.Len(t, stats, 1)
	activities, err := usecase.Activities(context.Background(), "alice", &request.Pagination{Page: 1, PageSize: 1}, customerusecase.DateRange{})
	require.NoError(t, err)
	require.Equal(t, int64(1), activities.Total)
	require.Equal(t, "alice", (*activities.Logs)[0].Username)
	bandwidth, err := usecase.Bandwidth(context.Background(), "alice", customerusecase.DateRange{})
	require.NoError(t, err)
	require.Equal(t, float64(10), bandwidth.RX)
}

func TestCustomerAuthMiddlewareProtectsRoutes(t *testing.T) {
	usecase, _, _ := newCustomerUsecase()
	e := echo.New()
	e.GET("/protected", func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) }, customerapi.AuthMiddleware(usecase))

	response := httptest.NewRecorder()
	e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/protected", nil))
	require.Equal(t, http.StatusUnauthorized, response.Code)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	response = httptest.NewRecorder()
	e.ServeHTTP(response, req)
	require.Equal(t, http.StatusUnauthorized, response.Code)

	login, err := usecase.Login(context.Background(), customerusecase.Credentials{Username: "alice", Password: "correct"})
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+login.Token)
	response = httptest.NewRecorder()
	e.ServeHTTP(response, req)
	require.Equal(t, http.StatusNoContent, response.Code)
}

func signedCustomerToken(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadJSON, err := json.Marshal(claims)
	require.NoError(t, err)
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	mac := hmac.New(sha256.New, []byte(secret))
	_, err = mac.Write([]byte(header + "." + payload))
	require.NoError(t, err)
	return header + "." + payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
