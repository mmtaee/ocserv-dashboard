package usecase

import (
	"errors"
	"time"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type OcservUserUseCase interface {
	ListUsers(adminID uint, role string) ([]models.OcservUser, error)
	GetUser(id uint, adminID uint, role string) (*models.OcservUser, error)
	CreateUser(username, password, group string, trafficType string, trafficSize int, description string, config *models.OcservUserConfig, ownerAdminID uint, expireAt *time.Time) (*models.OcservUser, error)
	UpdateUser(id uint, adminID uint, role string, group *string, password *string, expireAt *time.Time, unlimited bool, trafficType *string, trafficSize *int, description *string, config *models.OcservUserConfig) (*models.OcservUser, error)
	DeleteUser(id uint, adminID uint, role string) error
	LockUser(id uint, adminID uint, role string) error
	UnlockUser(id uint, adminID uint, role string) error
}

type ocservUserUseCase struct {
	userRepo repository.OcservUserRepository
}

func NewOcservUserUseCase(userRepo repository.OcservUserRepository) OcservUserUseCase {
	return &ocservUserUseCase{userRepo: userRepo}
}

func (uc *ocservUserUseCase) ListUsers(adminID uint, role string) ([]models.OcservUser, error) {
	return uc.userRepo.FindAll(adminID, role)
}

func (uc *ocservUserUseCase) GetUser(id uint, adminID uint, role string) (*models.OcservUser, error) {
	user, err := uc.userRepo.FindByID(id, adminID, role)
	if err != nil {
		return nil, errors.New("7001")
	}
	return user, nil
}

func (uc *ocservUserUseCase) CreateUser(username, password, group string, trafficType string, trafficSize int, description string, config *models.OcservUserConfig, ownerAdminID uint, expireAt *time.Time) (*models.OcservUser, error) {
	_, err := uc.userRepo.FindByUsername(username, ownerAdminID, models.AdminRoleSuper)
	if err == nil {
		return nil, errors.New("7002")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("5001")
	}

	user := &models.OcservUser{
		OwnerAdminID: ownerAdminID,
		Group:        group,
		Username:     username,
		Password:     string(hashedPassword),
		ExpireAt:     expireAt,
		TrafficType:  trafficType,
		TrafficSize:  trafficSize,
		Description:  description,
		Config:       config,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, errors.New("5001")
	}

	return user, nil
}

func (uc *ocservUserUseCase) UpdateUser(id uint, adminID uint, role string, group *string, password *string, expireAt *time.Time, unlimited bool, trafficType *string, trafficSize *int, description *string, config *models.OcservUserConfig) (*models.OcservUser, error) {
	user, err := uc.userRepo.FindByID(id, adminID, role)
	if err != nil {
		return nil, errors.New("7001")
	}

	if group != nil {
		user.Group = *group
	}
	if password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("5001")
		}
		user.Password = string(hashedPassword)
	}
	if unlimited {
		user.ExpireAt = nil
	} else if expireAt != nil {
		user.ExpireAt = expireAt
	}
	if trafficType != nil {
		user.TrafficType = *trafficType
	}
	if trafficSize != nil {
		user.TrafficSize = *trafficSize
	}
	if description != nil {
		user.Description = *description
	}
	if config != nil {
		user.Config = config
	}

	if err := uc.userRepo.Update(user); err != nil {
		return nil, errors.New("5001")
	}

	return user, nil
}

func (uc *ocservUserUseCase) DeleteUser(id uint, adminID uint, role string) error {
	_, err := uc.userRepo.FindByID(id, adminID, role)
	if err != nil {
		return errors.New("7001")
	}

	if err := uc.userRepo.Delete(id, adminID, role); err != nil {
		return errors.New("5001")
	}

	return nil
}

func (uc *ocservUserUseCase) LockUser(id uint, adminID uint, role string) error {
	_, err := uc.userRepo.FindByID(id, adminID, role)
	if err != nil {
		return errors.New("7001")
	}

	if err := uc.userRepo.Lock(id, adminID, role); err != nil {
		return errors.New("5001")
	}

	return nil
}

func (uc *ocservUserUseCase) UnlockUser(id uint, adminID uint, role string) error {
	_, err := uc.userRepo.FindByID(id, adminID, role)
	if err != nil {
		return errors.New("7001")
	}

	if err := uc.userRepo.Unlock(id, adminID, role); err != nil {
		return errors.New("5001")
	}

	return nil
}
