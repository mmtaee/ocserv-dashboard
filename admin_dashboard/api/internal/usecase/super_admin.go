package usecase

import (
	"errors"
	"time"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type SuperAdminUseCase interface {
	CreateAdmin(username, password string) (*models.Administrator, error)
	UpdateAdmin(id uint, username string) (*models.Administrator, error)
	ChangeAdminPassword(id uint, newPassword string) error
	SuspendAdmin(id uint, reason string) error
	UnsuspendAdmin(id uint) error
	ListAdmins() ([]models.Administrator, error)
	GetAdmin(id uint) (*models.Administrator, error)
}

type superAdminUseCase struct {
	adminRepo repository.AdminRepository
}

func NewSuperAdminUseCase(adminRepo repository.AdminRepository) SuperAdminUseCase {
	return &superAdminUseCase{adminRepo: adminRepo}
}

func (uc *superAdminUseCase) CreateAdmin(username, password string) (*models.Administrator, error) {
	// Check if username exists
	_, err := uc.adminRepo.FindByUsername(username)
	if err == nil {
		return nil, errors.New("4006")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("5001")
	}

	admin := &models.Administrator{
		Username: username,
		Password: string(hashedPassword),
		Role:     models.AdminRoleAdmin,
	}

	err = uc.adminRepo.Create(admin)
	if err != nil {
		return nil, errors.New("5001")
	}

	return admin, nil
}

func (uc *superAdminUseCase) UpdateAdmin(id uint, username string) (*models.Administrator, error) {
	admin, err := uc.adminRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("4003")
	}

	// Check if trying to modify super admin
	if admin.Role == models.AdminRoleSuper {
		return nil, errors.New("4004")
	}

	// Check if username is already taken by another admin
	if username != admin.Username {
		existingAdmin, err := uc.adminRepo.FindByUsername(username)
		if err == nil && existingAdmin.ID != id {
			return nil, errors.New("4006")
		}
	}

	admin.Username = username

	err = uc.adminRepo.Update(admin)
	if err != nil {
		return nil, errors.New("5001")
	}

	return admin, nil
}

func (uc *superAdminUseCase) ChangeAdminPassword(id uint, newPassword string) error {
	admin, err := uc.adminRepo.FindByID(id)
	if err != nil {
		return errors.New("4003")
	}

	// Check if trying to modify super admin
	if admin.Role == models.AdminRoleSuper {
		return errors.New("4004")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("5001")
	}

	admin.Password = string(hashedPassword)
	return uc.adminRepo.Update(admin)
}

func (uc *superAdminUseCase) SuspendAdmin(id uint, reason string) error {
	admin, err := uc.adminRepo.FindByID(id)
	if err != nil {
		return errors.New("4003")
	}

	// Check if trying to suspend super admin
	if admin.Role == models.AdminRoleSuper {
		return errors.New("4004")
	}

	now := time.Now()
	admin.IsSuspended = true
	admin.SuspendedAt = &now
	admin.SuspendedReason = reason

	return uc.adminRepo.Update(admin)
}

func (uc *superAdminUseCase) UnsuspendAdmin(id uint) error {
	admin, err := uc.adminRepo.FindByID(id)
	if err != nil {
		return errors.New("4003")
	}

	// Check if trying to unsuspend super admin
	if admin.Role == models.AdminRoleSuper {
		return errors.New("4004")
	}

	admin.IsSuspended = false
	admin.SuspendedAt = nil
	admin.SuspendedReason = ""

	return uc.adminRepo.Update(admin)
}

func (uc *superAdminUseCase) ListAdmins() ([]models.Administrator, error) {
	return uc.adminRepo.List()
}

func (uc *superAdminUseCase) GetAdmin(id uint) (*models.Administrator, error) {
	admin, err := uc.adminRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("4003")
	}
	return admin, nil
}
