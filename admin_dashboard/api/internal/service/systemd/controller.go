package systemd

import (
	"errors"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
)

type SystemdController struct {
	systemdUseCase usecase.SystemdUseCase
	req            *request.Request
	validator      *request.Validator
}

func NewSystemdController(systemdUseCase usecase.SystemdUseCase) *SystemdController {
	return &SystemdController{
		systemdUseCase: systemdUseCase,
		req:            &request.Request{},
		validator:      request.NewValidator(),
	}
}

func (ctrl *SystemdController) Status(c echo.Context) error {
	if os.Getenv("SYSTEMD") != "true" {
		return ctrl.req.BadRequest(c, errors.New("systemd is not running"))
	}

	statusLog, err := ctrl.systemdUseCase.Status(c.Request().Context())
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	output := ParseSystemctlShow(statusLog)
	return c.JSON(http.StatusOK, output)
}

func (ctrl *SystemdController) Restart(c echo.Context) error {
	if os.Getenv("SYSTEMD") != "true" {
		return ctrl.req.BadRequest(c, errors.New("systemd is not running"))
	}

	err := ctrl.systemdUseCase.Restart(c.Request().Context())
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, ActionResponse{
		Message: "service restarting started successfully",
	})
}

func (ctrl *SystemdController) Enable(c echo.Context) error {
	if os.Getenv("SYSTEMD") != "true" {
		return ctrl.req.BadRequest(c, errors.New("systemd is not running"))
	}

	statusLog, err := ctrl.systemdUseCase.Status(c.Request().Context())
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	output := ParseSystemctlShow(statusLog)

	if output.UnitFileState == "enabled" {
		return c.JSON(http.StatusOK, ActionResponse{
			Message: "service already enabled",
		})
	}

	err = ctrl.systemdUseCase.Enable(c.Request().Context())
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, ActionResponse{
		Message: "service enabling started successfully",
	})
}

func (ctrl *SystemdController) Disable(c echo.Context) error {
	if os.Getenv("SYSTEMD") != "true" {
		return ctrl.req.BadRequest(c, errors.New("systemd is not running"))
	}

	statusLog, err := ctrl.systemdUseCase.Status(c.Request().Context())
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	output := ParseSystemctlShow(statusLog)

	if output.UnitFileState == "disabled" {
		return c.JSON(http.StatusOK, ActionResponse{
			Message: "service already disabled",
		})
	}

	err = ctrl.systemdUseCase.Disable(c.Request().Context())
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, ActionResponse{
		Message: "service disabling started successfully",
	})
}

func (ctrl *SystemdController) GetMainConfig(c echo.Context) error {
	if os.Getenv("SYSTEMD") != "true" {
		return ctrl.req.BadRequest(c, errors.New("systemd is not running"))
	}

	config, err := ctrl.systemdUseCase.GetMainConfig(c.Request().Context())
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, config)
}

func (ctrl *SystemdController) UpdateMainConfig(c echo.Context) error {
	if os.Getenv("SYSTEMD") != "true" {
		return ctrl.req.BadRequest(c, errors.New("systemd is not running"))
	}

	var config models.OcservMainConfig
	if err := c.Bind(&config); err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	if err := ctrl.validator.Validate(config); err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	if err := ctrl.systemdUseCase.UpdateMainConfig(c.Request().Context(), &config); err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, ActionResponse{
		Message: "config updated and service restarted successfully",
	})
}
