package repository

import (
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"gorm.io/gorm"
)

type SystemRepository interface {
	Get() (*models.System, error)
	Update(system *models.System) error
}

type systemRepository struct {
	db *gorm.DB
}

func NewSystemRepository(db *gorm.DB) SystemRepository {
	return &systemRepository{db: db}
}

func (r *systemRepository) Get() (*models.System, error) {
	var system models.System
	err := r.db.First(&system).Error
	return &system, err
}

func (r *systemRepository) Update(system *models.System) error {
	return r.db.Save(system).Error
}
