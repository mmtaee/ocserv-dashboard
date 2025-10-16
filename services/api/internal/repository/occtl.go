package repository

import (
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/occtl"
)

type OcctlRepository struct {
	commonOcservOcctlRepo occtl.OcservOcctlInterface
}

type OcctlRepositoryInterface interface {
	Version() *models.ServerVersion
	Status() (interface{}, error)
	OnlineUsers() (*[]string, error)
	OnlineUsersInfo() (*[]models.OnlineUserSession, error)
	IPBans() (*[]models.IPBanPoints, error)
	IRoutes() (*[]models.IRoute, error)
	Reload() (string, error)
	Disconnect(username string) (string, error)
	ShowUserByUsername(username string) (models.OnlineUserSession, error)
	ShowUserByID(uid string) (models.OnlineUserSession, error)
	ShowSessionsAll() (*[]interface{}, error)
	ShowSessionsValid() (*[]interface{}, error)
	ShowSessionBySID(sid string) (map[string]interface{}, error)
	UnbanIP(ip string) (string, error)
	ShowEvent() string
}

func NewOcctlRepository() *OcctlRepository {
	return &OcctlRepository{commonOcservOcctlRepo: occtl.NewOcservOcctl()}
}

func (o *OcctlRepository) Version() *models.ServerVersion {
	return o.commonOcservOcctlRepo.Version()
}

func (o *OcctlRepository) Status() (interface{}, error) {
	status, err := o.commonOcservOcctlRepo.ShowStatus(false)
	if err != nil {
		return nil, err
	}
	return status, nil
}

func (o *OcctlRepository) OnlineUsers() (*[]string, error) {
	users, err := o.commonOcservOcctlRepo.OnlineUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (o *OcctlRepository) OnlineUsersInfo() (*[]models.OnlineUserSession, error) {
	sessions, err := o.commonOcservOcctlRepo.OnlineSessions()
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (o *OcctlRepository) IPBans() (*[]models.IPBanPoints, error) {
	ipBans, err := o.commonOcservOcctlRepo.ShowIPBans()
	if err != nil {
		return nil, err
	}

	return ipBans, nil
}

func (o *OcctlRepository) IRoutes() (*[]models.IRoute, error) {
	iRoutes, err := o.commonOcservOcctlRepo.ShowIRoutes()
	if err != nil {
		return nil, err
	}
	return iRoutes, nil
}

func (o *OcctlRepository) Reload() (string, error) {
	result, err := o.commonOcservOcctlRepo.ReloadConfigs()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (o *OcctlRepository) Disconnect(username string) (string, error) {
	result, err := o.commonOcservOcctlRepo.DisconnectUser(username)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (o *OcctlRepository) ShowUserByUsername(username string) (models.OnlineUserSession, error) {
	user, err := o.commonOcservOcctlRepo.ShowUser(username)
	if err != nil {
		return models.OnlineUserSession{}, err
	}
	return user, nil
}

func (o *OcctlRepository) ShowUserByID(uid string) (models.OnlineUserSession, error) {
	user, err := o.commonOcservOcctlRepo.ShowUserByID(uid)
	if err != nil {
		return models.OnlineUserSession{}, err
	}
	return user, nil
}

func (o *OcctlRepository) ShowSessionsAll() (*[]interface{}, error) {
	res, err := o.commonOcservOcctlRepo.ShowSessionAll()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *OcctlRepository) ShowSessionsValid() (*[]interface{}, error) {
	res, err := o.commonOcservOcctlRepo.ShowSessionsValid()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *OcctlRepository) ShowSessionBySID(sid string) (map[string]interface{}, error) {
	res, err := o.commonOcservOcctlRepo.ShowSession(sid)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *OcctlRepository) UnbanIP(ip string) (string, error) {
	res, err := o.commonOcservOcctlRepo.UnbanIP(ip)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (o *OcctlRepository) ShowEvent() string {
	return o.commonOcservOcctlRepo.ShowEvent()
}
