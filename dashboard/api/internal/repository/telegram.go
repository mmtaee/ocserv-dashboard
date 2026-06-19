package repository

import (
	"errors"
	"fmt"
	"os"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"gorm.io/gorm"
)

const telegramSettingsSingletonID uint = 1

type TelegramRepository interface {
	// Settings
	Settings() (*models.TelegramSettings, error)
	UpdateSettings(updates map[string]interface{}) (*models.TelegramSettings, error)

	// Accounts
	AccountsForOcservUser(ocservUserID uint) ([]models.TelegramAccount, error)
	DeleteAccount(id uint) error
	PreferredLanguageForChat(chatID int64) (string, error)

	// Packages
	Packages(includeInactive bool) ([]models.TelegramPackage, error)
	PackageByID(id uint) (*models.TelegramPackage, error)
	CreatePackage(pkg *models.TelegramPackage) (*models.TelegramPackage, error)
	UpdatePackage(id uint, updates map[string]interface{}) (*models.TelegramPackage, error)
	DeletePackage(id uint) error

	// Requests
	Requests(page, limit int, orderBy, sort, status, requestType string) ([]models.TelegramRequest, int64, error)
	RequestByID(id uint) (*models.TelegramRequest, error)
	UpdateRequestStatus(id uint, status string, adminNote *string) (*models.TelegramRequest, error)
	SetAwaitingPaymentMessageID(requestID uint, messageID int64) error
	ClearAwaitingPaymentMessageID(requestID uint) error
	DeleteRequest(id uint) error
	MarkDelivered(id uint, ocservUserID *uint) error
}

type telegramRepository struct {
	db *gorm.DB
}

func NewTelegramRepository(db *gorm.DB) TelegramRepository {
	return &telegramRepository{db: db}
}

// ==========================
// Settings
// ==========================

func (r *telegramRepository) Settings() (*models.TelegramSettings, error) {
	var s models.TelegramSettings
	err := r.db.Where("id = ?", telegramSettingsSingletonID).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s = models.TelegramSettings{
				ID:                  telegramSettingsSingletonID,
				DefaultLanguage:     models.TelegramLanguageEN,
				LowQuotaThresholdMB: 200,
			}
			if createErr := r.db.Create(&s).Error; createErr != nil {
				return nil, createErr
			}
			return &s, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *telegramRepository) UpdateSettings(updates map[string]interface{}) (*models.TelegramSettings, error) {
	if _, err := r.Settings(); err != nil {
		return nil, err
	}

	if err := r.db.
		Model(&models.TelegramSettings{}).
		Where("id = ?", telegramSettingsSingletonID).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.Settings()
}

// ==========================
// Accounts
// ==========================

func (r *telegramRepository) AccountsForOcservUser(ocservUserID uint) ([]models.TelegramAccount, error) {
	var accounts []models.TelegramAccount
	err := r.db.
		Where("ocserv_user_id = ?", ocservUserID).
		Order("created_at DESC").
		Find(&accounts).Error
	return accounts, err
}

func (r *telegramRepository) DeleteAccount(id uint) error {
	return r.db.
		Where("id = ?", id).
		Delete(&models.TelegramAccount{}).Error
}

func (r *telegramRepository) PreferredLanguageForChat(chatID int64) (string, error) {
	var acc models.TelegramAccount
	err := r.db.
		Where("chat_id = ?", chatID).
		Order("id ASC").
		First(&acc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return acc.Language, nil
}

// ==========================
// Packages
// ==========================

func (r *telegramRepository) Packages(includeInactive bool) ([]models.TelegramPackage, error) {
	var packages []models.TelegramPackage
	q := r.db.Order("id ASC")
	if !includeInactive {
		q = q.Where("is_active = ?", true)
	}
	err := q.Find(&packages).Error
	return packages, err
}

func (r *telegramRepository) PackageByID(id uint) (*models.TelegramPackage, error) {
	var pkg models.TelegramPackage
	err := r.db.Where("id = ?", id).First(&pkg).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *telegramRepository) CreatePackage(pkg *models.TelegramPackage) (*models.TelegramPackage, error) {
	if err := r.db.Create(pkg).Error; err != nil {
		return nil, err
	}
	return pkg, nil
}

func (r *telegramRepository) UpdatePackage(id uint, updates map[string]interface{}) (*models.TelegramPackage, error) {
	if err := r.db.
		Model(&models.TelegramPackage{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.PackageByID(id)
}

func (r *telegramRepository) DeletePackage(id uint) error {
	return r.db.
		Where("id = ?", id).
		Delete(&models.TelegramPackage{}).Error
}

// ==========================
// Requests
// ==========================

func (r *telegramRepository) Requests(
	page, limit int,
	orderBy, sort, status, requestType string,
) ([]models.TelegramRequest, int64, error) {
	applyFilters := func(q *gorm.DB) *gorm.DB {
		if status != "" {
			q = q.Where("status = ?", status)
		}
		if requestType != "" {
			q = q.Where("type = ?", requestType)
		}
		return q
	}

	var total int64
	countQuery := applyFilters(r.db.Model(&models.TelegramRequest{}))
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	allowedOrder := map[string]bool{
		"created_at": true, "id": true, "status": true, "type": true, "updated_at": true,
	}
	allowedSort := map[string]bool{"asc": true, "desc": true}
	if !allowedOrder[orderBy] {
		orderBy = "created_at"
	}
	if !allowedSort[sort] {
		sort = "desc"
	}

	var requests []models.TelegramRequest
	offset := (page - 1) * limit
	findQuery := applyFilters(r.db)
	err := findQuery.Order(orderBy + " " + sort).Offset(offset).Limit(limit).Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}
	return requests, total, nil
}

func (r *telegramRepository) RequestByID(id uint) (*models.TelegramRequest, error) {
	var req models.TelegramRequest
	err := r.db.Where("id = ?", id).First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *telegramRepository) UpdateRequestStatus(id uint, status string, adminNote *string) (*models.TelegramRequest, error) {
	updates := map[string]interface{}{"status": status}
	if adminNote != nil {
		updates["admin_note"] = *adminNote
	}
	if err := r.db.
		Model(&models.TelegramRequest{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.RequestByID(id)
}

func (r *telegramRepository) SetAwaitingPaymentMessageID(requestID uint, messageID int64) error {
	return r.db.
		Model(&models.TelegramRequest{}).
		Where("id = ?", requestID).
		Update("awaiting_payment_message_id", messageID).Error
}

func (r *telegramRepository) ClearAwaitingPaymentMessageID(requestID uint) error {
	return r.db.
		Model(&models.TelegramRequest{}).
		Where("id = ?", requestID).
		Updates(map[string]interface{}{"awaiting_payment_message_id": nil}).Error
}

// DeleteRequest removes a finished request row. Active pipeline statuses cannot be deleted.
func (r *telegramRepository) DeleteRequest(id uint) error {
	var req models.TelegramRequest
	if err := r.db.Where("id = ?", id).First(&req).Error; err != nil {
		return err
	}
	switch req.Status {
	case models.TelegramRequestStatusPending,
		models.TelegramRequestStatusAwaitingPayment,
		models.TelegramRequestStatusPaymentUploaded:
		return fmt.Errorf("cannot delete an active request (status=%s)", req.Status)
	}
	if req.ReceiptFilePath != "" {
		_ = os.Remove(req.ReceiptFilePath)
	}
	return r.db.Where("id = ?", id).Delete(&models.TelegramRequest{}).Error
}

func (r *telegramRepository) MarkDelivered(id uint, ocservUserID *uint) error {
	updates := map[string]interface{}{
		"status":       models.TelegramRequestStatusDelivered,
		"delivered_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}
	if ocservUserID != nil {
		updates["target_ocserv_id"] = *ocservUserID
	}
	return r.db.
		Model(&models.TelegramRequest{}).
		Where("id = ?", id).
		Updates(updates).Error
}
