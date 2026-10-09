package runtime

import (
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

// Config returns the supported main ocserv.conf settings.
// @Summary Get structured Ocserv configuration
// @Tags System
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} ConfigResponse
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
// @Description Available only in normal/systemd mode; Docker deployments reject updates.
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
