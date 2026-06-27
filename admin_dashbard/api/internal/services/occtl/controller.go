package occtl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
)

type CommandParamsData struct {
	Action int    `query:"action" validate:"required,min=1,max=16"`
	Value  string `query:"value" validate:"omitempty"`
}

type Controller struct {
	request      request.CustomRequestInterface
	occtlUsecase usecase.OcctlUsecaseInterface
}

func New(occtlUsecase usecase.OcctlUsecaseInterface) *Controller {
	return &Controller{
		request:      request.NewCustomRequest(),
		occtlUsecase: occtlUsecase,
	}
}

// ServerInfo 	 Server information
//
// @Summary      Server information
// @Description  Server information
// @Tags         OCCTL
// @Accept       json
// @Produce      json
// @Failure      400 {object} request.ErrorResponse
// @Success      200  {object}  models.OcservInfo
// @Router       /occtl/server_info [get]
func (ctl *Controller) ServerInfo(c echo.Context) error {
	info, _ := ctl.occtlUsecase.GetServerInfo()
	return c.JSON(http.StatusOK, info)
}

// Commands 	 Occtl Commands
//
// @Summary      Occtl Commands
// @Description  Occtl Commands
// @Tags         OCCTL
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param        action  query   int     true   "Command Action ID (1 to 15)"
// @Param        value   query   string  false  "Optional parameter depending on command"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object}  string
// @Router       /occtl/commands [get]
func (ctl *Controller) Commands(c echo.Context) error {
	var data CommandParamsData
	if err := c.Bind(&data); err != nil {
		return ctl.request.BadRequest(c, err, "1003")
	}

	res, err := ctl.occtlUsecase.ExecuteCommand(data.Action, data.Value)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}

	if res == nil {
		return ctl.request.BadRequest(c, fmt.Errorf("unknown action %d", data.Action), "1033")
	}

	results, err := json.Marshal(res)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(http.StatusOK, strings.TrimSpace(string(results)))
}
