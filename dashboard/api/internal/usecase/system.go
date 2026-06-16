package usecase

import (
	"errors"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
)

type SystemUseCase interface {
	Get() (*models.System, error)
	Update(system *models.System) (*models.System, error)
}

type systemUseCase struct {
	systemRepo repository.SystemRepository
}

func NewSystemUseCase(systemRepo repository.SystemRepository) SystemUseCase {
	return &systemUseCase{systemRepo: systemRepo}
}

func (uc *systemUseCase) Get() (*models.System, error) {
	system, err := uc.systemRepo.Get()
	if err != nil {
		return nil, errors.New("3001") // System config not found
	}
	return system, nil
}

func (uc *systemUseCase) Update(system *models.System) (*models.System, error) {
	if err := uc.systemRepo.Update(system); err != nil {
		return nil, errors.New("5001") // Internal server error
	}
	return uc.Get()
}
