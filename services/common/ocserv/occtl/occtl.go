package occtl

import (
	"encoding/json"
	"fmt"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/utils"
	"log"
	"net"
	"os/exec"
	"strings"
)

type OcservOcctl struct{}

type OcservOcctlInterface interface {
	OnlineUsers() (*[]string, error)
	OnlineSessions() (*[]models.OnlineUserSession, error)
	DisconnectUser(username string) (string, error)
	ReloadConfigs() (string, error)
	ShowIPBans() (*[]models.IPBanPoints, error)
	UnbanIP(ip string) (string, error)
	ShowStatus(raw bool) (interface{}, error)
	ShowIRoutes() (*[]models.IRoute, error)
	ShowUser(username string) (models.OnlineUserSession, error)
	Version() *models.ServerVersion
	ShowUserByID(id string) (models.OnlineUserSession, error)
	ShowSession(sid string) (map[string]interface{}, error)
	ShowSessionAll() (*[]interface{}, error)
	ShowSessionsValid() (*[]interface{}, error)
	ShowEvent() string
}

const occtlExec = "/usr/bin/occtl"

func NewOcservOcctl() *OcservOcctl {
	return &OcservOcctl{}
}

// OnlineUsers returns a list of currently connected usernames.
// Executes: occtl -j show users | jq -r '.[].Username'
func (o *OcservOcctl) OnlineUsers() (*[]string, error) {
	var users []string

	command := "-j show users | jq -r '.[].Username'"
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", occtlExec, command))
	result, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	for _, user := range strings.Split(string(result), "\n") {
		trimmed := strings.TrimSpace(user)
		if trimmed != "" {
			users = append(users, trimmed)
		}
	}

	return &users, nil
}

// OnlineSessions returns a list of currently connected user info.
// Executes: occtl -j show users
func (o *OcservOcctl) OnlineSessions() (*[]models.OnlineUserSession, error) {
	command := "-j show users"
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", occtlExec, command))
	result, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var sessions []models.OnlineUserSession
	if err = json.Unmarshal(result, &sessions); err != nil {
		return nil, err
	}
	return &sessions, nil
}

// DisconnectUser disconnects the given user.
// Executes: occtl disconnect user <username>
func (o *OcservOcctl) DisconnectUser(username string) (string, error) {
	cmd := exec.Command(occtlExec, "disconnect", "user", username)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ReloadConfigs reloads the ocserv configuration.
// Executes: occtl reload
func (o *OcservOcctl) ReloadConfigs() (string, error) {
	cmd := exec.Command(occtlExec, "reload")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ShowIPBans returns the current list of IP bans with scores.
// Executes: occtl -j show ip bans points
func (o *OcservOcctl) ShowIPBans() (*[]models.IPBanPoints, error) {
	cmd := exec.Command(occtlExec, "-j", "show", "ip", "bans", "points")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var ipBans []models.IPBanPoints
	if output := string(out); output != "" {
		if err = json.Unmarshal(out, &ipBans); err != nil {
			return nil, err
		}
	}

	return &ipBans, nil

}

// UnbanIP removes an IP ban from the given IP address.
// Executes: occtl unban ip <ip>
func (o *OcservOcctl) UnbanIP(ip string) (string, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", fmt.Errorf("invalid IP: %s", ip)
	}

	cmd := exec.Command(occtlExec, "unban", "ip", ip)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ShowStatus returns the current status of ocserv.
// Executes: occtl -j show status
// Executes: occtl show status
func (o *OcservOcctl) ShowStatus(raw bool) (interface{}, error) {
	cmd := exec.Command(occtlExec, "-j", "show", "status")
	if raw {
		cmd = exec.Command(occtlExec, "show", "status")
	}

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if raw {
		return string(out), nil
	}

	var status map[string]interface{}
	if err = json.Unmarshal(out, &status); err != nil {
		return nil, err
	}
	return status, nil
}

// ShowIRoutes returns the current iRoutes information.
// Executes: occtl -j show iroutes
func (o *OcservOcctl) ShowIRoutes() (*[]models.IRoute, error) {
	var routes []models.IRoute

	version := utils.GetOcservVersion()
	if version == "1.2.4" { // has bug on IRoute Command
		return &routes, nil
	}

	log.Println("common ShowIRoutes version: ", version)

	cmd := exec.Command(occtlExec, "-j", "show", "iroutes")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	outString := strings.TrimSpace(string(out))
	if strings.HasPrefix(outString, "{") {
		return &routes, nil
	}

	if err = json.Unmarshal(out, &routes); err != nil {
		return nil, err
	}
	return &routes, nil
}

// ShowUser returns detailed information about a specific user by username.
// Executes: occtl -j show user <username>
func (o *OcservOcctl) ShowUser(username string) (models.OnlineUserSession, error) {
	var session models.OnlineUserSession

	cmd := exec.Command(occtlExec, "-j", "show", "user", username)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return session, err
	}
	if err = json.Unmarshal(out, &session); err != nil {
		return session, err
	}
	return session, nil
}

// Version returns detailed information about ocserv version.
// Executes: ocserv -v
func (o *OcservOcctl) Version() *models.ServerVersion {
	version := utils.GetOcservVersion()
	occtlVersion := utils.GetOCCTLVersion()
	return &models.ServerVersion{
		Version:      version,
		OcctlVersion: occtlVersion,
	}
}

// ShowUserByID returns detailed information about a specific user by ID.
// Executes: occtl -j show id <id>
func (o *OcservOcctl) ShowUserByID(id string) (models.OnlineUserSession, error) {
	var session models.OnlineUserSession
	cmd := exec.Command(occtlExec, "-j", "show", "id", id)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return session, err
	}
	if err = json.Unmarshal(out, &session); err != nil {
		return session, err
	}
	return session, nil
}

// ShowSession returns detailed information about a specific session by SID.
// Executes: occtl -j show session <SID>
func (o *OcservOcctl) ShowSession(sid string) (map[string]interface{}, error) {
	cmd := exec.Command(occtlExec, "-j", "show", "session", sid)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var session map[string]interface{}
	if err = json.Unmarshal(out, &session); err != nil {
		return nil, err
	}
	return session, nil
}

// ShowSessionAll returns detailed information about all sessions.
// Executes: occtl -j show sessions all
func (o *OcservOcctl) ShowSessionAll() (*[]interface{}, error) {
	var sessions []interface{}

	cmd := exec.Command(occtlExec, "-j", "show", "sessions", "all")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(out, &sessions); err != nil {
		return nil, err
	}
	return &sessions, nil
}

// ShowSessionsValid returns detailed information  about all valid sessions.
// Executes: occtl -j show sessions valid
func (o *OcservOcctl) ShowSessionsValid() (*[]interface{}, error) {
	var sessions []interface{}

	cmd := exec.Command(occtlExec, "-j", "show", "sessions", "valid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(out, &sessions); err != nil {
		return nil, err
	}
	return &sessions, nil
}

// ShowEvent returns detailed information about events.
// Executes: occtl -j show events
func (o *OcservOcctl) ShowEvent() string {
	cmd := exec.Command(occtlExec, "-j", "show", "events")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err.Error()
	}

	go func() {
		defer stdin.Close()
		_, _ = stdin.Write([]byte("q\n"))
	}()

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	cleaned := strings.Replace(string(output), "Press 'q' or CTRL+C to quit", "", 1)
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}
