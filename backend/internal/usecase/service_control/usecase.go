package servicecontrol

import (
	"context"
	"fmt"
)

type Usecase struct {
	runtime Runtime
}

func New(runtime Runtime) *Usecase {
	return &Usecase{runtime: runtime}
}

func (u *Usecase) Status(ctx context.Context) (*Status, error) {
	status, err := u.runtime.Status(ctx)
	if err != nil {
		return nil, err
	}
	status.AllowedActions = u.runtime.AllowedActions()
	return status, nil
}

func (u *Usecase) Restart(ctx context.Context) (*ActionResult, error) {
	if err := u.ensureAllowed(ActionRestart); err != nil {
		return nil, err
	}
	if err := u.runtime.Restart(ctx); err != nil {
		return nil, err
	}
	return &ActionResult{Message: "service restarting started successfully"}, nil
}

func (u *Usecase) Enable(ctx context.Context) (*ActionResult, error) {
	if err := u.ensureAllowed(ActionEnable); err != nil {
		return nil, err
	}
	status, err := u.runtime.Status(ctx)
	if err != nil {
		return nil, err
	}
	if status.UnitFileState == "enabled" {
		return &ActionResult{Message: "service already enabled"}, nil
	}
	if err := u.runtime.Enable(ctx); err != nil {
		return nil, err
	}
	return &ActionResult{Message: "service enabling started successfully"}, nil
}

func (u *Usecase) Disable(ctx context.Context) (*ActionResult, error) {
	if err := u.ensureAllowed(ActionDisable); err != nil {
		return nil, err
	}
	status, err := u.runtime.Status(ctx)
	if err != nil {
		return nil, err
	}
	if status.UnitFileState == "disabled" {
		return &ActionResult{Message: "service already disabled"}, nil
	}
	if err := u.runtime.Disable(ctx); err != nil {
		return nil, err
	}
	return &ActionResult{Message: "service disabling started successfully"}, nil
}

func (u *Usecase) ensureAllowed(action Action) error {
	for _, allowed := range u.runtime.AllowedActions() {
		if action == allowed {
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrActionNotSupported, action)
}
