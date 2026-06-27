package repository

import (
	"time"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/user"
	"gorm.io/gorm"
)

type OcservUserRepository interface {
	FindAll(adminID uint, role string) ([]models.OcservUser, error)
	FindAllPaginated(adminID uint, role string, page, limit int, q, filter, group string, orderBy string, sort string) ([]models.OcservUser, int64, error)
	FindByID(id uint, adminID uint, role string) (*models.OcservUser, error)
	FindByIDUnrestricted(id uint) (*models.OcservUser, error)
	FindByUsername(username string, adminID uint, role string) (*models.OcservUser, error)
	Create(user *models.OcservUser) error
	CreateUnrestricted(user *models.OcservUser) (*models.OcservUser, error)
	Update(user *models.OcservUser) error
	UpdateUnrestricted(user *models.OcservUser) error
	Delete(id uint, adminID uint, role string) error
	Lock(id uint, adminID uint, role string) error
	Unlock(id uint, adminID uint, role string) error
	UpdateUsersByDeleteGroup(ownerAdminID uint, oldGroupName string) (int64, error)
	UserSessionLogs(username string, page, limit int, orderBy, sort string, startDate, endDate *time.Time) ([]models.OcservUserSessionLog, int64, error)
	UserStatistics(id uint, startDate, endDate *time.Time) ([]models.DailyTraffic, error)
	RestoreExpired(id uint, expireAt *time.Time) error
	CreateCertificate(id uint) error
	CertificatePath(id uint) (string, string, error)
}

type ocservUserRepository struct {
	db               *gorm.DB
	commonOcservUser user.OcservUserInterface
}

func NewOcservUserRepository(db *gorm.DB) OcservUserRepository {
	return &ocservUserRepository{
		db:               db,
		commonOcservUser: user.NewOcservUser(),
	}
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

func (r *ocservUserRepository) FindAllPaginated(adminID uint, role string, page, limit int, q, filter, group string, orderBy string, sort string) ([]models.OcservUser, int64, error) {
	var users []models.OcservUser
	var total int64

	query := r.db.Model(&models.OcservUser{})
	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}

	if q != "" {
		query = query.Where("username LIKE ?", "%"+q+"%")
	}

	if group != "" {
		query = query.Where("`group` = ?", group)
	}

	switch filter {
	case "active":
		query = query.Where("is_locked = ? AND deactivated_at IS NULL", false)
	case "deactivated":
		query = query.Where("deactivated_at IS NOT NULL")
	case "locked":
		query = query.Where("is_locked = ?", true)
	}

	// Set default order if not provided or invalid
	validOrderFields := map[string]bool{"id": true, "created_at": true}
	validSortOrders := map[string]bool{"asc": true, "desc": true}

	// Apply defaults
	if !validOrderFields[orderBy] {
		orderBy = "id"
	}
	if !validSortOrders[sort] {
		sort = "desc"
	}

	// Apply order
	query = query.Order(orderBy + " " + sort)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginate
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
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

func (r *ocservUserRepository) FindByIDUnrestricted(id uint) (*models.OcservUser, error) {
	var user models.OcservUser
	err := r.db.Where("id = ?", id).First(&user).Error
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

func (r *ocservUserRepository) UpdateUnrestricted(user *models.OcservUser) error {
	return r.db.Save(user).Error
}

func (r *ocservUserRepository) Create(user *models.OcservUser) error {
	return r.db.Create(user).Error
}

func (r *ocservUserRepository) CreateUnrestricted(user *models.OcservUser) (*models.OcservUser, error) {
	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
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

func (r *ocservUserRepository) UpdateUsersByDeleteGroup(ownerAdminID uint, oldGroupName string) (int64, error) {
	query := r.db.Model(&models.OcservUser{}).
		Where("owner_admin_id = ?", ownerAdminID).
		Where("`group` = ?", oldGroupName).
		Update("`group`", "defaults")
	return query.RowsAffected, query.Error
}

func (r *ocservUserRepository) UserSessionLogs(username string, page, limit int, orderBy, sort string, startDate, endDate *time.Time) ([]models.OcservUserSessionLog, int64, error) {
	var logs []models.OcservUserSessionLog
	var total int64

	query := r.db.Model(&models.OcservUserSessionLog{}).Where("username = ?", username)

	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	validOrderFields := map[string]bool{"id": true, "created_at": true}
	validSortOrders := map[string]bool{"asc": true, "desc": true}
	if !validOrderFields[orderBy] {
		orderBy = "id"
	}
	if !validSortOrders[sort] {
		sort = "desc"
	}

	offset := (page - 1) * limit
	err := query.Order(orderBy + " " + sort).Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
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

func (r *ocservUserRepository) RestoreExpired(id uint, expireAt *time.Time) error {
	return r.db.Model(&models.OcservUser{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"expire_at":      expireAt,
			"deactivated_at": nil,
			"is_locked":      false,
		}).Error
}

func (r *ocservUserRepository) CreateCertificate(id uint) error {
	var ocservUser models.OcservUser

	if err := r.db.Where("id = ?", id).First(&ocservUser).Error; err != nil {
		return err
	}

	return r.commonOcservUser.CreateCertificate(ocservUser.Username, ocservUser.Password)
}

func (r *ocservUserRepository) CertificatePath(id uint) (string, string, error) {
	var ocservUser models.OcservUser

	if err := r.db.Where("id = ?", id).First(&ocservUser).Error; err != nil {
		return "", "", err
	}

	path, err := r.commonOcservUser.CertificatePath(ocservUser.Username)
	if err != nil {
		return "", "", err
	}

	return ocservUser.Username, path, nil
}
