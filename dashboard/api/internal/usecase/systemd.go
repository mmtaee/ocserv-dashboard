package usecase

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
)

type SystemdUseCase interface {
	Status(ctx context.Context) (string, error)
	Restart(ctx context.Context) error
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
	GetMainConfig(ctx context.Context) (*models.OcservMainConfig, error)
	UpdateMainConfig(ctx context.Context, config *models.OcservMainConfig) error
}

type systemdUseCase struct {
	systemdRepo repository.SystemdRepository
}

func NewSystemdUseCase(systemdRepo repository.SystemdRepository) SystemdUseCase {
	return &systemdUseCase{
		systemdRepo: systemdRepo,
	}
}

func (uc *systemdUseCase) Status(ctx context.Context) (string, error) {
	return uc.systemdRepo.Status(ctx)
}

func (uc *systemdUseCase) Restart(ctx context.Context) error {
	return uc.systemdRepo.Restart(ctx)
}

func (uc *systemdUseCase) Enable(ctx context.Context) error {
	return uc.systemdRepo.Enable(ctx)
}

func (uc *systemdUseCase) Disable(ctx context.Context) error {
	return uc.systemdRepo.Disable(ctx)
}

func (uc *systemdUseCase) GetMainConfig(ctx context.Context) (*models.OcservMainConfig, error) {
	return uc.systemdRepo.GetMainConfig(ctx)
}

func (uc *systemdUseCase) UpdateMainConfig(ctx context.Context, config *models.OcservMainConfig) error {
	return uc.systemdRepo.UpdateMainConfig(ctx, config)
}
