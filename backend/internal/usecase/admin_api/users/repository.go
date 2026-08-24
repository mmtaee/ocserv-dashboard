package ocservuser

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
)

// Repository is the persistence boundary required by this usecase.
type Repository interface {
	repository.OcservUserRepositoryInterface
	repository.OcservUserBulkRepository
}

type AccountStore interface {
	user.OcservUserInterface
}

type OCCTL interface {
	OnlineSessions() ([]models.OnlineUserSession, error)
	ShowUserByID(id string) (models.OnlineUserSession, error)
	Terminate(username string) (string, error)
	Disconnect(username string) (string, error)
	DisconnectSession(id string) (string, error)
	TerminateSession(id string) (string, error)
	Reload() (string, error)
}

type Reports interface {
	TotalBandWidthUser(ctx context.Context, id uint) (repository.TotalBandwidths, error)
}
