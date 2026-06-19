package occtl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/occtl"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
)

type OcctlController struct {
	req       *request.Request
	validator *request.Validator
	occtl     *occtl.OcservOcctl
}

func NewOcctlController() *OcctlController {
	return &OcctlController{
		req:       &request.Request{},
		validator: request.NewValidator(),
		occtl:     occtl.NewOcservOcctl(),
	}
}

// ServerInfo godoc
// @Summary Server information
// @Description Server information
// @Tags OCCTL
// @Accept json
// @Produce json
// @Success 200 {object} models.OcservInfo
// @Failure 400 {object} request.ErrorResponse
// @Router /occtl/server_info [get]
func (ctrl *OcctlController) ServerInfo(c *echo.Context) error {
	serverVersion := ctrl.occtl.Version()
	info := models.OcservInfo{
		Version: serverVersion,
		Status:  "error",
	}

	serverStatus, err := ctrl.occtl.ShowStatus(false)
	if err != nil {
		return c.JSON(http.StatusOK, info)
	}

	serverStatusMap, ok := serverStatus.(map[string]interface{})
	if !ok {
		return c.JSON(http.StatusOK, info)
	}

	status := models.ParseServerStatus(serverStatusMap)
	if status.GeneralInfo.Status != "" {
		info.Status = status.GeneralInfo.Status
	}

	return c.JSON(http.StatusOK, info)
}

// Commands godoc
// @Summary Occtl Commands
// @Description Occtl Commands
// @Tags OCCTL
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param action query int true "Command Action ID (1 to 16)"
// @Param value query string false "Optional parameter depending on command"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Success 200 {object} string
// @Router /occtl/commands [get]
func (ctrl *OcctlController) Commands(c *echo.Context) error {
	var data CommandParamsData
	if err := ctrl.validator.Validate(c, &data); err != nil {
		return err
	}

	var results []byte

	actions := map[int]func(string) (interface{}, error){
		1: func(_ string) (interface{}, error) {
			usersPtr, err := ctrl.occtl.OnlineSessions()
			if err != nil {
				return nil, err
			}
			if usersPtr != nil {
				return *usersPtr, nil
			}
			return []models.OnlineUserSession{}, nil
		},
		2: func(val string) (interface{}, error) { return ctrl.occtl.ShowUser(val) },
		3: func(val string) (interface{}, error) { return ctrl.occtl.ShowUserByID(val) },
		4: func(val string) (interface{}, error) { return ctrl.occtl.DisconnectUser(val) },
		5: func(_ string) (interface{}, error) {
			sessionsPtr, err := ctrl.occtl.ShowSessionAll()
			if err != nil {
				return nil, err
			}
			if sessionsPtr != nil {
				return *sessionsPtr, nil
			}
			return []interface{}{}, nil
		},
		6: func(_ string) (interface{}, error) {
			sessionsPtr, err := ctrl.occtl.ShowSessionsValid()
			if err != nil {
				return nil, err
			}
			if sessionsPtr != nil {
				return *sessionsPtr, nil
			}
			return []interface{}{}, nil
		},
		7: func(val string) (interface{}, error) { return ctrl.occtl.ShowSession(val) },
		8: func(_ string) (interface{}, error) {
			ipBansPtr, err := ctrl.occtl.ShowIPBans()
			if err != nil {
				return nil, err
			}
			if ipBansPtr != nil {
				return *ipBansPtr, nil
			}
			return []models.IPBanPoints{}, nil
		},
		9: func(val string) (interface{}, error) { return ctrl.occtl.UnbanIP(val) },
		10: func(_ string) (interface{}, error) { return ctrl.occtl.ShowStatus(false) },
		11: func(_ string) (interface{}, error) { return ctrl.occtl.ShowEvent(), nil },
		12: func(_ string) (interface{}, error) {
			iRoutesPtr, err := ctrl.occtl.ShowIRoutes()
			if err != nil {
				return nil, err
			}
			if iRoutesPtr != nil {
				return *iRoutesPtr, nil
			}
			return []models.IRoute{}, nil
		},
		13: func(_ string) (interface{}, error) { return ctrl.occtl.ReloadConfigs() },
		14: func(val string) (interface{}, error) { return ctrl.occtl.DisconnectSession(val) },
		15: func(val string) (interface{}, error) { return ctrl.occtl.TerminateUser(val) },
		16: func(val string) (interface{}, error) { return ctrl.occtl.TerminateSession(val) },
	}

	var err error
	var res interface{}

	handler, exists := actions[data.Action]
	if !exists {
		return ctrl.req.BadRequest(c, fmt.Errorf("unknown action %d", data.Action))
	}

	res, err = handler(data.Value)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	results, err = json.Marshal(res)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, strings.TrimSpace(string(results)))
}
