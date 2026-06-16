package repository

import (
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"gorm.io/gorm"
)

type OcservGroupRepository interface {
	FindAll(adminID uint, role string) ([]models.OcservGroupResponse, error)
	FindByName(name string, adminID uint, role string) (*models.OcservGroup, error)
	FindByID(id uint, adminID uint, role string) (*models.OcservGroupResponse, error)
	Create(group *models.OcservGroup) error
	Update(group *models.OcservGroup) error
	Delete(id uint, adminID uint, role string) error
	GetGroupsLookup(adminID uint, role string) ([]string, error)
}

type ocservGroupRepository struct {
	db *gorm.DB
}

func NewOcservGroupRepository(db *gorm.DB) OcservGroupRepository {
	return &ocservGroupRepository{db: db}
}

func (r *ocservGroupRepository) FindAll(adminID uint, role string) ([]models.OcservGroupResponse, error) {
	var groups []models.OcservGroup
	query := r.db
	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}
	err := query.Find(&groups).Error
	if err != nil {
		return nil, err
	}

	// Calculate total RX and TX for each group
	var responses []models.OcservGroupResponse
	for _, group := range groups {
		var stats struct {
			TotalRx int
			TotalTx int
		}
		userQuery := r.db.Model(&models.OcservUser{}).
			Where("`group` = ?", group.Name)
		if role != models.AdminRoleSuper {
			userQuery = userQuery.Where("owner_admin_id = ?", adminID)
		}
		userQuery.Select("COALESCE(SUM(rx), 0) as total_rx, COALESCE(SUM(tx), 0) as total_tx").
			Scan(&stats)

		responses = append(responses, models.OcservGroupResponse{
			OcservGroup: &group,
			TotalRx:     stats.TotalRx,
			TotalTx:     stats.TotalTx,
		})
	}

	return responses, nil
}

func (r *ocservGroupRepository) FindByName(name string, adminID uint, role string) (*models.OcservGroup, error) {
	var group models.OcservGroup
	query := r.db.Where("name = ?", name)
	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}
	err := query.First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *ocservGroupRepository) FindByID(id uint, adminID uint, role string) (*models.OcservGroupResponse, error) {
	var group models.OcservGroup
	query := r.db.Where("id = ?", id)
	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}
	err := query.First(&group).Error
	if err != nil {
		return nil, err
	}

	// Calculate total RX and TX for the group
	var stats struct {
		TotalRx int
		TotalTx int
	}
	userQuery := r.db.Model(&models.OcservUser{}).
		Where("`group` = ?", group.Name)
	if role != models.AdminRoleSuper {
		userQuery = userQuery.Where("owner_admin_id = ?", adminID)
	}
	userQuery.Select("COALESCE(SUM(rx), 0) as total_rx, COALESCE(SUM(tx), 0) as total_tx").
		Scan(&stats)

	return &models.OcservGroupResponse{
		OcservGroup: &group,
		TotalRx:     stats.TotalRx,
		TotalTx:     stats.TotalTx,
	}, nil
}

func (r *ocservGroupRepository) Create(group *models.OcservGroup) error {
	return r.db.Create(group).Error
}

func (r *ocservGroupRepository) Update(group *models.OcservGroup) error {
	return r.db.Save(group).Error
}

func (r *ocservGroupRepository) Delete(id uint, adminID uint, role string) error {
	query := r.db.Where("id = ?", id)
	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}
	return query.Delete(&models.OcservGroup{}).Error
}

func (r *ocservGroupRepository) GetGroupsLookup(adminID uint, role string) ([]string, error) {
	var names []string
	query := r.db.Model(&models.OcservGroup{})
	if role != models.AdminRoleSuper {
		query = query.Where("owner_admin_id = ?", adminID)
	}
	err := query.Pluck("name", &names).Error
	return names, err
}
