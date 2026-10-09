package ocserv

import (
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/occtl"
)

type Client struct {
	commonOcservOcctlRepo occtl.OcservOcctlInterface
}

type OcctlServerInfo interface {
	Version() *models.ServerVersion
	Status() (interface{}, error)
	ShowEvent() string
}

type OcctlUserManager interface {
	OnlineSessions() ([]models.OnlineUserSession, error)
	ShowUserByUsername(username string) (models.OnlineUserSession, error)
	ShowUserByID(id string) (models.OnlineUserSession, error)
	ShowSessionsAll() (*[]interface{}, error)
	ShowSessionsValid() (*[]interface{}, error)
	ShowSessionBySID(sid string) (map[string]interface{}, error)

	Disconnect(username string) (string, error)
	DisconnectSession(id string) (string, error)

	Terminate(id string) (string, error)
	TerminateSession(id string) (string, error)
}

type OcctlSecurityManager interface {
	IPBans() (*[]models.IPBanPoints, error)
	UnbanIP(ip string) (string, error)
	IRoutes() (*[]models.IRoute, error)
	Reload() (string, error)
}

type ClientInterface interface {
	OcctlServerInfo
	OcctlUserManager
	OcctlSecurityManager
}

func NewClient() *Client {
	return &Client{commonOcservOcctlRepo: occtl.NewOcservOcctl()}
}

func (o *Client) Version() *models.ServerVersion {
	return o.commonOcservOcctlRepo.Version()
}

func (o *Client) Status() (interface{}, error) {
	status, err := o.commonOcservOcctlRepo.ShowStatus(false)
	if err != nil {
		return nil, err
	}
	return status, nil
}

func (o *Client) OnlineSessions() ([]models.OnlineUserSession, error) {
	users, err := o.commonOcservOcctlRepo.OnlineSessions()
	if err != nil {
		return nil, err
	}
	return users, nil
}

//func (o *Client) OnlineUsersInfo() (*[]models.OnlineUserSession, error) {
//	sessions, err := o.commonOcservOcctlRepo.OnlineSessions()
//	if err != nil {
//		return nil, err
//	}
//
//	return sessions, nil
//}

func (o *Client) IPBans() (*[]models.IPBanPoints, error) {
	ipBans, err := o.commonOcservOcctlRepo.ShowIPBans()
	if err != nil {
		return nil, err
	}

	return ipBans, nil
}

func (o *Client) IRoutes() (*[]models.IRoute, error) {
	iRoutes, err := o.commonOcservOcctlRepo.ShowIRoutes()
	if err != nil {
		return nil, err
	}
	return iRoutes, nil
}

func (o *Client) Reload() (string, error) {
	result, err := o.commonOcservOcctlRepo.ReloadConfigs()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (o *Client) Disconnect(username string) (string, error) {
	result, err := o.commonOcservOcctlRepo.DisconnectUser(username)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (o *Client) DisconnectSession(id string) (string, error) {
	result, err := o.commonOcservOcctlRepo.DisconnectSession(id)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (o *Client) Terminate(username string) (string, error) {
	result, err := o.commonOcservOcctlRepo.TerminateUser(username)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (o *Client) TerminateSession(id string) (string, error) {
	result, err := o.commonOcservOcctlRepo.TerminateSession(id)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (o *Client) ShowUserByUsername(username string) (models.OnlineUserSession, error) {
	user, err := o.commonOcservOcctlRepo.ShowUser(username)
	if err != nil {
		return models.OnlineUserSession{}, err
	}
	return user, nil
}

func (o *Client) ShowUserByID(id string) (models.OnlineUserSession, error) {
	user, err := o.commonOcservOcctlRepo.ShowUserByID(id)
	if err != nil {
		return models.OnlineUserSession{}, err
	}
	return user, nil
}

func (o *Client) ShowSessionsAll() (*[]interface{}, error) {
	res, err := o.commonOcservOcctlRepo.ShowSessionAll()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *Client) ShowSessionsValid() (*[]interface{}, error) {
	res, err := o.commonOcservOcctlRepo.ShowSessionsValid()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *Client) ShowSessionBySID(sid string) (map[string]interface{}, error) {
	res, err := o.commonOcservOcctlRepo.ShowSession(sid)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *Client) UnbanIP(ip string) (string, error) {
	res, err := o.commonOcservOcctlRepo.UnbanIP(ip)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (o *Client) ShowEvent() string {
	return o.commonOcservOcctlRepo.ShowEvent()
}
