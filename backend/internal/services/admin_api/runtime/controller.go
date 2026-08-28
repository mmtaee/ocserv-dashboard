package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v5"
	systemusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/system"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	system  *systemusecase.Usecase
}

func New(usecase *systemusecase.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), system: usecase}
}

// Status returns the managed Ocserv runtime status.
// @Summary Ocserv runtime status
// @Tags System
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} StatusResponse
// @Router /systemd/status [get]
func (ctl *Controller) Status(c *echo.Context) error {
	result, err := ctl.system.Status(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Restart restarts the managed Ocserv runtime.
// @Summary Restart Ocserv runtime
// @Tags System
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} ActionResponse
// @Router /systemd/restart [post]
func (ctl *Controller) Restart(c *echo.Context) error {
	return ctl.action(c, ctl.system.Restart)
}

// Enable enables and starts the managed Ocserv runtime.
// @Summary Enable Ocserv runtime
// @Tags System
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} ActionResponse
// @Router /systemd/enable [post]
func (ctl *Controller) Enable(c *echo.Context) error {
	return ctl.action(c, ctl.system.Enable)
}

// Disable disables and stops the managed Ocserv runtime.
// @Summary Disable Ocserv runtime
// @Tags System
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} ActionResponse
// @Router /systemd/disable [post]
func (ctl *Controller) Disable(c *echo.Context) error {
	return ctl.action(c, ctl.system.Disable)
}

// Config returns the supported main ocserv.conf settings.
// @Summary Get structured Ocserv configuration
// @Tags System
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} OcservConfig
// @Router /system/ocserv-config [get]
func (ctl *Controller) Config(c *echo.Context) error {
	result, err := ctl.system.Config(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// UpdateConfig validates, atomically writes, and activates supported settings.
// @Summary Update structured Ocserv configuration
// @Tags System
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body OcservConfig true "Supported Ocserv configuration changes"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} OcservConfig
// @Router /system/ocserv-config [patch]
func (ctl *Controller) UpdateConfig(c *echo.Context) error {
	var changes OcservConfig
	if err := decodeStrictJSON(c, &changes); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.system.UpdateConfig(c.Request().Context(), changes)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (ctl *Controller) action(c *echo.Context, run func(context.Context) (*systemusecase.ActionResult, error)) error {
	result, err := run(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func decodeStrictJSON(c *echo.Context, target interface{}) error {
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 64<<10)
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body must contain one JSON object")
		}
		return err
	}
	return nil
}
