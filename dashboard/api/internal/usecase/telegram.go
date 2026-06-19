package usecase

import (
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
)

type TelegramUseCase interface {
	// Settings
	GetSettings() (*models.TelegramSettings, error)
	UpdateSettings(updates map[string]interface{}) (*models.TelegramSettings, error)

	// Accounts
	GetAccountsForOcservUser(ocservUserID uint) ([]models.TelegramAccount, error)
	DeleteAccount(id uint) error

	// Packages
	GetPackages(includeInactive bool) ([]models.TelegramPackage, error)
	GetPackageByID(id uint) (*models.TelegramPackage, error)
	CreatePackage(pkg *models.TelegramPackage) (*models.TelegramPackage, error)
	UpdatePackage(id uint, updates map[string]interface{}) (*models.TelegramPackage, error)
	DeletePackage(id uint) error

	// Requests
	GetRequests(page, limit int, orderBy, sort, status, requestType string) ([]models.TelegramRequest, int64, error)
	GetRequestByID(id uint) (*models.TelegramRequest, error)
	UpdateRequestStatus(id uint, status string, adminNote *string) (*models.TelegramRequest, error)
	DeleteRequest(id uint) error
}

type telegramUseCase struct {
	telegramRepo repository.TelegramRepository
}

func NewTelegramUseCase(telegramRepo repository.TelegramRepository) TelegramUseCase {
	return &telegramUseCase{telegramRepo: telegramRepo}
}

// ==========================
// Settings
// ==========================

func (u *telegramUseCase) GetSettings() (*models.TelegramSettings, error) {
	return u.telegramRepo.Settings()
}

func (u *telegramUseCase) UpdateSettings(updates map[string]interface{}) (*models.TelegramSettings, error) {
	return u.telegramRepo.UpdateSettings(updates)
}

// ==========================
// Accounts
// ==========================

func (u *telegramUseCase) GetAccountsForOcservUser(ocservUserID uint) ([]models.TelegramAccount, error) {
	return u.telegramRepo.AccountsForOcservUser(ocservUserID)
}

func (u *telegramUseCase) DeleteAccount(id uint) error {
	return u.telegramRepo.DeleteAccount(id)
}

// ==========================
// Packages
// ==========================

func (u *telegramUseCase) GetPackages(includeInactive bool) ([]models.TelegramPackage, error) {
	return u.telegramRepo.Packages(includeInactive)
}

func (u *telegramUseCase) GetPackageByID(id uint) (*models.TelegramPackage, error) {
	return u.telegramRepo.PackageByID(id)
}

func (u *telegramUseCase) CreatePackage(pkg *models.TelegramPackage) (*models.TelegramPackage, error) {
	return u.telegramRepo.CreatePackage(pkg)
}

func (u *telegramUseCase) UpdatePackage(id uint, updates map[string]interface{}) (*models.TelegramPackage, error) {
	return u.telegramRepo.UpdatePackage(id, updates)
}

func (u *telegramUseCase) DeletePackage(id uint) error {
	return u.telegramRepo.DeletePackage(id)
}

// ==========================
// Requests
// ==========================

func (u *telegramUseCase) GetRequests(page, limit int, orderBy, sort, status, requestType string) ([]models.TelegramRequest, int64, error) {
	return u.telegramRepo.Requests(page, limit, orderBy, sort, status, requestType)
}

func (u *telegramUseCase) GetRequestByID(id uint) (*models.TelegramRequest, error) {
	return u.telegramRepo.RequestByID(id)
}

func (u *telegramUseCase) UpdateRequestStatus(id uint, status string, adminNote *string) (*models.TelegramRequest, error) {
	return u.telegramRepo.UpdateRequestStatus(id, status, adminNote)
}

func (u *telegramUseCase) DeleteRequest(id uint) error {
	return u.telegramRepo.DeleteRequest(id)
}
