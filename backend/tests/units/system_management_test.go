package units

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	platformsystemd "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/systemd"
	runtimeservice "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/runtime"
	servicecontrolservice "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/service_control"
	servicecontrolusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/service_control"
	systemusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/system"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
	"github.com/stretchr/testify/require"
)

type managementRuntime struct {
	enabled      bool
	restarts     int
	enables      int
	disables     int
	restartError error
}

func (*managementRuntime) AllowedActions() []servicecontrolusecase.Action {
	return []servicecontrolusecase.Action{
		servicecontrolusecase.ActionRestart,
		servicecontrolusecase.ActionEnable,
		servicecontrolusecase.ActionDisable,
	}
}

func (r *managementRuntime) Status(context.Context) (*servicecontrolusecase.Status, error) {
	state := "disabled"
	if r.enabled {
		state = "enabled"
	}
	return &servicecontrolusecase.Status{ID: "ocserv", ActiveState: "active", UnitFileState: state}, nil
}

func (r *managementRuntime) Restart(context.Context) error {
	r.restarts++
	return r.restartError
}

func (r *managementRuntime) Enable(context.Context) error {
	r.enables++
	r.enabled = true
	return nil
}

func (r *managementRuntime) Disable(context.Context) error {
	r.disables++
	r.enabled = false
	return nil
}

type managementConfigStore struct {
	config     systemusecase.OcservConfig
	writes     int
	writeError error
}

func (s *managementConfigStore) Read(context.Context) (*systemusecase.OcservConfig, error) {
	return &s.config, nil
}

func (s *managementConfigStore) Write(_ context.Context, changes systemusecase.OcservConfig) (*systemusecase.OcservConfig, error) {
	s.writes++
	if s.writeError != nil {
		return nil, s.writeError
	}
	s.config = changes
	return &s.config, nil
}

func TestServiceControlEndpointsRequireSuperadminAndAllowActions(t *testing.T) {
	config.Init(false, "", 0)
	runtime := &managementRuntime{}
	controller := servicecontrolservice.New(servicecontrolusecase.New(runtime))
	configController := runtimeservice.New(systemusecase.New(runtime, &managementConfigStore{}, true))

	normalToken := managementToken(t, false)
	for _, test := range []struct {
		method  string
		path    string
		handler echo.HandlerFunc
	}{
		{http.MethodGet, "/systemd/status", controller.Status},
		{http.MethodPost, "/systemd/restart", controller.Restart},
		{http.MethodPost, "/systemd/enable", controller.Enable},
		{http.MethodPost, "/systemd/disable", controller.Disable},
		{http.MethodGet, "/system/ocserv-config", configController.Config},
	} {
		recorder, err := runManagementHandler(test.method, test.path, nil, normalToken, test.handler)
		requireHTTPStatus(t, recorder, err, http.StatusForbidden)
	}
	require.Zero(t, runtime.restarts)
	require.Zero(t, runtime.enables)
	require.Zero(t, runtime.disables)

	superadminToken := managementToken(t, true)
	for _, test := range []struct {
		method  string
		path    string
		handler echo.HandlerFunc
	}{
		{http.MethodGet, "/systemd/status", controller.Status},
		{http.MethodPost, "/systemd/restart", controller.Restart},
		{http.MethodPost, "/systemd/enable", controller.Enable},
		{http.MethodPost, "/systemd/disable", controller.Disable},
	} {
		recorder, err := runManagementHandler(test.method, test.path, nil, superadminToken, test.handler)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, recorder.Code)
	}
	require.Equal(t, 1, runtime.restarts)
	require.Equal(t, 1, runtime.enables)
	require.Equal(t, 1, runtime.disables)
	statusRecorder, err := runManagementHandler(http.MethodGet, "/systemd/status", nil, superadminToken, controller.Status)
	require.NoError(t, err)
	var status servicecontrolusecase.Status
	require.NoError(t, json.Unmarshal(statusRecorder.Body.Bytes(), &status))
	require.Equal(t, []servicecontrolusecase.Action{
		servicecontrolusecase.ActionRestart,
		servicecontrolusecase.ActionEnable,
		servicecontrolusecase.ActionDisable,
	}, status.AllowedActions)

	runtime.restartError = errors.New("restart failed")
	recorder, err := runManagementHandler(http.MethodPost, "/systemd/restart", nil, superadminToken, controller.Restart)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOcservConfigParsingAndAtomicUpdate(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "ocserv.conf")
	original := strings.Join([]string{
		"# package comment",
		"auth = \"certificate\"",
		"tcp-port = 443",
		"udp-port = 443",
		"dns = 1.1.1.1",
		"dns = 8.8.8.8",
		"ipv4-network = 172.16.24.0/24",
		"rekey-method = ssl",
		"banner = \"Old banner\"",
		"custom-directive = keep-this-line",
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(path, []byte(original), 0o640))

	parsed, err := runtimeservice.ParseConfig([]byte(original))
	require.NoError(t, err)
	require.Equal(t, 443, *parsed.TCPPort)
	require.Equal(t, []string{"1.1.1.1", "8.8.8.8"}, *parsed.DNS)
	require.Equal(t, systemusecase.RekeyMethodSSL, *parsed.RekeyMethod)

	runtime := &managementRuntime{}
	usecase := systemusecase.New(runtime, runtimeservice.NewConfigFile(path), true)
	port := 444
	dns := []string{"9.9.9.9"}
	banner := "New banner"
	updated, err := usecase.UpdateConfig(context.Background(), systemusecase.OcservConfig{
		TCPPort: &port, DNS: &dns, Banner: &banner,
	})
	require.NoError(t, err)
	require.Equal(t, 444, *updated.TCPPort)
	require.Equal(t, []string{"9.9.9.9"}, *updated.DNS)
	require.Equal(t, 1, runtime.restarts)
	response, err := usecase.Config(context.Background())
	require.NoError(t, err)
	require.True(t, response.AllowUpdate)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(content), "# package comment")
	require.Contains(t, string(content), "auth = \"certificate\"")
	require.Contains(t, string(content), "custom-directive = keep-this-line")
	require.Contains(t, string(content), "tcp-port = 444")
	require.Equal(t, 1, strings.Count(string(content), "dns = "))
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o640), info.Mode().Perm())
}

func TestOcservConfigRejectsInvalidAndUnsupportedValues(t *testing.T) {
	runtime := &managementRuntime{}
	store := &managementConfigStore{}
	usecase := systemusecase.New(runtime, store, true)
	invalidPort := 70000
	_, err := usecase.UpdateConfig(context.Background(), systemusecase.OcservConfig{TCPPort: &invalidPort})
	require.ErrorIs(t, err, systemusecase.ErrInvalidConfig)
	invalidMethod := systemusecase.RekeyMethod("shell-command")
	_, err = usecase.UpdateConfig(context.Background(), systemusecase.OcservConfig{RekeyMethod: &invalidMethod})
	require.ErrorIs(t, err, systemusecase.ErrInvalidConfig)
	require.Zero(t, store.writes)
	require.Zero(t, runtime.restarts)

	config.Init(false, "", 0)
	controller := runtimeservice.New(usecase)
	body := bytes.NewBufferString(`{"tcp_port":444,"include":"/etc/passwd"}`)
	recorder, handlerErr := runManagementHandler(
		http.MethodPatch, "/system/ocserv-config", body, managementToken(t, true), controller.UpdateConfig,
	)
	require.NoError(t, handlerErr)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Zero(t, store.writes)
}

func TestOcservConfigWriteAndRestartErrors(t *testing.T) {
	port := 444
	writeFailure := errors.New("write failed")
	runtime := &managementRuntime{}
	store := &managementConfigStore{writeError: writeFailure}
	usecase := systemusecase.New(runtime, store, true)

	_, err := usecase.UpdateConfig(context.Background(), systemusecase.OcservConfig{TCPPort: &port})
	require.ErrorIs(t, err, writeFailure)
	require.Zero(t, runtime.restarts)

	restartFailure := errors.New("restart failed")
	runtime.restartError = restartFailure
	store.writeError = nil
	_, err = usecase.UpdateConfig(context.Background(), systemusecase.OcservConfig{TCPPort: &port})
	require.ErrorIs(t, err, restartFailure)
	require.Equal(t, 1, runtime.restarts)
}

func TestDockerModeOcservConfigIsReadOnly(t *testing.T) {
	port := 443
	runtime := &managementRuntime{}
	store := &managementConfigStore{config: systemusecase.OcservConfig{TCPPort: &port}}
	controller := runtimeservice.New(systemusecase.New(runtime, store, false))
	config.Init(false, "", 0)
	token := managementToken(t, true)

	recorder, err := runManagementHandler(http.MethodGet, "/system/ocserv-config", nil, token, controller.Config)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, recorder.Code)
	var response systemusecase.ConfigResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.False(t, response.AllowUpdate)
	require.Equal(t, 443, *response.TCPPort)

	body := bytes.NewBufferString(`{"tcp_port":444}`)
	recorder, err = runManagementHandler(http.MethodPatch, "/system/ocserv-config", body, token, controller.UpdateConfig)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Zero(t, store.writes)
	require.Zero(t, runtime.restarts)
}

type dockerManagementClient struct {
	restarted  string
	restartErr error
	updates    int
	starts     int
	stops      int
}

func (c *dockerManagementClient) ContainerInspect(context.Context, string) (container.InspectResponse, error) {
	return container.InspectResponse{ContainerJSONBase: &container.ContainerJSONBase{
		State:      &container.State{Running: true, Status: container.StateRunning},
		HostConfig: &container.HostConfig{RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyUnlessStopped}},
	}}, nil
}

func (c *dockerManagementClient) ContainerRestart(_ context.Context, containerID string, _ container.StopOptions) error {
	c.restarted = containerID
	return c.restartErr
}

func (c *dockerManagementClient) ContainerUpdate(context.Context, string, container.UpdateConfig) (container.UpdateResponse, error) {
	c.updates++
	return container.UpdateResponse{}, nil
}

func (c *dockerManagementClient) ContainerStart(context.Context, string, container.StartOptions) error {
	c.starts++
	return nil
}

func (c *dockerManagementClient) ContainerStop(context.Context, string, container.StopOptions) error {
	c.stops++
	return nil
}

func TestDockerModeRestartsOcservContainer(t *testing.T) {
	dockerClient := &dockerManagementClient{}
	runtime := servicecontrolservice.NewDockerRuntime(dockerClient, servicecontrolservice.OcservDockerContainer)
	require.NoError(t, runtime.Restart(context.Background()))
	require.Equal(t, "ocserv", dockerClient.restarted)

	dockerClient.restartErr = errors.New("docker restart failed")
	require.ErrorContains(t, runtime.Restart(context.Background()), "docker restart failed")
}

func TestDockerModeReportsNoActionsAndRejectsMutations(t *testing.T) {
	dockerClient := &dockerManagementClient{}
	runtime := servicecontrolservice.NewDockerRuntime(dockerClient, servicecontrolservice.OcservDockerContainer)
	usecase := servicecontrolusecase.New(runtime)

	status, err := usecase.Status(context.Background())
	require.NoError(t, err)
	require.Equal(t, "ocserv", status.ID)
	require.Empty(t, status.AllowedActions)
	require.NotNil(t, status.AllowedActions)

	_, err = usecase.Restart(context.Background())
	require.ErrorIs(t, err, servicecontrolusecase.ErrActionNotSupported)
	_, err = usecase.Enable(context.Background())
	require.ErrorIs(t, err, servicecontrolusecase.ErrActionNotSupported)
	_, err = usecase.Disable(context.Background())
	require.ErrorIs(t, err, servicecontrolusecase.ErrActionNotSupported)
	require.Empty(t, dockerClient.restarted)
	require.Zero(t, dockerClient.updates)
	require.Zero(t, dockerClient.starts)
	require.Zero(t, dockerClient.stops)
}

type systemdManagementClient struct {
	platformsystemd.ClientInterface
	restarts int
	err      error
}

func (*systemdManagementClient) Status(context.Context) (string, error) {
	return "Id=ocserv.service\nActiveState=active\nSubState=running\nUnitFileState=enabled\nMainPID=42\nMemoryCurrent=1024\nCPUUsageNSec=2048\nTasksCurrent=3\n", nil
}

func (c *systemdManagementClient) Restart(context.Context) error {
	c.restarts++
	return c.err
}

func TestSystemdModeUsesExistingRestartClient(t *testing.T) {
	client := &systemdManagementClient{}
	runtime := servicecontrolservice.NewSystemdRuntime(client, true)
	require.Equal(t, []servicecontrolusecase.Action{
		servicecontrolusecase.ActionRestart,
		servicecontrolusecase.ActionEnable,
		servicecontrolusecase.ActionDisable,
	}, runtime.AllowedActions())
	status, err := runtime.Status(context.Background())
	require.NoError(t, err)
	require.Equal(t, "ocserv.service", status.ID)
	require.Equal(t, 42, status.MainPID)
	require.Equal(t, int64(1024), status.Memory)
	require.NoError(t, runtime.Restart(context.Background()))
	require.Equal(t, 1, client.restarts)

	client.err = errors.New("systemctl restart failed")
	require.ErrorContains(t, runtime.Restart(context.Background()), "systemctl restart failed")
}

func managementToken(t *testing.T, superadmin bool) string {
	t.Helper()
	if superadmin {
		return "superadmin"
	}
	return "normal"
}

func runManagementHandler(method, path string, body *bytes.Buffer, token string, handler echo.HandlerFunc) (*httptest.ResponseRecorder, error) {
	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		requestBody = bytes.NewReader(body.Bytes())
	}
	e := echo.New()
	req := httptest.NewRequest(method, path, requestBody)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	ctx := e.NewContext(req, recorder)
	wrapped := middlewares.AuthMiddleware(staticAuthenticator{})(middlewares.SuperadminPermission()(handler))
	return recorder, wrapped(ctx)
}

func requireHTTPStatus(t *testing.T, recorder *httptest.ResponseRecorder, err error, expected int) {
	t.Helper()
	if httpError, ok := err.(*echo.HTTPError); ok {
		require.Equal(t, expected, httpError.Code)
		return
	}
	require.NoError(t, err)
	require.Equal(t, expected, recorder.Code)
}
