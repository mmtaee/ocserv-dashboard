package occtl

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/internal/repository"
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
	"net/http"
	"strings"
)

type Controller struct {
	request   request.CustomRequestInterface
	occtlRepo repository.OcctlRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request:   request.NewCustomRequest(),
		occtlRepo: repository.NewOcctlRepository(),
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
// @Success      200  {object}  models.ServerVersion
// @Router       /occtl/server_info [get]
func (ctl *Controller) ServerInfo(c echo.Context) error {
	version := ctl.occtlRepo.Version()
	return c.JSON(http.StatusOK, version)
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
		return ctl.request.BadRequest(c, err)
	}

	var results []byte

	actions := map[int]func(string) (interface{}, error){
		1:  func(_ string) (interface{}, error) { return ctl.occtlRepo.OnlineUsersInfo() },
		2:  func(val string) (interface{}, error) { return ctl.occtlRepo.ShowUserByUsername(val) },
		3:  func(val string) (interface{}, error) { return ctl.occtlRepo.ShowUserByID(val) },
		4:  func(val string) (interface{}, error) { return ctl.occtlRepo.Disconnect(val) },
		5:  func(_ string) (interface{}, error) { return ctl.occtlRepo.ShowSessionsAll() },
		6:  func(_ string) (interface{}, error) { return ctl.occtlRepo.ShowSessionsValid() },
		7:  func(val string) (interface{}, error) { return ctl.occtlRepo.ShowSessionBySID(val) },
		8:  func(_ string) (interface{}, error) { return ctl.occtlRepo.IPBans() },
		9:  func(val string) (interface{}, error) { return ctl.occtlRepo.UnbanIP(val) },
		10: func(_ string) (interface{}, error) { return ctl.occtlRepo.Status() },
		11: func(_ string) (interface{}, error) { return ctl.occtlRepo.ShowEvent(), nil },
		12: func(_ string) (interface{}, error) { return ctl.occtlRepo.IRoutes() },
		13: func(_ string) (interface{}, error) { return ctl.occtlRepo.Reload() },
	}

	var err error
	var res interface{}

	handler, exists := actions[data.Action]
	if !exists {
		return ctl.request.BadRequest(c, fmt.Errorf("unknown action %d", data.Action))
	}

	res, err = handler(data.Value)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	results, err = json.Marshal(res)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, strings.TrimSpace(string(results)))
}
