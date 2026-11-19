package units

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	adminsystem "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/system"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func releaseClient(version string) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"tag_name":"` + version + `"}`)),
			Header:     make(http.Header),
		}, nil
	})}
}

type adminSystemRepository struct {
	updated *models.System
}

func (r *adminSystemRepository) System(context.Context) (*models.System, error) {
	return &models.System{}, nil
}

func (r *adminSystemRepository) SystemUpdate(_ context.Context, settings *models.System) (*models.System, error) {
	r.updated = settings
	return settings, nil
}

func TestAdminSystemUpdateMarksFirstInit(t *testing.T) {
	repository := &adminSystemRepository{}
	usecase := adminsystem.New(repository, nil, nil, nil, nil, adminsystem.Options{})

	siteKey := "site"
	secretKey := "secret"
	autoDelete := false
	keepInactiveDays := 30
	serverAddress := "vpn.example.com"
	serverPort := 443
	connectionName := "VPN"

	response, err := usecase.Update(context.Background(), 1, adminsystem.PatchSystemUpdateData{
		GoogleCaptchaSiteKey:        &siteKey,
		GoogleCaptchaSecretKey:      &secretKey,
		AutoDeleteInactiveUsers:     &autoDelete,
		KeepInactiveUserDays:        &keepInactiveDays,
		ClientProfileServerAddress:  &serverAddress,
		ClientProfileServerPort:     &serverPort,
		ClientProfileConnectionName: &connectionName,
	})

	require.NoError(t, err)
	require.NotNil(t, repository.updated)
	require.True(t, repository.updated.FirstInit)
	require.True(t, response.FirstInit)
}

func TestAdminSystemInitReportsNewerRelease(t *testing.T) {
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, ".release"), []byte("v5.1\n"), 0600))
	usecase := adminsystem.New(&adminSystemRepository{}, nil, nil, nil, nil, adminsystem.Options{
		HTTPClient: releaseClient("v5.2"), InstallDir: directory,
	})

	response, err := usecase.Init(context.Background())

	require.NoError(t, err)
	require.True(t, response.UpdateAvailable)
	require.Equal(t, "v5.2", response.LatestVersion)
}

func TestAdminSystemReleaseUsesNumericVersionOrder(t *testing.T) {
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, ".release"), []byte("v5.10\n"), 0600))
	usecase := adminsystem.New(&adminSystemRepository{}, nil, nil, nil, nil, adminsystem.Options{
		HTTPClient: releaseClient("v5.2"), InstallDir: directory,
	})

	release, err := usecase.Release(context.Background())

	require.NoError(t, err)
	require.False(t, release.Available)
}

func TestAdminSystemUpdateCannotRunConcurrently(t *testing.T) {
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, ".release"), []byte("v5.1\n"), 0600))
	usecase := adminsystem.New(&adminSystemRepository{}, nil, nil, nil, nil, adminsystem.Options{
		HTTPClient: releaseClient("v5.2"), InstallDir: directory,
		UpdateStarter: func(string) error { return nil },
	})

	status, err := usecase.StartUpdate(context.Background())

	require.NoError(t, err)
	require.Equal(t, "running", status.State)
	_, err = usecase.StartUpdate(context.Background())
	require.ErrorIs(t, err, adminsystem.ErrUpdateRunning)
}
