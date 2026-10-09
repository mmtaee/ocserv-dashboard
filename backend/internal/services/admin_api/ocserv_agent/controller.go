package ocserv_agent

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/agents"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	agents  *agents.Usecase
}

func New(usecase *agents.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), agents: usecase}
}

// List returns configured master-node Ocserv agents.
// @Summary List Ocserv agents
// @Tags Ocserv Agents
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {array} models.OcservAgent
// @Router /ocserv/agents [get]
func (ctl *Controller) List(c *echo.Context) error {
	result, err := ctl.agents.List(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Get returns one configured Ocserv agent.
// @Summary Get Ocserv agent
// @Tags Ocserv Agents
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Agent ID"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} models.OcservAgent
// @Router /ocserv/agents/{id} [get]
func (ctl *Controller) Get(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.agents.Get(c.Request().Context(), id)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Create stores a manually supplied agent token on the master node.
// @Summary Create Ocserv agent
// @Tags Ocserv Agents
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body CreateInput true "Agent"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 201 {object} models.OcservAgent
// @Router /ocserv/agents [post]
func (ctl *Controller) Create(c *echo.Context) error {
	var input CreateInput
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.agents.Create(c.Request().Context(), input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusCreated, result)
}

// Update replaces an Ocserv agent configuration.
// @Summary Update Ocserv agent
// @Tags Ocserv Agents
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Agent ID"
// @Param request body UpdateInput true "Agent"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} models.OcservAgent
// @Router /ocserv/agents/{id} [patch]
func (ctl *Controller) Update(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	var input UpdateInput
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.agents.Update(c.Request().Context(), id, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Delete removes an Ocserv agent.
// @Summary Delete Ocserv agent
// @Tags Ocserv Agents
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Agent ID"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 204
// @Router /ocserv/agents/{id} [delete]
func (ctl *Controller) Delete(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.agents.Delete(c.Request().Context(), id); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func parseID(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid agent id")
	}
	return uint(id), nil
}
