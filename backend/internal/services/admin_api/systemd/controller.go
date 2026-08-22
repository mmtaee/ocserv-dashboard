package systemd

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	systemdusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/systemd"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	systemd *systemdusecase.Usecase
}

func New(usecase *systemdusecase.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), systemd: usecase}
}

// Status
// @Summary Ocserv systemctl status
// @Tags Systemd
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Success 200 {object} OcservSystemdStatus
// @Router /systemd/status [get]
func (ctl *Controller) Status(c *echo.Context) error {
	result, err := ctl.systemd.Status(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Restart
// @Summary Restart ocserv service
// @Tags Systemd
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Success 200 {object} ActionResponse
// @Router /systemd/restart [post]
func (ctl *Controller) Restart(c *echo.Context) error {
	return ctl.action(c, ctl.systemd.Restart)
}

// Enable
// @Summary Enable ocserv service
// @Tags Systemd
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Success 200 {object} ActionResponse
// @Router /systemd/enable [post]
func (ctl *Controller) Enable(c *echo.Context) error {
	return ctl.action(c, ctl.systemd.Enable)
}

// Disable
// @Summary Disable ocserv service
// @Tags Systemd
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Success 200 {object} ActionResponse
// @Router /systemd/disable [post]
func (ctl *Controller) Disable(c *echo.Context) error {
	return ctl.action(c, ctl.systemd.Disable)
}

func (ctl *Controller) action(c *echo.Context, run func(context.Context) (*systemdusecase.ActionResult, error)) error {
	result, err := run(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}
