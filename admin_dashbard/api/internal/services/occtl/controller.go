package occtl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
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
// @Success      200  {object}  models.OcservInfo
// @Router       /occtl/server_info [get]
func (ctl *Controller) ServerInfo(c echo.Context) error {
	serverVersion := ctl.occtlRepo.Version()
	info := models.OcservInfo{
		Version: serverVersion,
		Status:  "error",
	}

	serverStatus, err := ctl.occtlRepo.Status()
	if err != nil {
		logger.Error("Get server status error: %v", err)
		info.Status = "error"
		return c.JSON(http.StatusOK, info)
	}

	serverStatusMap, ok := serverStatus.(map[string]interface{})
	if !ok {
		logger.Error("Invalid server status format")
		info.Status = "error"
		return c.JSON(http.StatusOK, info)
	}

	status := models.ParseOcservServerStatus(serverStatusMap)
	if status.GeneralInfo.Status != "" {
		info.Status = status.GeneralInfo.Status
	}

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
		return ctl.request.BadRequest(c, err)
	}

	var results []byte

	actions := map[int]func(string) (interface{}, error){
		1:  func(_ string) (interface{}, error) { return ctl.occtlRepo.OnlineSessions() },
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
		14: func(val string) (interface{}, error) { return ctl.occtlRepo.DisconnectSession(val) },
		15: func(val string) (interface{}, error) { return ctl.occtlRepo.Terminate(val) },
		16: func(val string) (interface{}, error) { return ctl.occtlRepo.TerminateSession(val) },
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
