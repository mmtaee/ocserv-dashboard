package occtl

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
)

type Usecase struct {
	Repository
}

func New(repo Repository) *Usecase {
	return &Usecase{Repository: repo}
}

func (u *Usecase) ServerInfo() models.OcservInfo {
	info := models.OcservInfo{Version: u.Version(), Status: "error"}
	status, err := u.Status()
	if err != nil {
		logger.Error("Get server status error: %v", err)
		return info
	}
	values, ok := status.(map[string]interface{})
	if !ok {
		logger.Error("Invalid server status format")
		return info
	}
	if value, ok := values["Status"].(string); ok && value != "" {
		info.Status = value
	}
	return info
}

func (u *Usecase) Command(input CommandInput) (string, error) {
	actions := map[int]func(string) (interface{}, error){
		1:  func(_ string) (interface{}, error) { return u.OnlineSessions() },
		2:  func(value string) (interface{}, error) { return u.ShowUserByUsername(value) },
		3:  func(value string) (interface{}, error) { return u.ShowUserByID(value) },
		4:  func(value string) (interface{}, error) { return u.Disconnect(value) },
		5:  func(_ string) (interface{}, error) { return u.ShowSessionsAll() },
		6:  func(_ string) (interface{}, error) { return u.ShowSessionsValid() },
		7:  func(value string) (interface{}, error) { return u.ShowSessionBySID(value) },
		8:  func(_ string) (interface{}, error) { return u.IPBans() },
		9:  func(value string) (interface{}, error) { return u.UnbanIP(value) },
		10: func(_ string) (interface{}, error) { return u.Status() },
		11: func(_ string) (interface{}, error) { return u.ShowEvent(), nil },
		12: func(_ string) (interface{}, error) { return u.IRoutes() },
		13: func(_ string) (interface{}, error) { return u.Reload() },
		14: func(value string) (interface{}, error) { return u.DisconnectSession(value) },
		15: func(value string) (interface{}, error) { return u.Terminate(value) },
		16: func(value string) (interface{}, error) { return u.TerminateSession(value) },
	}
	handler, ok := actions[input.Action]
	if !ok {
		return "", fmt.Errorf("unknown action %d", input.Action)
	}
	result, err := handler(input.Value)
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(encoded)), nil
}
