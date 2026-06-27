package repository

import (
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"gorm.io/gorm"
)

type SecondaryServerRepository interface {
	FindAll() ([]models.SecondaryServer, error)
	FindByID(id uint) (*models.SecondaryServer, error)
	Create(server *models.SecondaryServer) error
	Update(server *models.SecondaryServer) error
	Delete(id uint) error
}

type secondaryServerRepository struct {
	db *gorm.DB
}

func NewSecondaryServerRepository(db *gorm.DB) SecondaryServerRepository {
	return &secondaryServerRepository{db: db}
}

func (r *secondaryServerRepository) FindAll() ([]models.SecondaryServer, error) {
	var servers []models.SecondaryServer
	err := r.db.Find(&servers).Error
	return servers, err
}

func (r *secondaryServerRepository) FindByID(id uint) (*models.SecondaryServer, error) {
	var server models.SecondaryServer
	err := r.db.First(&server, id).Error
	return &server, err
}

func (r *secondaryServerRepository) Create(server *models.SecondaryServer) error {
	return r.db.Create(server).Error
}

func (r *secondaryServerRepository) Update(server *models.SecondaryServer) error {
	return r.db.Save(server).Error
}

func (r *secondaryServerRepository) Delete(id uint) error {
	return r.db.Delete(&models.SecondaryServer{}, id).Error
}
