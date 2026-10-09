package servicecontrol

import (
	"context"
	"errors"
)

var (
	ErrUnavailable        = errors.New("system service is unavailable")
	ErrActionNotSupported = errors.New("service control action is not supported in this deployment mode")
)

type Action string

const (
	ActionRestart Action = "restart"
	ActionEnable  Action = "enable"
	ActionDisable Action = "disable"
)

// Status is a deployment-neutral view of the managed Ocserv runtime.
type Status struct {
	ID             string   `json:"id"`
	Description    string   `json:"description"`
	ActiveState    string   `json:"active_state"`
	SubState       string   `json:"sub_state"`
	UnitFileState  string   `json:"unit_file_state"`
	MainPID        int      `json:"main_pid"`
	StartTime      string   `json:"start_time"`
	Memory         int64    `json:"memory"`
	CPUUsageNSec   int64    `json:"cpu_usage_nsec"`
	Tasks          int      `json:"tasks"`
	AllowedActions []Action `json:"allowed_actions" enums:"restart,enable,disable"`
}

type ActionResult struct {
	Message string `json:"message" validate:"required"`
}

type Runtime interface {
	AllowedActions() []Action
	Status(ctx context.Context) (*Status, error)
	Restart(ctx context.Context) error
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
}
