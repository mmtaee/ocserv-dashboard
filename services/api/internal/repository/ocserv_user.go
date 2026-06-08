package repository

import (
	"context"
	"fmt"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	"github.com/mmtaee/ocserv-dashboard/common/ocserv/occtl"
	"github.com/mmtaee/ocserv-dashboard/common/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/database"
	"gorm.io/gorm"
	"strings"
	"time"
)

type TopBandwidthUsers struct {
	TopRX []models.OcservUser `json:"top_rx"`
	TopTX []models.OcservUser `json:"top_tx"`
}

type TotalBandwidths struct {
	RX float64 `json:"rx" validate:"required"`
	TX float64 `json:"tx" validate:"required"`
}

type OcpasswdUser struct {
	Username string `json:"username" validate:"required"`
	Group    string `json:"group" validate:"required"`
}

type OcservUserRepository struct {
	db                    *gorm.DB
	commonOcservUserRepo  user.OcservUserInterface
	commonOcservOcctlRepo occtl.OcservOcctlInterface
}

type OcservUserCRUD interface {
	Users(ctx context.Context, pagination *request.Pagination, owner string, q string, filters string, group string) ([]models.OcservUser, int64, error)
	UsersByUsername(ctx context.Context, pagination *request.Pagination, owner string, usernames []string, q string, group string) ([]models.OcservUser, int64, error)
	Create(ctx context.Context, user *models.OcservUser) (*models.OcservUser, error)
	GetByUID(ctx context.Context, uid string) (*models.OcservUser, error)
	GetByUsername(ctx context.Context, username string) (*models.OcservUser, error)
	Update(ctx context.Context, ocservUser *models.OcservUser) (*models.OcservUser, error)
	Delete(ctx context.Context, uid string) (string, error)
}

type OcservUserStats interface {
	UserStatistics(ctx context.Context, uid string, dateStart, dateEnd *time.Time) ([]models.DailyTraffic, error)

	TotalBandwidthUserDateRange(ctx context.Context, uid string, dateStart, dateEnd *time.Time) (TotalBandwidths, error)
	UserSessionLogs(ctx context.Context, pagination *request.Pagination, username string, dateStart, dateEnd *time.Time) (*[]models.OcservUserSessionLog, int64, error)
}

type OcservUserPassword interface {
	Ocpasswd(ctx context.Context, pagination *request.Pagination) ([]user.Ocpasswd, int, error)
	OcpasswdSyncToDB(ctx context.Context, users []models.OcservUser) ([]models.OcservUser, error)
}

type OcservUserGroup interface {
	UpdateUsersByDeleteGroup(ctx context.Context, groupName string) ([]models.OcservUser, error)
}

type OcservUserActions interface {
	Lock(ctx context.Context, uid string) error
	UnLock(ctx context.Context, uid string) error
	RestoreExpired(ctx context.Context, uid string, expireAt *time.Time) error
	CreateCertificate(ctx context.Context, uid string) error
	CertificatePath(ctx context.Context, uid string) (string, string, error)
	CertificatePathByUsername(ctx context.Context, username string) (string, error)
}

type OcservUserRepositoryInterface interface {
	OcservUserCRUD
	OcservUserStats
	OcservUserPassword
	OcservUserGroup
	OcservUserActions
}

func NewtOcservUserRepository() *OcservUserRepository {
	return &OcservUserRepository{
		db:                    database.GetConnection(),
		commonOcservUserRepo:  user.NewOcservUser(),
		commonOcservOcctlRepo: occtl.NewOcservOcctl(),
	}
}

func (o *OcservUserRepository) applyCertificateStatus(ocservUser *models.OcservUser) {
	status := o.commonOcservUserRepo.CertificateStatus(ocservUser.Username)
	ocservUser.CertificateEnabled = status.Enabled
	ocservUser.CertificateAvailable = status.Available
}

func (o *OcservUserRepository) Users(
	ctx context.Context,
	pagination *request.Pagination,
	owner string,
	q string,
	filter string,
	group string,
) (
	[]models.OcservUser, int64, error,
) {
	var totalRecords int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if owner != "" {
			db = db.Where("owner = ?", owner)
		}
		if len(q) >= 2 {
			db = db.Where("LOWER(username) LIKE ?", "%"+strings.ToLower(q)+"%")
		}
		if group != "" {
			db = db.Where(`"group" = ?`, group)
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

	query := applyFilters(txPaginator.Model(&ocservUser))
	if err := query.Find(&ocservUser).Error; err != nil {
		return nil, 0, err
	}

	for i := range ocservUser {
		o.applyCertificateStatus(&ocservUser[i])
	}

	return ocservUser, totalRecords, nil
}

func (o *OcservUserRepository) UsersByUsername(
	ctx context.Context,
	pagination *request.Pagination,
	owner string,
	usernames []string,
	q string,
	group string,
) ([]models.OcservUser, int64, error) {
	applyFilters := func(db *gorm.DB) *gorm.DB {
		if owner != "" {
			db = db.Where("owner = ?", owner)
		}

		if len(q) >= 2 {
			db = db.Where("LOWER(username) LIKE ?", "%"+strings.ToLower(q)+"%")
		}

		if group != "" {
			db = db.Where(`"group" = ?`, group)
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

	if err := queryDB.Find(&ocservUser).Error; err != nil {
		return nil, 0, err
	}

	return ocservUser, totalRecords, nil
}

func (o *OcservUserRepository) Create(ctx context.Context, ocservUser *models.OcservUser) (*models.OcservUser, error) {
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ocservUser).Error; err != nil {
			return err
		}
		if err := o.commonOcservUserRepo.Create(ocservUser.Group, ocservUser.Username, ocservUser.Password, ocservUser.Config); err != nil {
			return err
		}

		if err := o.commonOcservUserRepo.CreateCertificate(ocservUser.Username, ocservUser.Password); err != nil {
			_, _ = o.commonOcservUserRepo.Delete(ocservUser.Username)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	go func() {
		_, _ = o.commonOcservOcctlRepo.ReloadConfigs()
	}()
	o.applyCertificateStatus(ocservUser)
	return ocservUser, err
}

func (o *OcservUserRepository) GetByUID(ctx context.Context, uid string) (*models.OcservUser, error) {
	var ocservUser models.OcservUser
	err := o.db.WithContext(ctx).Where("uid = ?", uid).First(&ocservUser).Error
	if err != nil {
		return nil, err
	}
	o.applyCertificateStatus(&ocservUser)
	return &ocservUser, nil
}

func (o *OcservUserRepository) GetByUsername(ctx context.Context, username string) (*models.OcservUser, error) {
	var ocservUser models.OcservUser
	err := o.db.WithContext(ctx).Where("username = ?", username).First(&ocservUser).Error
	if err != nil {
		return nil, err
	}

	o.applyCertificateStatus(&ocservUser)

	return &ocservUser, nil
}

func (o *OcservUserRepository) Update(ctx context.Context, ocservUser *models.OcservUser) (*models.OcservUser, error) {
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&ocservUser).Error; err != nil {
			return err
		}
		if err := o.commonOcservUserRepo.Create(ocservUser.Group, ocservUser.Username, ocservUser.Password, ocservUser.Config); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	go func() {
		_, _ = o.commonOcservOcctlRepo.ReloadConfigs()
	}()
	return ocservUser, nil
}

func (o *OcservUserRepository) Lock(ctx context.Context, uid string) error {
	var ocservUser models.OcservUser
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("uid = ?", uid).First(&ocservUser).Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.OcservUser{}).
			Where("uid = ?", uid).
			Updates(map[string]interface{}{"is_locked": true}).Error; err != nil {
			return err
		}

		if _, err := o.commonOcservUserRepo.Lock(ocservUser.Username); err != nil {
			return err
		}
		return nil
	})
	return err
}

func (o *OcservUserRepository) UnLock(ctx context.Context, uid string) error {
	var ocservUser models.OcservUser
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("uid = ?", uid).First(&ocservUser).Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.OcservUser{}).
			Where("uid = ?", uid).
			Updates(map[string]interface{}{"is_locked": false}).Error; err != nil {
			return err
		}

		if _, err := o.commonOcservUserRepo.UnLock(ocservUser.Username); err != nil {
			return err
		}
		return nil
	})
	return err
}

func (o *OcservUserRepository) Delete(ctx context.Context, uid string) (string, error) {
	var ocservUser models.OcservUser
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("uid = ?", uid).First(&ocservUser).Error; err != nil {
			return err
		}
		if err := tx.Delete(&ocservUser).Error; err != nil {
			return err
		}
		if _, err := o.commonOcservUserRepo.Delete(ocservUser.Username); err != nil {
			return err
		}
		return nil
	})

	go func() {
		_, _ = o.commonOcservOcctlRepo.ReloadConfigs()
	}()

	return ocservUser.Username, err
}

func (o *OcservUserRepository) UpdateUsersByDeleteGroup(ctx context.Context, groupName string) ([]models.OcservUser, error) {
	var users []models.OcservUser

	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`group` = ?", groupName).Select("id", "group", "username").Find(&users).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.OcservUser{}).
			Where("`group` = ?", groupName).
			Update("group", "defaults").Error; err != nil {
			return err
		}

		return nil
	})

	go func() {
		_, _ = o.commonOcservOcctlRepo.ReloadConfigs()
	}()

	return users, err
}

func (o *OcservUserRepository) UserStatistics(ctx context.Context, uid string, dateStart, dateEnd *time.Time) ([]models.DailyTraffic, error) {
	var results []models.DailyTraffic

	query := o.db.WithContext(ctx).
		Model(&models.OcservUserTrafficStatistics{}).
		Joins("JOIN ocserv_users ou ON ou.id = ocserv_user_traffic_statistics.oc_user_id").
		Where("ou.uid = ?", uid).
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

func (o *OcservUserRepository) TotalBandwidthUserDateRange(ctx context.Context, uid string, dateStart, dateEnd *time.Time) (TotalBandwidths, error) {
	var total TotalBandwidths

	query := o.db.WithContext(ctx).
		Model(&models.OcservUserTrafficStatistics{}).
		Where("oc_user_id = ? ", uid).
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

func (o *OcservUserRepository) Ocpasswd(ctx context.Context, pagination *request.Pagination) ([]user.Ocpasswd, int, error) {
	users, _, err := o.commonOcservUserRepo.Ocpasswd(ctx)
	if err != nil {
		return nil, 0, err
	}
	if len(*users) == 0 {
		return []user.Ocpasswd{}, 0, nil
	}

	usernames := make([]string, len(*users))
	for i, u := range *users {
		usernames[i] = u.Username
	}

	var existing []string
	if err = o.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Where("username IN ?", usernames).
		Pluck("username", &existing).Error; err != nil {
		return nil, 0, err
	}

	existingSet := make(map[string]struct{}, len(existing))
	for _, u := range existing {
		existingSet[u] = struct{}{}
	}

	newUsers := make([]user.Ocpasswd, 0)
	for _, u := range *users {
		if _, exists := existingSet[u.Username]; !exists {
			newUsers = append(newUsers, user.Ocpasswd{
				Username: u.Username,
				Group:    u.Group,
			})
		}
	}

	totalNew := len(newUsers)
	if totalNew == 0 {
		return []user.Ocpasswd{}, 0, nil
	}

	start := (pagination.Page - 1) * pagination.PageSize
	if start >= totalNew {
		return []user.Ocpasswd{}, totalNew, nil
	}

	end := start + pagination.PageSize
	if end > totalNew {
		end = totalNew
	}

	paged := newUsers[start:end]

	return paged, totalNew, nil
}

func (o *OcservUserRepository) OcpasswdSyncToDB(ctx context.Context, users []models.OcservUser) ([]models.OcservUser, error) {
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&users).Error; err != nil {
			return err
		}

		//for _, i := range users {
		//	if err := o.commonOcservUserRepo.Create(i.Group, i.Username, i.Password, i.Config); err != nil {
		//		return err
		//	}
		//}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reload configs if any user has a custom config
	for _, u := range users {
		if u.Config != nil {
			go func() {
				_, _ = o.commonOcservOcctlRepo.ReloadConfigs()
			}()
			break
		}
	}

	return users, nil
}

func (o *OcservUserRepository) RestoreExpired(ctx context.Context, uid string, expireAt *time.Time) error {
	return o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var u models.OcservUser
		if err := tx.
			Where("uid = ?", uid).
			First(&u).Error; err != nil {
			return err
		}

		terminateOutput, err := o.commonOcservOcctlRepo.TerminateUser(u.Username)
		if err != nil && !isNoActiveOcctlUserError(terminateOutput, err) {
			return fmt.Errorf("failed to terminate ocserv user %q: %s: %w", u.Username, strings.TrimSpace(terminateOutput), err)
		}

		unlockOutput, err := o.commonOcservUserRepo.UnLock(u.Username)
		if err != nil && !isAlreadyUnlockedOcpasswdError(unlockOutput, err) {
			return fmt.Errorf("failed to unlock ocserv user %q: %s: %w", u.Username, strings.TrimSpace(unlockOutput), err)
		}

		now := time.Now()

		if err := tx.
			Model(&u).
			Updates(map[string]interface{}{
				"expire_at":      expireAt,
				"deactivated_at": nil,
				"usage_reset_at": &now,
				"is_locked":      false,
				"rx":             0,
				"tx":             0,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (o *OcservUserRepository) CreateCertificate(ctx context.Context, uid string) error {
	var ocservUser models.OcservUser

	if err := o.db.WithContext(ctx).
		Where("uid = ?", uid).
		First(&ocservUser).Error; err != nil {
		return err
	}

	return o.commonOcservUserRepo.CreateCertificate(ocservUser.Username, ocservUser.Password)
}

func (o *OcservUserRepository) CertificatePath(ctx context.Context, uid string) (string, string, error) {
	var ocservUser models.OcservUser

	if err := o.db.WithContext(ctx).
		Where("uid = ?", uid).
		First(&ocservUser).Error; err != nil {
		return "", "", err
	}

	path, err := o.commonOcservUserRepo.CertificatePath(ocservUser.Username)
	if err != nil {
		return "", "", err
	}

	return ocservUser.Username, path, nil
}

func (o *OcservUserRepository) CertificatePathByUsername(ctx context.Context, username string) (string, error) {
	var ocservUser models.OcservUser

	if err := o.db.WithContext(ctx).
		Where("username = ?", username).
		First(&ocservUser).Error; err != nil {
		return "", err
	}

	return o.commonOcservUserRepo.CertificatePath(ocservUser.Username)
}

func isAlreadyUnlockedOcpasswdError(output string, err error) bool {
	text := strings.ToLower(strings.TrimSpace(output + " " + err.Error()))

	return strings.Contains(text, "not locked") ||
		strings.Contains(text, "not disabled") ||
		strings.Contains(text, "already unlocked") ||
		strings.Contains(text, "already enabled")
}

func isNoActiveOcctlUserError(output string, err error) bool {
	text := strings.ToLower(strings.TrimSpace(output + " " + err.Error()))

	return strings.Contains(text, "could not terminate user") ||
		strings.Contains(text, "could not disconnect user") ||
		strings.Contains(text, "not found") ||
		strings.Contains(text, "no such user")
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
