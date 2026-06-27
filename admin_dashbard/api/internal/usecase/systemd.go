package usecase

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
)

type SystemdUsecaseInterface interface {
	Status() (*OcservSystemdStatus, error)
	Restart() error
	Enable() (message string, err error)
	Disable() (message string, err error)
}

type SystemdUsecase struct {
	systemd repository.SystemdRepositoryInterface
}

func NewSystemdUsecase(systemd repository.SystemdRepositoryInterface) *SystemdUsecase {
	return &SystemdUsecase{
		systemd: systemd,
	}
}

func (uc *SystemdUsecase) Status() (*OcservSystemdStatus, error) {
	if os.Getenv("SYSTEMD") != "true" {
		return nil, ErrorSystemdNotRunning
	}

	statusLog, err := uc.systemd.Status(context.Background())
	if err != nil {
		return nil, err
	}

	return ParseSystemctlShow(statusLog), nil
}

func (uc *SystemdUsecase) Restart() error {
	if os.Getenv("SYSTEMD") != "true" {
		return ErrorSystemdNotRunning
	}

	return uc.systemd.Restart(context.Background())
}

func (uc *SystemdUsecase) Enable() (message string, err error) {
	if os.Getenv("SYSTEMD") != "true" {
		return "", ErrorSystemdNotRunning
	}

	statusLog, err := uc.systemd.Status(context.Background())
	if err != nil {
		return "", err
	}

	output := ParseSystemctlShow(statusLog)

	// IMPORTANT CHECK
	if output.UnitFileState == "enabled" {
		return "service already enabled", nil
	}

	err = uc.systemd.Enable(context.Background())
	if err != nil {
		return "", err
	}

	return "service enabling started successfully", nil
}

func (uc *SystemdUsecase) Disable() (message string, err error) {
	if os.Getenv("SYSTEMD") != "true" {
		return "", ErrorSystemdNotRunning
	}

	statusLog, err := uc.systemd.Status(context.Background())
	if err != nil {
		return "", err
	}

	output := ParseSystemctlShow(statusLog)

	// IMPORTANT CHECK
	if output.UnitFileState == "disabled" {
		return "service already disabled", nil
	}

	err = uc.systemd.Disable(context.Background())
	if err != nil {
		return "", err
	}

	return "service disabling started successfully", nil
}

var ErrorSystemdNotRunning = errors.New("systemd is not running")

type OcservSystemdStatus struct {
	ID            string `json:"id"`
	Description   string `json:"description"`
	ActiveState   string `json:"active_state"`
	SubState      string `json:"sub_state"`
	UnitFileState string `json:"unit_file_state"`

	MainPID      int    `json:"main_pid"`
	StartTime    string `json:"start_time"`
	Memory       int64  `json:"memory"`
	CPUUsageNSec int64  `json:"cpu_usage_nsec"`
	Tasks        int    `json:"tasks"`
}

type ActionResponse struct {
	Message string `json:"message" validate:"required"`
}

func ParseSystemctlShow(output string) *OcservSystemdStatus {
	lines := strings.Split(output, "\n")

	data := make(map[string]string)

	for _, line := range lines {
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		data[kv[0]] = kv[1]
	}

	return &OcservSystemdStatus{
		ID:            data["Id"],
		Description:   data["Description"],
		ActiveState:   data["ActiveState"],
		SubState:      data["SubState"],
		UnitFileState: data["UnitFileState"],

		MainPID:      toInt(data["MainPID"]),
		StartTime:    data["ExecMainStartTimestamp"],
		Memory:       toInt64(data["MemoryCurrent"]),
		CPUUsageNSec: toInt64(data["CPUUsageNSec"]),
		Tasks:        toInt(data["TasksCurrent"]),
	}
}

func toInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func toInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
