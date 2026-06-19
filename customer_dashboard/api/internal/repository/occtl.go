package repository

import (
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/occtl"
)

type OcctlRepository interface {
	OnlineSessions() ([]models.OnlineUserSession, error)
	Disconnect(username string) (string, error)
	DisconnectSession(id string) (string, error)
	Terminate(username string) (string, error)
	TerminateSession(id string) (string, error)
}

type occtlRepository struct {
	occtlRepo occtl.OcservOcctlInterface
}

func NewOcctlRepository(occtlRepo occtl.OcservOcctlInterface) OcctlRepository {
	return &occtlRepository{occtlRepo: occtlRepo}
}

func (r *occtlRepository) OnlineSessions() ([]models.OnlineUserSession, error) {
	sessions, err := r.occtlRepo.OnlineSessions()
	if err != nil {
		return nil, err
	}
	return *sessions, nil
}

func (r *occtlRepository) Disconnect(username string) (string, error) {
	return r.occtlRepo.DisconnectUser(username)
}

func (r *occtlRepository) DisconnectSession(id string) (string, error) {
	return r.occtlRepo.DisconnectSession(id)
}

func (r *occtlRepository) Terminate(username string) (string, error) {
	return r.occtlRepo.TerminateUser(username)
}

func (r *occtlRepository) TerminateSession(id string) (string, error) {
	return r.occtlRepo.TerminateSession(id)
}
