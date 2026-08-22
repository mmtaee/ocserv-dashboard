package repository

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// SystemdRepository provides methods to interact with systemd services
// using systemctl commands.
type SystemdRepository struct {
	service string
}

// SystemdRepositoryInterface defines available systemd actions
// for a given service.
type SystemdRepositoryInterface interface {
	Status(ctx context.Context) (string, error)
	Restart(ctx context.Context) error
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
}

// NewSystemdRepository creates a new SystemdRepository instance
// for a given systemd service name (e.g., "ocserv").
func NewSystemdRepository(service string) *SystemdRepository {
	return &SystemdRepository{
		service: service,
	}
}

// runCommand executes a systemctl command with sudo and returns
// stdout output or an error containing stderr details.
func (s *SystemdRepository) runCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "sudo", append([]string{"systemctl"}, args...)...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("systemctl %v error: %v - %s", args, err, stderr.String())
	}

	return stdout.String(), nil
}

// Status retrieves detailed information about the service using
// `systemctl show`. It returns raw key=value output including:
// - ActiveState (e.g., active, inactive)
// - SubState (e.g., running)
// - MainPID
// - Memory usage
// - CPU usage
// - Task count
// This is useful for monitoring and diagnostics.
func (s *SystemdRepository) Status(ctx context.Context) (string, error) {
	return s.runCommand(
		ctx,
		"show", s.service,
		"-p", "Id",
		"-p", "Description",
		"-p", "ActiveState",
		"-p", "SubState",
		"-p", "UnitFileState",
		"-p", "MainPID",
		"-p", "ExecMainStartTimestamp",
		"-p", "MemoryCurrent",
		"-p", "CPUUsageNSec",
		"-p", "TasksCurrent",
		"--no-page",
	)
}

// Restart restarts the systemd service.
// This stops and then starts the service again. It is useful when:
// - Applying configuration changes
// - Recovering from transient errors
// Command executed:
//
//	sudo systemctl restart <service>
func (s *SystemdRepository) Restart(ctx context.Context) error {
	_, err := s.runCommand(ctx, "restart", s.service)
	return err
}

// Enable enables the service to start automatically at system boot.
// This does not start the service immediately; it only ensures
// it will be started on next boot.
// Command executed:
//
//	sudo systemctl enable <service>
func (s *SystemdRepository) Enable(ctx context.Context) error {
	_, err := s.runCommand(ctx, "enable", s.service)
	if err != nil {
		return err
	}
	_, err = s.runCommand(ctx, "start", s.service)

	return err
}

// Disable disables the service from starting automatically at boot.
// This does not stop the service if it is currently running.
// Command executed:
//
//	sudo systemctl disable <service>
func (s *SystemdRepository) Disable(ctx context.Context) error {
	_, err := s.runCommand(ctx, "disable", s.service)
	if err != nil {
		return err
	}
	_, err = s.runCommand(ctx, "stop", s.service)
	return err
}
