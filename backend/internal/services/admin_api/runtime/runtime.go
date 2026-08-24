package runtime

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	platformsystemd "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/systemd"
	systemusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/system"
)

const (
	OcservSystemdService  = "ocserv"
	OcservDockerContainer = "ocserv"
)

type SystemdRuntime struct {
	client  platformsystemd.ClientInterface
	enabled bool
}

func NewSystemdRuntime(client platformsystemd.ClientInterface, enabled bool) *SystemdRuntime {
	return &SystemdRuntime{client: client, enabled: enabled}
}

func (r *SystemdRuntime) Status(ctx context.Context) (*systemusecase.Status, error) {
	if !r.enabled {
		return nil, systemusecase.ErrUnavailable
	}
	output, err := r.client.Status(ctx)
	if err != nil {
		return nil, err
	}
	return parseSystemdStatus(output), nil
}

func (r *SystemdRuntime) Restart(ctx context.Context) error {
	if !r.enabled {
		return systemusecase.ErrUnavailable
	}
	return r.client.Restart(ctx)
}

func (r *SystemdRuntime) Enable(ctx context.Context) error {
	if !r.enabled {
		return systemusecase.ErrUnavailable
	}
	return r.client.Enable(ctx)
}

func (r *SystemdRuntime) Disable(ctx context.Context) error {
	if !r.enabled {
		return systemusecase.ErrUnavailable
	}
	return r.client.Disable(ctx)
}

type DockerClient interface {
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerUpdate(ctx context.Context, containerID string, updateConfig container.UpdateConfig) (container.UpdateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
}

type DockerRuntime struct {
	client        DockerClient
	containerName string
}

func NewDockerRuntime(client DockerClient, containerName string) *DockerRuntime {
	return &DockerRuntime{client: client, containerName: containerName}
}

func NewRuntime(dockerMode, systemdEnabled bool) (systemusecase.Runtime, error) {
	if !dockerMode {
		return NewSystemdRuntime(platformsystemd.NewClient(OcservSystemdService), systemdEnabled), nil
	}
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return NewDockerRuntime(dockerClient, OcservDockerContainer), nil
}

func (r *DockerRuntime) Status(ctx context.Context) (*systemusecase.Status, error) {
	inspection, err := r.client.ContainerInspect(ctx, r.containerName)
	if err != nil {
		return nil, err
	}
	if inspection.ContainerJSONBase == nil || inspection.State == nil {
		return nil, fmt.Errorf("docker inspect returned no state for %s", r.containerName)
	}
	active := "inactive"
	if inspection.State.Running {
		active = "active"
	}
	unitState := "disabled"
	if inspection.HostConfig != nil && !inspection.HostConfig.RestartPolicy.IsNone() {
		unitState = "enabled"
	}
	return &systemusecase.Status{
		ID: r.containerName, Description: "Docker container " + r.containerName,
		ActiveState: active, SubState: string(inspection.State.Status), UnitFileState: unitState,
		MainPID: inspection.State.Pid, StartTime: inspection.State.StartedAt,
	}, nil
}

func (r *DockerRuntime) Restart(ctx context.Context) error {
	timeout := 30
	return r.client.ContainerRestart(ctx, r.containerName, container.StopOptions{Timeout: &timeout})
}

func (r *DockerRuntime) Enable(ctx context.Context) error {
	inspection, err := r.client.ContainerInspect(ctx, r.containerName)
	if err != nil {
		return err
	}
	if _, err := r.client.ContainerUpdate(ctx, r.containerName, container.UpdateConfig{
		RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyUnlessStopped},
	}); err != nil {
		return err
	}
	if inspection.State != nil && !inspection.State.Running {
		return r.client.ContainerStart(ctx, r.containerName, container.StartOptions{})
	}
	return nil
}

func (r *DockerRuntime) Disable(ctx context.Context) error {
	inspection, err := r.client.ContainerInspect(ctx, r.containerName)
	if err != nil {
		return err
	}
	if _, err := r.client.ContainerUpdate(ctx, r.containerName, container.UpdateConfig{
		RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyDisabled},
	}); err != nil {
		return err
	}
	if inspection.State != nil && inspection.State.Running {
		timeout := 30
		return r.client.ContainerStop(ctx, r.containerName, container.StopOptions{Timeout: &timeout})
	}
	return nil
}

func parseSystemdStatus(output string) *systemusecase.Status {
	data := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			data[parts[0]] = parts[1]
		}
	}
	return &systemusecase.Status{
		ID: data["Id"], Description: data["Description"], ActiveState: data["ActiveState"],
		SubState: data["SubState"], UnitFileState: data["UnitFileState"],
		MainPID: toInt(data["MainPID"]), StartTime: data["ExecMainStartTimestamp"],
		Memory: toInt64(data["MemoryCurrent"]), CPUUsageNSec: toInt64(data["CPUUsageNSec"]),
		Tasks: toInt(data["TasksCurrent"]),
	}
}

func toInt(value string) int {
	result, _ := strconv.Atoi(value)
	return result
}

func toInt64(value string) int64 {
	result, _ := strconv.ParseInt(value, 10, 64)
	return result
}
