package repository

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/mainconfig"
)

type SystemdRepository interface {
	Status(ctx context.Context) (string, error)
	Restart(ctx context.Context) error
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
	GetMainConfig(ctx context.Context) (*models.OcservMainConfig, error)
	UpdateMainConfig(ctx context.Context, config *models.OcservMainConfig) error
}

type systemdRepository struct {
	service     string
	mainConfig  mainconfig.MainConfigInterface
}

func NewSystemdRepository(service string) SystemdRepository {
	return &systemdRepository{
		service:    service,
		mainConfig: mainconfig.NewMainConfigRepository(),
	}
}

func (s *systemdRepository) runCommand(ctx context.Context, args ...string) (string, error) {
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

func (s *systemdRepository) Status(ctx context.Context) (string, error) {
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

func (s *systemdRepository) Restart(ctx context.Context) error {
	_, err := s.runCommand(ctx, "restart", s.service)
	return err
}

func (s *systemdRepository) Enable(ctx context.Context) error {
	_, err := s.runCommand(ctx, "enable", s.service)
	if err != nil {
		return err
	}
	_, err = s.runCommand(ctx, "start", s.service)
	return err
}

func (s *systemdRepository) Disable(ctx context.Context) error {
	_, err := s.runCommand(ctx, "disable", s.service)
	if err != nil {
		return err
	}
	_, err = s.runCommand(ctx, "stop", s.service)
	return err
}

func (s *systemdRepository) GetMainConfig(ctx context.Context) (*models.OcservMainConfig, error) {
	return s.mainConfig.Read(ctx)
}

func (s *systemdRepository) UpdateMainConfig(ctx context.Context, config *models.OcservMainConfig) error {
	if err := s.mainConfig.Write(ctx, config); err != nil {
		return err
	}
	// Restart service after successful update
	return s.Restart(ctx)
}
