package agent_settings

import (
	"net/http"

	"github.com/labstack/echo/v5"
	agentsettings "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/agent_settings"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request  request.CustomRequestInterface
	settings *agentsettings.Usecase
}

func New(usecase *agentsettings.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), settings: usecase}
}

// GetToken gets or creates the local agent token.
// @Summary Get or create local agent token
// @Tags Agent Settings
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} models.AgentToken
// @Router /agent/settings/token [get]
func (ctl *Controller) GetToken(c *echo.Context) error {
	result, err := ctl.settings.Get(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// RenewToken replaces the local agent token.
// @Summary Renew local agent token
// @Tags Agent Settings
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} models.AgentToken
// @Router /agent/settings/token/renew [post]
func (ctl *Controller) RenewToken(c *echo.Context) error {
	result, err := ctl.settings.Renew(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// RemoveToken revokes the local agent token.
// @Summary Remove local agent token
// @Tags Agent Settings
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 204
// @Router /agent/settings/token [delete]
func (ctl *Controller) RemoveToken(c *echo.Context) error {
	if err := ctl.settings.Remove(c.Request().Context()); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
