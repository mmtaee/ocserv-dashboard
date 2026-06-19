package usecase

import (
	"errors"
	"time"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

type AdminUseCase interface {
	Login(username, password string) (string, *models.Administrator, error)
	GetProfile(adminID uint) (*models.Administrator, error)
	ChangePassword(adminID uint, oldPassword, newPassword string) error
}

type adminUseCase struct {
	adminRepo repository.AdminRepository
}

func NewAdminUseCase(adminRepo repository.AdminRepository) AdminUseCase {
	return &adminUseCase{adminRepo: adminRepo}
}

func (uc *adminUseCase) Login(username, password string) (string, *models.Administrator, error) {
	admin, err := uc.adminRepo.FindByUsername(username)
	if err != nil {
		return "", nil, errors.New("2003") // Invalid credentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", nil, errors.New("2003") // Invalid credentials
	}

	now := time.Now()
	admin.LastLogin = &now
	if err := uc.adminRepo.Update(admin); err != nil {
		return "", nil, errors.New("5001") // Internal server error
	}

	token, err := auth.CreateAdministratorToken(admin.ID, admin.Role, false)
	if err != nil {
		return "", nil, errors.New("5001")
	}

	return token, admin, nil
}

func (uc *adminUseCase) GetProfile(adminID uint) (*models.Administrator, error) {
	return uc.adminRepo.FindByID(adminID)
}

func (uc *adminUseCase) ChangePassword(adminID uint, oldPassword, newPassword string) error {
	admin, err := uc.adminRepo.FindByID(adminID)
	if err != nil {
		return errors.New("5001")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(oldPassword)); err != nil {
		return errors.New("2004") // Current password is incorrect
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("5001")
	}

	admin.Password = string(hashedPassword)
	return uc.adminRepo.Update(admin)
}
