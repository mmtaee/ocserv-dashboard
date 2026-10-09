package repository

import (
	"context"
	"strings"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TopBandwidthUsers struct {
	TopRX []models.OcservUser `json:"top_rx"`
	TopTX []models.OcservUser `json:"top_tx"`
}

type TotalBandwidths struct {
	RX float64 `json:"rx" validate:"required"`
	TX float64 `json:"tx" validate:"required"`
}

type OcservUserRepository struct {
	db *gorm.DB
}

// OcservUserExpiryWindow limits users by their calculated effective expiry.
type OcservUserExpiryWindow struct {
	StartsAt time.Time
	EndsAt   time.Time
}

type OcservUserCRUD interface {
	Users(ctx context.Context, pagination *request.Pagination, ownerID uint, q string, filters string, group string, expiryWindow *OcservUserExpiryWindow) ([]models.OcservUser, int64, error)
	UsersByUsername(ctx context.Context, pagination *request.Pagination, ownerID uint, usernames []string, q string, group string, expiryWindow *OcservUserExpiryWindow) ([]models.OcservUser, int64, error)
	Create(ctx context.Context, user *models.OcservUser) (*models.OcservUser, error)
	GetByID(ctx context.Context, id uint) (*models.OcservUser, error)
	GetByUsername(ctx context.Context, username string) (*models.OcservUser, error)
	GetByIDsForUpdate(ctx context.Context, ids []uint) ([]models.OcservUser, error)
	Update(ctx context.Context, ocservUser *models.OcservUser) (*models.OcservUser, error)
	Delete(ctx context.Context, id uint) (*models.OcservUser, error)
}

type OcservUserBulkTx interface {
	GetByIDsForUpdate(ctx context.Context, ids []uint) ([]models.OcservUser, error)
	Update(ctx context.Context, user *models.OcservUser) (*models.OcservUser, error)
	Delete(ctx context.Context, id uint) (*models.OcservUser, error)
	Lock(ctx context.Context, id uint) error
	UnLock(ctx context.Context, id uint) error
}

type OcservUserBulkRepository interface {
	WithTransaction(ctx context.Context, operation func(OcservUserBulkTx) error) error
}

type OcservUserStats interface {
	UserStatistics(ctx context.Context, id uint, dateStart, dateEnd *time.Time) ([]models.DailyTraffic, error)

	TotalBandwidthUserDateRange(ctx context.Context, id uint, dateStart, dateEnd *time.Time) (TotalBandwidths, error)
	UserSessionLogs(ctx context.Context, pagination *request.Pagination, username string, dateStart, dateEnd *time.Time) (*[]models.OcservUserSessionLog, int64, error)
}

type OcservUserPassword interface {
	ExistingUsernames(ctx context.Context, usernames []string) ([]string, error)
	OcpasswdSyncToDB(ctx context.Context, users []models.OcservUser) ([]models.OcservUser, error)
}

type OcservUserGroup interface {
	UpdateUsersByDeleteGroup(ctx context.Context, groupName string) ([]models.OcservUser, error)
}

type OcservUserActions interface {
	Lock(ctx context.Context, id uint) error
	UnLock(ctx context.Context, id uint) error
	ResetUsage(ctx context.Context, user *models.OcservUser) (*models.OcservUser, error)
	RestoreExpired(ctx context.Context, user *models.OcservUser) error
}

type OcservUserRepositoryInterface interface {
	OcservUserCRUD
	OcservUserStats
	OcservUserPassword
	OcservUserGroup
	OcservUserActions
}

func NewtOcservUserRepository() *OcservUserRepository {
	return &OcservUserRepository{db: database.GetConnection()}
}

func (o *OcservUserRepository) WithTransaction(ctx context.Context, operation func(OcservUserBulkTx) error) error {
	return o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return operation(&OcservUserRepository{db: tx})
	})
}

func (o *OcservUserRepository) Users(
	ctx context.Context,
	pagination *request.Pagination,
	ownerID uint,
	q string,
	filter string,
	group string,
	expiryWindow *OcservUserExpiryWindow,
) (
	[]models.OcservUser, int64, error,
) {
	var totalRecords int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if ownerID != 0 {
			db = db.Where("owner_id = ?", ownerID)
		}
		if len(q) >= 2 {
			db = db.Where("LOWER(username) LIKE ?", "%"+strings.ToLower(q)+"%")
		}
		if group != "" {
			db = db.Where(`"group" = ?`, group)
		}
		if expiryWindow != nil {
			db = db.Where("expire_at >= ? AND expire_at <= ?", expiryWindow.StartsAt, expiryWindow.EndsAt)
		}

		switch filter {
		case "active":
			db = db.Where("deactivated_at IS NULL AND is_locked = false")
		case "deactivated":
			db = db.Where("deactivated_at IS NOT NULL")
		case "locked":
			db = db.Where("is_locked = true")
		default:
		}

		return db
	}

	totalQuery := applyFilters(o.db.WithContext(ctx).Model(&models.OcservUser{}))
	if err := totalQuery.Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	var ocservUser []models.OcservUser
	txPaginator := request.Paginator(ctx, o.db, pagination)

	query := applyFilters(txPaginator.Model(&ocservUser)).Preload("Owner")
	if err := query.Find(&ocservUser).Error; err != nil {
		return nil, 0, err
	}

	return ocservUser, totalRecords, nil
}

func (o *OcservUserRepository) UsersByUsername(
	ctx context.Context,
	pagination *request.Pagination,
	ownerID uint,
	usernames []string,
	q string,
	group string,
	expiryWindow *OcservUserExpiryWindow,
) ([]models.OcservUser, int64, error) {
	applyFilters := func(db *gorm.DB) *gorm.DB {
		if ownerID != 0 {
			db = db.Where("owner_id = ?", ownerID)
		}

		if len(q) >= 2 {
			db = db.Where("LOWER(username) LIKE ?", "%"+strings.ToLower(q)+"%")
		}

		if group != "" {
			db = db.Where(`"group" = ?`, group)
		}
		if expiryWindow != nil {
			db = db.Where("expire_at >= ? AND expire_at <= ?", expiryWindow.StartsAt, expiryWindow.EndsAt)
		}

		return db
	}

	base := o.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Where("username IN ?", usernames)

	countDB := applyFilters(base)

	var totalRecords int64
	if err := countDB.Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	if totalRecords == 0 {
		return []models.OcservUser{}, 0, nil
	}

	queryDB := applyFilters(base)

	var ocservUser []models.OcservUser

	queryDB = request.Paginator(ctx, queryDB, pagination)

	if err := queryDB.Preload("Owner").Find(&ocservUser).Error; err != nil {
		return nil, 0, err
	}

	return ocservUser, totalRecords, nil
}

func (o *OcservUserRepository) Create(ctx context.Context, ocservUser *models.OcservUser) (*models.OcservUser, error) {
	return ocservUser, o.db.WithContext(ctx).Omit("Owner").Create(ocservUser).Error
}

func (o *OcservUserRepository) GetByID(ctx context.Context, id uint) (*models.OcservUser, error) {
	var ocservUser models.OcservUser
	err := o.db.WithContext(ctx).Preload("Owner").Where("id = ?", id).First(&ocservUser).Error
	if err != nil {
		return nil, err
	}
	return &ocservUser, nil
}

func (o *OcservUserRepository) GetByUsername(ctx context.Context, username string) (*models.OcservUser, error) {
	var ocservUser models.OcservUser
	err := o.db.WithContext(ctx).Preload("Owner").Where("username = ?", username).First(&ocservUser).Error
	if err != nil {
		return nil, err
	}

	return &ocservUser, nil
}

func (o *OcservUserRepository) GetByIDsForUpdate(ctx context.Context, ids []uint) ([]models.OcservUser, error) {
	var users []models.OcservUser
	err := o.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Owner").
		Where("id IN ?", ids).
		Order("id ASC").
		Find(&users).Error
	return users, err
}

func (o *OcservUserRepository) Update(ctx context.Context, ocservUser *models.OcservUser) (*models.OcservUser, error) {
	return ocservUser, o.db.WithContext(ctx).Omit("Owner").Save(ocservUser).Error
}

func (o *OcservUserRepository) Lock(ctx context.Context, id uint) error {
	var ocservUser models.OcservUser
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).First(&ocservUser).Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.OcservUser{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{"is_locked": true}).Error; err != nil {
			return err
		}

		return nil
	})
	return err
}

func (o *OcservUserRepository) UnLock(ctx context.Context, id uint) error {
	var ocservUser models.OcservUser
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).First(&ocservUser).Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.OcservUser{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{"is_locked": false}).Error; err != nil {
			return err
		}

		return nil
	})
	return err
}

func (o *OcservUserRepository) ResetUsage(ctx context.Context, user *models.OcservUser) (*models.OcservUser, error) {
	result := o.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"running_rx":     user.RunningRx,
			"running_tx":     user.RunningTx,
			"usage_reset_at": user.UsageResetAt,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (o *OcservUserRepository) Delete(ctx context.Context, id uint) (*models.OcservUser, error) {
	var ocservUser models.OcservUser
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).First(&ocservUser).Error; err != nil {
			return err
		}
		if err := tx.Delete(&ocservUser).Error; err != nil {
			return err
		}
		return nil
	})
	return &ocservUser, err
}

func (o *OcservUserRepository) UpdateUsersByDeleteGroup(ctx context.Context, groupName string) ([]models.OcservUser, error) {
	var users []models.OcservUser

	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`group` = ?", groupName).Find(&users).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.OcservUser{}).
			Where("`group` = ?", groupName).
			Update("group", "defaults").Error; err != nil {
			return err
		}

		return nil
	})

	return users, err
}

func (o *OcservUserRepository) UserStatistics(ctx context.Context, id uint, dateStart, dateEnd *time.Time) ([]models.DailyTraffic, error) {
	var results []models.DailyTraffic

	query := o.db.WithContext(ctx).
		Model(&models.OcservUserTrafficStatistics{}).
		Where("ocserv_user_id = ?", id).
		Select(`
		DATE(ocserv_user_traffic_statistics.created_at) AS date,
		SUM(ocserv_user_traffic_statistics.rx) / 1073741824.0 AS rx,
		SUM(ocserv_user_traffic_statistics.tx) / 1073741824.0 AS tx
	`)

	if dateStart != nil {
		query = query.Where("ocserv_user_traffic_statistics.created_at >= ?", *dateStart)
	}
	if dateEnd != nil {
		query = query.Where("ocserv_user_traffic_statistics.created_at <= ?", *dateEnd)
	}

	err := query.
		Group("DATE(ocserv_user_traffic_statistics.created_at)").
		Order("DATE(ocserv_user_traffic_statistics.created_at)").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}
	return results, nil
}

func (o *OcservUserRepository) TotalBandwidthUserDateRange(ctx context.Context, id uint, dateStart, dateEnd *time.Time) (TotalBandwidths, error) {
	var total TotalBandwidths

	query := o.db.WithContext(ctx).
		Model(&models.OcservUserTrafficStatistics{}).
		Where("ocserv_user_id = ?", id).
		Select(`
			COALESCE(SUM(rx),0) / 1073741824.0 AS rx,
			COALESCE(SUM(tx),0) / 1073741824.0 AS tx`)

	// Apply filters based on dateStart and dateEnd
	if dateStart != nil {
		query = query.Where("created_at >= ?", *dateStart)
	}
	if dateEnd != nil {
		query = query.Where("created_at <= ?", *dateEnd)
	}

	err := query.Scan(&total).Error
	if err != nil {
		return total, err
	}
	return total, nil
}

func (o *OcservUserRepository) ExistingUsernames(ctx context.Context, usernames []string) ([]string, error) {
	var existing []string
	err := o.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Where("username IN ?", usernames).
		Pluck("username", &existing).Error
	return existing, err
}

func (o *OcservUserRepository) OcpasswdSyncToDB(ctx context.Context, users []models.OcservUser) ([]models.OcservUser, error) {
	return users, o.db.WithContext(ctx).Omit("Owner").Create(&users).Error
}

func (o *OcservUserRepository) RestoreExpired(ctx context.Context, user *models.OcservUser) error {
	return o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var u models.OcservUser
		if err := tx.
			Where("id = ?", user.ID).
			First(&u).Error; err != nil {
			return err
		}

		now := time.Now()

		if err := tx.
			Model(&u).
			Updates(map[string]interface{}{
				"expiry_mode":                        user.ExpiryMode,
				"expire_at":                          user.ExpireAt,
				"expire_days_after_first_connection": user.ExpireDaysAfterFirstConnection,
				"first_connected_at":                 user.FirstConnectedAt,
				"deactivated_at":                     nil,
				"usage_reset_at":                     &now,
				"is_locked":                          false,
				"running_rx":                         0,
				"running_tx":                         0,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (o *OcservUserRepository) UserSessionLogs(
	ctx context.Context,
	pagination *request.Pagination,
	username string,
	dateStart, dateEnd *time.Time,
) (*[]models.OcservUserSessionLog, int64, error) {
	var totalRecords int64

	query := o.db.WithContext(ctx).
		Model(&models.OcservUserSessionLog{}).
		Where("username = ?", username)

	if dateStart != nil {
		query = query.Where("created_at >= ?", *dateStart)
	}

	if dateEnd != nil {
		query = query.Where("created_at < ?", dateEnd.AddDate(0, 0, 1))
	}

	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	var logs []models.OcservUserSessionLog
	if err := request.Paginator(ctx, query, pagination).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return &logs, totalRecords, nil
}
