package repository

import (
	"context"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
	"gorm.io/gorm"
)

type UserTokenRepository struct {
	db *gorm.DB
}

func NewUserTokenRepository() *UserTokenRepository {
	return &UserTokenRepository{db: database.GetConnection()}
}

func (r *UserTokenRepository) Create(ctx context.Context, session *models.UserToken) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *UserTokenRepository) FindActiveByToken(ctx context.Context, token string) (*models.UserToken, error) {
	var session models.UserToken
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("token = ? AND expire_at > ?", crypto.HashToken(token), time.Now()).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *UserTokenRepository) DeleteByID(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.UserToken{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserTokenRepository) List(ctx context.Context, pagination *request.Pagination) ([]models.UserToken, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.UserToken{}).Where("expire_at > ?", time.Now()).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var sessions []models.UserToken
	err := request.Paginator(ctx, r.db, pagination).
		Preload("User").
		Where("expire_at > ?", time.Now()).
		Find(&sessions).Error
	return sessions, total, err
}
