package systemd

import (
	"context"
	"strconv"
	"strings"
)

type Usecase struct {
	repository Repository
	enabled    bool
}

func New(repository Repository, enabled bool) *Usecase {
	return &Usecase{repository: repository, enabled: enabled}
}

func (u *Usecase) Status(ctx context.Context) (*Status, error) {
	if !u.enabled {
		return nil, ErrUnavailable
	}
	output, err := u.repository.Status(ctx)
	if err != nil {
		return nil, err
	}
	status := parseStatus(output)
	return &status, nil
}

func (u *Usecase) Restart(ctx context.Context) (*ActionResult, error) {
	if !u.enabled {
		return nil, ErrUnavailable
	}
	if err := u.repository.Restart(ctx); err != nil {
		return nil, err
	}
	return &ActionResult{Message: "service restarting started successfully"}, nil
}

func (u *Usecase) Enable(ctx context.Context) (*ActionResult, error) {
	status, err := u.Status(ctx)
	if err != nil {
		return nil, err
	}
	if status.UnitFileState == "enabled" {
		return &ActionResult{Message: "service already enabled"}, nil
	}
	if err := u.repository.Enable(ctx); err != nil {
		return nil, err
	}
	return &ActionResult{Message: "service enabling started successfully"}, nil
}

func (u *Usecase) Disable(ctx context.Context) (*ActionResult, error) {
	status, err := u.Status(ctx)
	if err != nil {
		return nil, err
	}
	if status.UnitFileState == "disabled" {
		return &ActionResult{Message: "service already disabled"}, nil
	}
	if err := u.repository.Disable(ctx); err != nil {
		return nil, err
	}
	return &ActionResult{Message: "service disabling started successfully"}, nil
}

func parseStatus(output string) Status {
	data := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			data[parts[0]] = parts[1]
		}
	}
	return Status{
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
