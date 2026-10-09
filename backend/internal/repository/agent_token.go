package repository

import (
	"context"
	"errors"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	"gorm.io/gorm"
)

type AgentTokenRepository struct {
	db *gorm.DB
}

func NewAgentTokenRepository() *AgentTokenRepository {
	return &AgentTokenRepository{db: database.GetConnection()}
}

func (r *AgentTokenRepository) Get(ctx context.Context) (*models.AgentToken, error) {
	var result models.AgentToken
	err := r.db.WithContext(ctx).First(&result).Error
	return &result, err
}

func (r *AgentTokenRepository) Create(ctx context.Context, token string) (*models.AgentToken, error) {
	var result models.AgentToken
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("LOCK TABLE agent_tokens IN EXCLUSIVE MODE").Error; err != nil {
			return err
		}
		if err := tx.First(&result).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
			if err == nil {
				return gorm.ErrDuplicatedKey
			}
			return err
		}
		result.Token = token
		return tx.Create(&result).Error
	})
	return &result, err
}

func (r *AgentTokenRepository) Replace(ctx context.Context, token string) (*models.AgentToken, error) {
	var result models.AgentToken
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("LOCK TABLE agent_tokens IN EXCLUSIVE MODE").Error; err != nil {
			return err
		}
		if err := tx.First(&result).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			result.Token = token
			return tx.Create(&result).Error
		} else if err != nil {
			return err
		}
		return tx.Model(&result).Update("token", token).Error
	})
	result.Token = token
	return &result, err
}

func (r *AgentTokenRepository) Delete(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("1 = 1").Delete(&models.AgentToken{}).Error
}
