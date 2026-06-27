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
	ChangePassword(adminID uint, oldPassword, newPassword string) (string, *models.Administrator, error)
	Logout(token string) error
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

	token, err := auth.CreateAdministratorToken()
	if err != nil {
		return "", nil, errors.New("5001")
	}

	adminToken := &models.AdministratorToken{
		AdministratorID: admin.ID,
		Token:           token,
	}

	if err := uc.adminRepo.CreateToken(adminToken); err != nil {
		return "", nil, errors.New("5001")
	}

	return token, admin, nil
}

func (uc *adminUseCase) GetProfile(adminID uint) (*models.Administrator, error) {
	return uc.adminRepo.FindByID(adminID)
}

func (uc *adminUseCase) ChangePassword(adminID uint, oldPassword, newPassword string) (string, *models.Administrator, error) {
	admin, err := uc.adminRepo.FindByID(adminID)
	if err != nil {
		return "", nil, errors.New("5001")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(oldPassword)); err != nil {
		return "", nil, errors.New("2004") // Current password is incorrect
	}

	// Delete all existing tokens for this admin
	if err := uc.adminRepo.DeleteAllTokensByAdmin(adminID); err != nil {
		return "", nil, errors.New("5001")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, errors.New("5001")
	}

	admin.Password = string(hashedPassword)
	if err := uc.adminRepo.Update(admin); err != nil {
		return "", nil, errors.New("5001")
	}

	// Create new token
	newToken, err := auth.CreateAdministratorToken()
	if err != nil {
		return "", nil, errors.New("5001")
	}

	adminToken := &models.AdministratorToken{
		AdministratorID: admin.ID,
		Token:           newToken,
	}

	if err := uc.adminRepo.CreateToken(adminToken); err != nil {
		return "", nil, errors.New("5001")
	}

	return newToken, admin, nil
}

func (uc *adminUseCase) Logout(token string) error {
	return uc.adminRepo.DeleteToken(token)
}
