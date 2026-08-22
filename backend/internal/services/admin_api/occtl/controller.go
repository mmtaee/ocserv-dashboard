package occtl

import (
	"net/http"

	"github.com/labstack/echo/v5"
	occtlusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/occtl"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request   request.CustomRequestInterface
	occtlRepo *occtlusecase.Usecase
}

func New(usecase *occtlusecase.Usecase) *Controller {
	return &Controller{
		request:   request.NewCustomRequest(),
		occtlRepo: usecase,
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
func (ctl *Controller) ServerInfo(c *echo.Context) error {
	return c.JSON(http.StatusOK, ctl.occtlRepo.ServerInfo())
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
func (ctl *Controller) Commands(c *echo.Context) error {
	var data CommandParamsData
	if err := c.Bind(&data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	result, err := ctl.occtlRepo.Command(data)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}
