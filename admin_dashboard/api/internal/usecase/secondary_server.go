package usecase

import (
	"errors"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
)

type SecondaryServerUseCase interface {
	GetAll() ([]models.SecondaryServer, error)
	GetByID(id uint) (*models.SecondaryServer, error)
	Create(title, ip string, port int, token string) (*models.SecondaryServer, error)
	Update(id uint, title, ip string, port int, token string) (*models.SecondaryServer, error)
	Delete(id uint) error
}

type secondaryServerUseCase struct {
	serverRepo repository.SecondaryServerRepository
}

func NewSecondaryServerUseCase(serverRepo repository.SecondaryServerRepository) SecondaryServerUseCase {
	return &secondaryServerUseCase{serverRepo: serverRepo}
}

func (uc *secondaryServerUseCase) GetAll() ([]models.SecondaryServer, error) {
	return uc.serverRepo.FindAll()
}

func (uc *secondaryServerUseCase) GetByID(id uint) (*models.SecondaryServer, error) {
	server, err := uc.serverRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("9001")
	}
	return server, nil
}

func (uc *secondaryServerUseCase) Create(title, ip string, port int, token string) (*models.SecondaryServer, error) {
	server := &models.SecondaryServer{
		Title: title,
		IP:    ip,
		Port:  port,
		Token: token,
	}
	if err := uc.serverRepo.Create(server); err != nil {
		return nil, errors.New("5001")
	}
	return server, nil
}

func (uc *secondaryServerUseCase) Update(id uint, title, ip string, port int, token string) (*models.SecondaryServer, error) {
	server, err := uc.serverRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("9001")
	}
	
	server.Title = title
	server.IP = ip
	server.Port = port
	server.Token = token
	
	if err := uc.serverRepo.Update(server); err != nil {
		return nil, errors.New("5001")
	}
	return server, nil
}

func (uc *secondaryServerUseCase) Delete(id uint) error {
	if _, err := uc.serverRepo.FindByID(id); err != nil {
		return errors.New("9001")
	}
	if err := uc.serverRepo.Delete(id); err != nil {
		return errors.New("5001")
	}
	return nil
}
