package repository

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	"gorm.io/gorm"
)

type OcservAgentRepository struct {
	db *gorm.DB
}

func NewOcservAgentRepository() *OcservAgentRepository {
	return &OcservAgentRepository{db: database.GetConnection()}
}

func (r *OcservAgentRepository) List(ctx context.Context) ([]models.OcservAgent, error) {
	var agents []models.OcservAgent
	err := r.db.WithContext(ctx).Order("id ASC").Find(&agents).Error
	return agents, err
}

func (r *OcservAgentRepository) GetByID(ctx context.Context, id uint) (*models.OcservAgent, error) {
	var agent models.OcservAgent
	if err := r.db.WithContext(ctx).First(&agent, id).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *OcservAgentRepository) Create(ctx context.Context, agent *models.OcservAgent) error {
	return r.db.WithContext(ctx).Create(agent).Error
}

func (r *OcservAgentRepository) Update(ctx context.Context, agent *models.OcservAgent) error {
	return r.db.WithContext(ctx).Save(agent).Error
}

func (r *OcservAgentRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.OcservAgent{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
