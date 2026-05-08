package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/database"
	"gorm.io/gorm"
)

const telegramSettingsSingletonID uint = 1

type TelegramRepository struct {
	db *gorm.DB
}

type TelegramSettingsRepo interface {
	Settings(ctx context.Context) (*models.TelegramSettings, error)
	UpdateSettings(ctx context.Context, updates map[string]interface{}) (*models.TelegramSettings, error)
}

type TelegramAccountRepo interface {
	AccountsForOcservUser(ctx context.Context, ocservUserID uint) ([]models.TelegramAccount, error)
	DeleteAccount(ctx context.Context, id uint) error
}

type TelegramPackageRepo interface {
	Packages(ctx context.Context, includeInactive bool) ([]models.TelegramPackage, error)
	PackageByID(ctx context.Context, id uint) (*models.TelegramPackage, error)
	CreatePackage(ctx context.Context, pkg *models.TelegramPackage) (*models.TelegramPackage, error)
	UpdatePackage(ctx context.Context, id uint, updates map[string]interface{}) (*models.TelegramPackage, error)
	DeletePackage(ctx context.Context, id uint) error
}

type TelegramRequestRepo interface {
	Requests(ctx context.Context, pagination *request.Pagination, status, requestType string) ([]models.TelegramRequest, int64, error)
	RequestByID(ctx context.Context, id uint) (*models.TelegramRequest, error)
	UpdateRequestStatus(ctx context.Context, id uint, status string, adminNote *string) (*models.TelegramRequest, error)
	MarkDelivered(ctx context.Context, id uint, ocservUserID *uint) error
}

type TelegramRepositoryInterface interface {
	TelegramSettingsRepo
	TelegramAccountRepo
	TelegramPackageRepo
	TelegramRequestRepo
}

func NewTelegramRepository() *TelegramRepository {
	return &TelegramRepository{db: database.GetConnection()}
}

// ==========================
// Settings
// ==========================

func (r *TelegramRepository) Settings(ctx context.Context) (*models.TelegramSettings, error) {
	var s models.TelegramSettings
	err := r.db.WithContext(ctx).Where("id = ?", telegramSettingsSingletonID).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s = models.TelegramSettings{
				ID:                  telegramSettingsSingletonID,
				DefaultLanguage:     models.TelegramLanguageEN,
				LowQuotaThresholdMB: 200,
			}
			if createErr := r.db.WithContext(ctx).Create(&s).Error; createErr != nil {
				return nil, createErr
			}
			return &s, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *TelegramRepository) UpdateSettings(ctx context.Context, updates map[string]interface{}) (*models.TelegramSettings, error) {
	if _, err := r.Settings(ctx); err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Model(&models.TelegramSettings{}).
		Where("id = ?", telegramSettingsSingletonID).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.Settings(ctx)
}

// ==========================
// Accounts
// ==========================

func (r *TelegramRepository) AccountsForOcservUser(ctx context.Context, ocservUserID uint) ([]models.TelegramAccount, error) {
	var accounts []models.TelegramAccount
	err := r.db.WithContext(ctx).
		Where("ocserv_user_id = ?", ocservUserID).
		Order("created_at DESC").
		Find(&accounts).Error
	return accounts, err
}

func (r *TelegramRepository) DeleteAccount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.TelegramAccount{}).Error
}

// ==========================
// Packages
// ==========================

func (r *TelegramRepository) Packages(ctx context.Context, includeInactive bool) ([]models.TelegramPackage, error) {
	var packages []models.TelegramPackage
	q := r.db.WithContext(ctx).Order("id ASC")
	if !includeInactive {
		q = q.Where("is_active = ?", true)
	}
	err := q.Find(&packages).Error
	return packages, err
}

func (r *TelegramRepository) PackageByID(ctx context.Context, id uint) (*models.TelegramPackage, error) {
	var pkg models.TelegramPackage
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&pkg).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *TelegramRepository) CreatePackage(ctx context.Context, pkg *models.TelegramPackage) (*models.TelegramPackage, error) {
	if err := r.db.WithContext(ctx).Create(pkg).Error; err != nil {
		return nil, err
	}
	return pkg, nil
}

func (r *TelegramRepository) UpdatePackage(ctx context.Context, id uint, updates map[string]interface{}) (*models.TelegramPackage, error) {
	if err := r.db.WithContext(ctx).
		Model(&models.TelegramPackage{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.PackageByID(ctx, id)
}

func (r *TelegramRepository) DeletePackage(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.TelegramPackage{}).Error
}

// ==========================
// Requests
// ==========================

func (r *TelegramRepository) Requests(
	ctx context.Context,
	pagination *request.Pagination,
	status, requestType string,
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
	if err := applyFilters(r.db.WithContext(ctx).Model(&models.TelegramRequest{})).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var requests []models.TelegramRequest
	tx := request.Paginator(ctx, r.db, pagination)
	tx = applyFilters(tx).Order("created_at DESC")
	if err := tx.Find(&requests).Error; err != nil {
		return nil, 0, err
	}
	return requests, total, nil
}

func (r *TelegramRepository) RequestByID(ctx context.Context, id uint) (*models.TelegramRequest, error) {
	var req models.TelegramRequest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *TelegramRepository) UpdateRequestStatus(ctx context.Context, id uint, status string, adminNote *string) (*models.TelegramRequest, error) {
	updates := map[string]interface{}{"status": status}
	if adminNote != nil {
		updates["admin_note"] = *adminNote
	}
	if err := r.db.WithContext(ctx).
		Model(&models.TelegramRequest{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.RequestByID(ctx, id)
}

func (r *TelegramRepository) MarkDelivered(ctx context.Context, id uint, ocservUserID *uint) error {
	updates := map[string]interface{}{
		"status":       models.TelegramRequestStatusDelivered,
		"delivered_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}
	if ocservUserID != nil {
		updates["target_ocserv_id"] = *ocservUserID
	}
	return r.db.WithContext(ctx).
		Model(&models.TelegramRequest{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// FormatTelegramAccountSummary returns a readable string summarizing how many
// links a chat has, useful for admin notifications.
func FormatTelegramAccountSummary(accounts []models.TelegramAccount) string {
	if len(accounts) == 0 {
		return "no linked accounts"
	}
	return fmt.Sprintf("%d linked account(s)", len(accounts))
}
