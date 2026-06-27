package systemd

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	usecase usecase.SystemdUsecaseInterface
}

func New(systemdUsecase usecase.SystemdUsecaseInterface) *Controller {
	return &Controller{
		request: request.NewCustomRequest(),
		usecase: systemdUsecase,
	}
}

func (ctl *Controller) Status(c echo.Context) error {
	status, err := ctl.usecase.Status()
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, status)
}

func (ctl *Controller) Restart(c echo.Context) error {
	err := ctl.usecase.Restart()
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(http.StatusOK, usecase.ActionResponse{
		Message: "service restarting started successfully",
	})
}

func (ctl *Controller) Enable(c echo.Context) error {
	message, err := ctl.usecase.Enable()
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(http.StatusOK, usecase.ActionResponse{
		Message: message,
	})
}

func (ctl *Controller) Disable(c echo.Context) error {
	message, err := ctl.usecase.Disable()
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(http.StatusOK, usecase.ActionResponse{
		Message: message,
	})
}
