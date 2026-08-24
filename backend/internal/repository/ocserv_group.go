package repository

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
	"gorm.io/gorm"
)

type OcservGroupRepository struct {
	db *gorm.DB
}

type OcservGroupRepositoryInterface interface {
	Groups(ctx context.Context, pagination *request.Pagination) ([]models.OcservGroup, int64, error)
	GroupsLookup(ctx context.Context) ([]string, error)
	GetByID(ctx context.Context, id string) (*models.OcservGroup, error)
	Create(ctx context.Context, group *models.OcservGroup) (*models.OcservGroup, error)
	Update(ctx context.Context, group *models.OcservGroup) (*models.OcservGroup, error)
	Delete(ctx context.Context, id string) (*models.OcservGroup, error)
	ExistingNames(ctx context.Context, names []string) ([]string, error)
	CreateMany(ctx context.Context, groups []models.OcservGroup) ([]models.OcservGroup, error)
}

func NewOcservGroupRepository() *OcservGroupRepository {
	return &OcservGroupRepository{db: database.GetConnection()}
}

func (r *OcservGroupRepository) Groups(ctx context.Context, pagination *request.Pagination) ([]models.OcservGroup, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.OcservGroup{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var groups []models.OcservGroup
	query = request.Paginator(ctx, r.db, pagination).Model(&models.OcservGroup{})
	if err := query.Find(&groups).Error; err != nil {
		return nil, 0, err
	}
	return groups, total, nil
}

func (r *OcservGroupRepository) GroupsLookup(ctx context.Context) ([]string, error) {
	query := r.db.WithContext(ctx).Model(&models.OcservGroup{})
	var names []string
	return names, query.Pluck("name", &names).Error
}

func (r *OcservGroupRepository) GetByID(ctx context.Context, id string) (*models.OcservGroup, error) {
	var group models.OcservGroup
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *OcservGroupRepository) Create(ctx context.Context, group *models.OcservGroup) (*models.OcservGroup, error) {
	return group, r.db.WithContext(ctx).Create(group).Error
}

func (r *OcservGroupRepository) Update(ctx context.Context, group *models.OcservGroup) (*models.OcservGroup, error) {
	return group, r.db.WithContext(ctx).Save(group).Error
}

func (r *OcservGroupRepository) Delete(ctx context.Context, id string) (*models.OcservGroup, error) {
	var group models.OcservGroup
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).First(&group).Error; err != nil {
			return err
		}
		return tx.Delete(&group).Error
	})
	return &group, err
}

func (r *OcservGroupRepository) ExistingNames(ctx context.Context, names []string) ([]string, error) {
	var existing []string
	err := r.db.WithContext(ctx).Model(&models.OcservGroup{}).Where("name IN ?", names).Pluck("name", &existing).Error
	return existing, err
}

func (r *OcservGroupRepository) CreateMany(ctx context.Context, groups []models.OcservGroup) ([]models.OcservGroup, error) {
	if len(groups) == 0 {
		return groups, nil
	}
	err := r.db.WithContext(ctx).Create(&groups).Error
	return groups, err
}
