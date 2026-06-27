package systemd

import (
	"errors"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	systemd repository.SystemdRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request: request.NewCustomRequest(),
		systemd: repository.NewSystemdRepository("ocserv"),
	}
}

// Status
// @Summary      Ocserv systemctl status
// @Description  Ocserv systemctl status
// @Tags         Systemd
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      429 {object} middlewares.TooManyRequests
// @Success      200 {object}  OcservSystemdStatus
// @Router       /systemd/status [get]
func (ctl *Controller) Status(c echo.Context) error {
	if os.Getenv("SYSTEMD") != "true" {
		return ctl.request.BadRequest(c, errors.New("systemd is not running"))
	}

	statusLog, err := ctl.systemd.Status(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	output := ParseSystemctlShow(statusLog)
	return c.JSON(http.StatusOK, output)
}

// Restart
// @Summary      Restart ocserv service
// @Description  Restart ocserv systemd service
// @Tags         Systemd
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      429 {object} middlewares.TooManyRequests
// @Success      200 {object}  ActionResponse
// @Router       /systemd/restart [post]
func (ctl *Controller) Restart(c echo.Context) error {
	if os.Getenv("SYSTEMD") != "true" {
		return ctl.request.BadRequest(c, errors.New("systemd is not running"))
	}

	err := ctl.systemd.Restart(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, ActionResponse{
		Message: "service restarting started successfully",
	})
}

// Enable
// @Summary      Enable ocserv service
// @Description  Enable ocserv systemd service (auto start on boot)
// @Tags         Systemd
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      429 {object} middlewares.TooManyRequests
// @Success      200 {object}  ActionResponse
// @Router       /systemd/enable [post]
func (ctl *Controller) Enable(c echo.Context) error {
	if os.Getenv("SYSTEMD") != "true" {
		return ctl.request.BadRequest(c, errors.New("systemd is not running"))
	}

	statusLog, err := ctl.systemd.Status(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	output := ParseSystemctlShow(statusLog)

	// IMPORTANT CHECK
	if output.UnitFileState == "enabled" {
		return c.JSON(http.StatusOK, ActionResponse{
			Message: "service already enabled",
		})
	}

	err = ctl.systemd.Enable(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, ActionResponse{
		Message: "service enabling started successfully",
	})
}

// Disable
// @Summary      Disable ocserv service
// @Description  Disable ocserv systemd service (remove from auto start)
// @Tags         Systemd
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      429 {object} middlewares.TooManyRequests
// @Success      200 {object}  ActionResponse
// @Router       /systemd/disable [post]
func (ctl *Controller) Disable(c echo.Context) error {
	if os.Getenv("SYSTEMD") != "true" {
		return ctl.request.BadRequest(c, errors.New("systemd is not running"))
	}

	statusLog, err := ctl.systemd.Status(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	output := ParseSystemctlShow(statusLog)

	// IMPORTANT CHECK
	if output.UnitFileState == "disabled" {
		return c.JSON(http.StatusOK, ActionResponse{
			Message: "service already disabled",
		})
	}

	err = ctl.systemd.Disable(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, ActionResponse{
		Message: "service disabling started successfully",
	})
}
