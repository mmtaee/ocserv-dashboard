package system

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

type RuntimeController struct {
	request request.CustomRequestInterface
	system  *systemusecase.Usecase
}

func NewRuntimeController(usecase *systemusecase.Usecase) *RuntimeController {
	return &RuntimeController{request: request.NewCustomRequest(), system: usecase}
}

// Status returns the managed OCServ runtime status.
// @Summary OCServ runtime status
// @Tags System
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} StatusResponse
// @Router /systemd/status [get]
func (ctl *RuntimeController) Status(c *echo.Context) error {
	result, err := ctl.system.Status(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Restart restarts the managed OCServ runtime.
// @Summary Restart OCServ runtime
// @Tags System
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} ActionResponse
// @Router /systemd/restart [post]
func (ctl *RuntimeController) Restart(c *echo.Context) error {
	return ctl.action(c, ctl.system.Restart)
}

// Enable enables and starts the managed OCServ runtime.
// @Summary Enable OCServ runtime
// @Tags System
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} ActionResponse
// @Router /systemd/enable [post]
func (ctl *RuntimeController) Enable(c *echo.Context) error {
	return ctl.action(c, ctl.system.Enable)
}

// Disable disables and stops the managed OCServ runtime.
// @Summary Disable OCServ runtime
// @Tags System
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} ActionResponse
// @Router /systemd/disable [post]
func (ctl *RuntimeController) Disable(c *echo.Context) error {
	return ctl.action(c, ctl.system.Disable)
}

// Config returns the supported main ocserv.conf settings.
// @Summary Get structured OCServ configuration
// @Tags System
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} OcservConfig
// @Router /system/ocserv-config [get]
func (ctl *RuntimeController) Config(c *echo.Context) error {
	result, err := ctl.system.Config(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// UpdateConfig validates, atomically writes, and activates supported settings.
// @Summary Update structured OCServ configuration
// @Tags System
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body OcservConfig true "Supported OCServ configuration changes"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} OcservConfig
// @Router /system/ocserv-config [patch]
func (ctl *RuntimeController) UpdateConfig(c *echo.Context) error {
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

func (ctl *RuntimeController) action(c *echo.Context, run func(context.Context) (*systemusecase.ActionResult, error)) error {
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
