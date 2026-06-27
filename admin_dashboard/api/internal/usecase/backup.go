package usecase

import (
	"context"
	"io"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
)

type BackupUseCase interface {
	OcservGroupBackup(ctx context.Context, writer io.Writer, defaultGroup *models.OcservGroupConfig) error
	OcservGroupRestore(ctx context.Context, ownerAdminID uint, groups *[]models.OcservGroup) (*[]string, *[]string, error)
	OcservUserBackup(ctx context.Context, writer io.Writer) error
	OcservUserRestore(ctx context.Context, ownerAdminID uint, users *[]models.OcservUser) (*[]string, *[]string, error)
}

type backupUseCase struct {
	backupRepo repository.BackupRepository
}

func NewBackupUseCase(backupRepo repository.BackupRepository) BackupUseCase {
	return &backupUseCase{
		backupRepo: backupRepo,
	}
}

func (uc *backupUseCase) OcservGroupBackup(ctx context.Context, writer io.Writer, defaultGroup *models.OcservGroupConfig) error {
	return uc.backupRepo.OcservGroupBackup(ctx, writer, defaultGroup)
}

func (uc *backupUseCase) OcservGroupRestore(ctx context.Context, ownerAdminID uint, groups *[]models.OcservGroup) (*[]string, *[]string, error) {
	return uc.backupRepo.OcservGroupRestore(ctx, ownerAdminID, groups)
}

func (uc *backupUseCase) OcservUserBackup(ctx context.Context, writer io.Writer) error {
	return uc.backupRepo.OcservUserBackup(ctx, writer)
}

func (uc *backupUseCase) OcservUserRestore(ctx context.Context, ownerAdminID uint, users *[]models.OcservUser) (*[]string, *[]string, error) {
	return uc.backupRepo.OcservUserRestore(ctx, ownerAdminID, users)
}
