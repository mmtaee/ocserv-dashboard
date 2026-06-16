package repository

import (
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"gorm.io/gorm"
)

type AdminRepository interface {
	FindByUsername(username string) (*models.Administrator, error)
	FindByID(id uint) (*models.Administrator, error)
	Create(admin *models.Administrator) error
	Update(admin *models.Administrator) error
	List() ([]models.Administrator, error)
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) FindByUsername(username string) (*models.Administrator, error) {
	var admin models.Administrator
	err := r.db.Where("username = ?", username).First(&admin).Error
	return &admin, err
}

func (r *adminRepository) FindByID(id uint) (*models.Administrator, error) {
	var admin models.Administrator
	err := r.db.First(&admin, id).Error
	return &admin, err
}

func (r *adminRepository) Create(admin *models.Administrator) error {
	return r.db.Create(admin).Error
}

func (r *adminRepository) Update(admin *models.Administrator) error {
	return r.db.Save(admin).Error
}

func (r *adminRepository) List() ([]models.Administrator, error) {
	var admins []models.Administrator
	err := r.db.Find(&admins).Error
	return admins, err
}
