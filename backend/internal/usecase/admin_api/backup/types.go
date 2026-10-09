package backup

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/authz"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	groupusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/groups"
)

// Repository is the persistence boundary required by this usecase.
type Repository interface {
	repository.BackupRepositoryInterface
}

type GroupManager interface {
	DefaultGroup() (*models.OcservGroupConfig, error)
	UpdateDefaultGroup(config *models.OcservGroupConfig) error
	Create(ctx context.Context, input groupusecase.CreateInput) (*models.OcservGroup, error)
}

type UserManager interface {
	Create(ctx context.Context, account *models.OcservUser) (*models.OcservUser, error)
	DeleteUser(ctx context.Context, principal authz.Principal, id uint) error
}

type CertificateStore interface {
	CertificateBackup(username string) (*models.OcservUserCertificateBackup, error)
	RestoreCertificateBackup(username string, certificate *models.OcservUserCertificateBackup) error
}

type GroupFile struct {
	DefaultGroup *models.OcservGroupConfig `json:"default_group" validate:"required"`
	Groups       []models.OcservGroup      `json:"groups"`
}

type RestoreResult struct {
	Inserted *[]string `json:"inserted"`
	Existing *[]string `json:"existing"`
}
