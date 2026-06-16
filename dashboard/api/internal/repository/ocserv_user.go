package repository

import (
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"gorm.io/gorm"
)

type OcservUserRepository interface {
	FindAll(adminID uint, role string) ([]models.OcservUser, error)
	FindByID(id uint, adminID uint, role string) (*models.OcservUser, error)
	FindByUsername(username string, adminID uint, role string) (*models.OcservUser, error)
	Create(user *models.OcservUser) error
	Update(user *models.OcservUser) error
	Delete(id uint, adminID uint, role string) error
	Lock(id uint, adminID uint, role string) error
	Unlock(id uint, adminID uint, role string) error
}

type ocservUserRepository struct {
	db *gorm.DB
}

func NewOcservUserRepository(db *gorm.DB) OcservUserRepository {
	return &ocservUserRepository{db: db}
}

func (r *ocservUserRepository) FindAll(adminID uint, role string) ([]models.OcservUser, error) {
	var users []models.OcservUser
	query := r.db
	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}
	err := query.Find(&users).Error
	return users, err
}

func (r *ocservUserRepository) FindByID(id uint, adminID uint, role string) (*models.OcservUser, error) {
	var user models.OcservUser
	query := r.db.Where("id = ?", id)
	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}
	err := query.First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *ocservUserRepository) FindByUsername(username string, adminID uint, role string) (*models.OcservUser, error) {
	var user models.OcservUser
	query := r.db.Where("username = ?", username)
	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}
	err := query.First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *ocservUserRepository) Create(user *models.OcservUser) error {
	return r.db.Create(user).Error
}

func (r *ocservUserRepository) Update(user *models.OcservUser) error {
	return r.db.Save(user).Error
}

func (r *ocservUserRepository) Delete(id uint, adminID uint, role string) error {
	query := r.db.Where("id = ?", id)
	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}
	return query.Delete(&models.OcservUser{}).Error
}

func (r *ocservUserRepository) Lock(id uint, adminID uint, role string) error {
	query := r.db.Model(&models.OcservUser{}).Where("id = ?", id)
	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}
	return query.Update("is_locked", true).Error
}

func (r *ocservUserRepository) Unlock(id uint, adminID uint, role string) error {
	query := r.db.Model(&models.OcservUser{}).Where("id = ?", id)
	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}
	return query.Update("is_locked", false).Error
}
