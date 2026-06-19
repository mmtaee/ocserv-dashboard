package repository

import (
	"time"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/user"
	"gorm.io/gorm"
)

type OcservUserRepository interface {
	FindByUsername(username string) (*models.OcservUser, error)
	CreateCertificate(username string) error
	CertificatePath(username string) (string, error)
	UserStatistics(id uint, startDate, endDate *time.Time) ([]models.DailyTraffic, error)
	UpdatePassword(username string, newPassword string) error
	UserSessionLogs(username string, page, limit int, orderBy, sort string, startDate, endDate *time.Time) ([]models.OcservUserSessionLog, int64, error)
}

type ocservUserRepository struct {
	db                 *gorm.DB
	ocservUserRepo     user.OcservUserInterface
}

func NewOcservUserRepository(db *gorm.DB, ocservUserRepo user.OcservUserInterface) OcservUserRepository {
	return &ocservUserRepository{
		db:             db,
		ocservUserRepo: ocservUserRepo,
	}
}

func (r *ocservUserRepository) FindByUsername(username string) (*models.OcservUser, error) {
	var user models.OcservUser
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *ocservUserRepository) CreateCertificate(username string) error {
	var user models.OcservUser
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return err
	}
	return r.ocservUserRepo.CreateCertificate(username, user.Password)
}

func (r *ocservUserRepository) CertificatePath(username string) (string, error) {
	return r.ocservUserRepo.CertificatePath(username)
}

func (r *ocservUserRepository) UserStatistics(id uint, startDate, endDate *time.Time) ([]models.DailyTraffic, error) {
	var stats []models.DailyTraffic
	query := r.db.Model(&models.OcservUserTrafficStatistics{}).Where("oc_user_id = ?", id)

	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}

	err := query.Select("DATE(created_at) as date, SUM(rx)/1073741824.0 as rx, SUM(tx)/1073741824.0 as tx").
		Group("DATE(created_at)").
		Order("date desc").
		Find(&stats).Error
	return stats, err
}

func (r *ocservUserRepository) UpdatePassword(username string, newPassword string) error {
	var user models.OcservUser
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return err
	}

	user.Password = newPassword
	if err := r.db.Save(&user).Error; err != nil {
		return err
	}

	if err := r.ocservUserRepo.Create(user.Group, user.Username, newPassword, user.Config); err != nil {
		return err
	}

	return nil
}

func (r *ocservUserRepository) UserSessionLogs(username string, page, limit int, orderBy, sort string, startDate, endDate *time.Time) ([]models.OcservUserSessionLog, int64, error) {
	var logs []models.OcservUserSessionLog
	var total int64

	query := r.db.Model(&models.OcservUserSessionLog{}).Where("username = ?", username)

	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at < ?", endDate.AddDate(0, 0, 1))
	}

	validOrderFields := map[string]bool{"id": true, "created_at": true}
	validSortOrders := map[string]bool{"asc": true, "desc": true}

	if !validOrderFields[orderBy] {
		orderBy = "id"
	}
	if !validSortOrders[sort] {
		sort = "desc"
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order(orderBy + " " + sort).
		Offset(offset).
		Limit(limit).
		Find(&logs).Error

	return logs, total, err
}
