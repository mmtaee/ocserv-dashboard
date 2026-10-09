package servicecontrol

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	servicecontrolusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/service_control"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	service *servicecontrolusecase.Usecase
}

func New(usecase *servicecontrolusecase.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), service: usecase}
}

// Status returns the managed Ocserv runtime status.
// @Summary Ocserv runtime status
// @Tags Service Control
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} StatusResponse
// @Router /systemd/status [get]
func (ctl *Controller) Status(c *echo.Context) error {
	result, err := ctl.service.Status(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Restart restarts the managed Ocserv runtime.
// @Summary Restart Ocserv runtime
// @Tags Service Control
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} ActionResponse
// @Router /systemd/restart [post]
func (ctl *Controller) Restart(c *echo.Context) error {
	return ctl.action(c, ctl.service.Restart)
}

// Enable enables and starts the managed Ocserv runtime.
// @Summary Enable Ocserv runtime
// @Tags Service Control
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} ActionResponse
// @Router /systemd/enable [post]
func (ctl *Controller) Enable(c *echo.Context) error {
	return ctl.action(c, ctl.service.Enable)
}

// Disable disables and stops the managed Ocserv runtime.
// @Summary Disable Ocserv runtime
// @Tags Service Control
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} ActionResponse
// @Router /systemd/disable [post]
func (ctl *Controller) Disable(c *echo.Context) error {
	return ctl.action(c, ctl.service.Disable)
}

func (ctl *Controller) action(c *echo.Context, run func(context.Context) (*servicecontrolusecase.ActionResult, error)) error {
	result, err := run(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}
