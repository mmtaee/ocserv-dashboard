package usecase

import (
	"io"

	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/core/models"
)

type BackupGroupFile struct {
	DefaultGroup *models.OcservGroupConfig `json:"default_group" validate:"required"`
	Groups       []models.OcservGroup      `json:"groups"`
}

type RestoreResponse struct {
	Inserted *[]string `json:"inserted"`
	Existing *[]string `json:"existing"`
}

type BackupUsecase struct {
	ocservUserRepo  repository.OcservUserRepositoryInterface
	ocservGroupRepo repository.OcservGroupRepositoryInterface
	backupRepo      repository.BackupRepositoryInterface
}

type BackupUsecaseInterface interface {
	OcservGroupBackup(writer io.Writer) error
	OcservGroupRestore(owner string, groupData *BackupGroupFile) (*[]string, *[]string, error)
	OcservUserBackup(writer io.Writer) error
	OcservUserRestore(owner string, users []models.OcservUser) (*[]string, *[]string, error)
}

func NewBackupUsecase(
	ocservUserRepo repository.OcservUserRepositoryInterface,
	ocservGroupRepo repository.OcservGroupRepositoryInterface,
	backupRepo repository.BackupRepositoryInterface,
) BackupUsecaseInterface {
	return &BackupUsecase{
		ocservUserRepo:  ocservUserRepo,
		ocservGroupRepo: ocservGroupRepo,
		backupRepo:      backupRepo,
	}
}

func (u *BackupUsecase) OcservGroupBackup(writer io.Writer) error {
	defaultGroup, err := u.ocservGroupRepo.DefaultGroup()
	if err != nil {
		return err
	}
	return u.backupRepo.OcservGroupBackup(nil, writer, defaultGroup)
}

func (u *BackupUsecase) OcservGroupRestore(owner string, groupData *BackupGroupFile) (*[]string, *[]string, error) {
	if err := u.ocservGroupRepo.UpdateDefaultGroup(groupData.DefaultGroup); err != nil {
		return nil, nil, err
	}

	if len(groupData.Groups) > 0 {
		return u.backupRepo.OcservGroupRestore(nil, owner, &groupData.Groups)
	}

	return nil, nil, nil
}

func (u *BackupUsecase) OcservUserBackup(writer io.Writer) error {
	return u.backupRepo.OcservUserBackup(nil, writer)
}

func (u *BackupUsecase) OcservUserRestore(owner string, users []models.OcservUser) (*[]string, *[]string, error) {
	if len(users) == 0 {
		return nil, nil, nil
	}
	return u.backupRepo.OcservUserRestore(nil, owner, &users)
}
