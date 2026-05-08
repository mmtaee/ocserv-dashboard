package repository

import (
	"context"
	"errors"
	"time"

	"github.com/mmtaee/ocserv-dashboard/common/models"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/database"
	"gorm.io/gorm"
)

const settingsSingletonID uint = 1

type Repository struct {
	db *gorm.DB
}

func New() *Repository {
	return &Repository{db: database.GetConnection()}
}

// =============================================================================
// Settings
// =============================================================================

func (r *Repository) Settings(ctx context.Context) (*models.TelegramSettings, error) {
	var s models.TelegramSettings
	err := r.db.WithContext(ctx).Where("id = ?", settingsSingletonID).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s = models.TelegramSettings{
				ID:                  settingsSingletonID,
				DefaultLanguage:     models.TelegramLanguageEN,
				LowQuotaThresholdMB: 200,
			}
			if cerr := r.db.WithContext(ctx).Create(&s).Error; cerr != nil {
				return nil, cerr
			}
			return &s, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *Repository) SetBotUsername(ctx context.Context, username string) error {
	return r.db.WithContext(ctx).
		Model(&models.TelegramSettings{}).
		Where("id = ?", settingsSingletonID).
		Update("bot_username", username).Error
}

// =============================================================================
// Accounts
// =============================================================================

func (r *Repository) AccountsByChatID(ctx context.Context, chatID int64) ([]models.TelegramAccount, error) {
	var accounts []models.TelegramAccount
	err := r.db.WithContext(ctx).Where("chat_id = ?", chatID).Order("created_at DESC").Find(&accounts).Error
	return accounts, err
}

func (r *Repository) AccountByID(ctx context.Context, id uint) (*models.TelegramAccount, error) {
	var account models.TelegramAccount
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *Repository) UpsertAccount(ctx context.Context, chatID int64, telegramUsername, language string, ocservUserID uint) (*models.TelegramAccount, error) {
	account := &models.TelegramAccount{
		ChatID:           chatID,
		TelegramUsername: telegramUsername,
		Language:         language,
		OcservUserID:     ocservUserID,
	}
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND ocserv_user_id = ?", chatID, ocservUserID).
		FirstOrCreate(account).Error
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *Repository) DeleteAccount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.TelegramAccount{}).Error
}

func (r *Repository) UpdateLanguageForChat(ctx context.Context, chatID int64, language string) error {
	return r.db.WithContext(ctx).
		Model(&models.TelegramAccount{}).
		Where("chat_id = ?", chatID).
		Update("language", language).Error
}

func (r *Repository) AllAccounts(ctx context.Context) ([]models.TelegramAccount, error) {
	var accounts []models.TelegramAccount
	err := r.db.WithContext(ctx).Find(&accounts).Error
	return accounts, err
}

func (r *Repository) MarkLowQuotaNotified(ctx context.Context, id uint, at time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.TelegramAccount{}).
		Where("id = ?", id).
		Update("last_low_quota_notified_at", at).Error
}

// =============================================================================
// Ocserv users (read-only convenience)
// =============================================================================

func (r *Repository) OcservUserByID(ctx context.Context, id uint) (*models.OcservUser, error) {
	var user models.OcservUser
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) OcservUserByUsername(ctx context.Context, username string) (*models.OcservUser, error) {
	var user models.OcservUser
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// =============================================================================
// Packages
// =============================================================================

func (r *Repository) ActivePackages(ctx context.Context) ([]models.TelegramPackage, error) {
	var packages []models.TelegramPackage
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("id ASC").
		Find(&packages).Error
	return packages, err
}

func (r *Repository) PackageByID(ctx context.Context, id uint) (*models.TelegramPackage, error) {
	var pkg models.TelegramPackage
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&pkg).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

// =============================================================================
// Requests
// =============================================================================

func (r *Repository) PendingByChat(ctx context.Context, chatID int64) (*models.TelegramRequest, error) {
	var req models.TelegramRequest
	err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Where("status IN ?", []string{
			models.TelegramRequestStatusPending,
			models.TelegramRequestStatusAwaitingPayment,
			models.TelegramRequestStatusPaymentUploaded,
		}).
		Order("created_at DESC").
		First(&req).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *Repository) CreateRequest(ctx context.Context, req *models.TelegramRequest) (*models.TelegramRequest, error) {
	if err := r.db.WithContext(ctx).Create(req).Error; err != nil {
		return nil, err
	}
	return req, nil
}

func (r *Repository) AttachReceipt(ctx context.Context, id uint, path string) error {
	return r.db.WithContext(ctx).
		Model(&models.TelegramRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"receipt_file_path": path,
			"status":            models.TelegramRequestStatusPaymentUploaded,
		}).Error
}

func (r *Repository) RequestByID(ctx context.Context, id uint) (*models.TelegramRequest, error) {
	var req models.TelegramRequest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}
