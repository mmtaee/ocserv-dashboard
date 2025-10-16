package repository

import (
	"context"
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/group"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"gorm.io/gorm"
	"log"
)

type OcservGroupRepository struct {
	db                    *gorm.DB
	commonOcservGroupRepo group.OcservGroupInterface
}

type OcservGroupRepositoryInterface interface {
	Groups(ctx context.Context, pagination *request.Pagination) (*[]models.OcservGroup, int64, error)
	GroupsLookup(ctx context.Context) ([]string, error)
	GetByID(ctx context.Context, id string) (*models.OcservGroup, error)
	Create(ctx context.Context, ocservGroup *models.OcservGroup) (*models.OcservGroup, error)
	Update(ctx context.Context, ocservGroup *models.OcservGroup) (*models.OcservGroup, error)
	Delete(ctx context.Context, id string) (*models.OcservGroup, error)
	DefaultGroup() (*models.OcservGroupConfig, error)
	UpdateDefaultGroup(groupConfig *models.OcservGroupConfig) error
}

func NewOcservGroupRepository() *OcservGroupRepository {
	return &OcservGroupRepository{
		db:                    database.GetConnection(),
		commonOcservGroupRepo: group.NewOcservGroup(),
	}
}

func (o *OcservGroupRepository) Groups(
	ctx context.Context, pagination *request.Pagination,
) (*[]models.OcservGroup, int64, error) {
	var totalRecords int64

	err := o.db.WithContext(ctx).Model(&models.OcservGroup{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}

	var ocservGroups []models.OcservGroup
	txPaginator := request.Paginator(ctx, o.db, pagination)
	err = txPaginator.Model(&ocservGroups).Find(&ocservGroups).Error
	if err != nil {
		return nil, 0, err
	}
	return &ocservGroups, totalRecords, nil
}

func (o *OcservGroupRepository) GroupsLookup(ctx context.Context) ([]string, error) {
	var ocservGroups []models.OcservGroup

	err := o.db.WithContext(ctx).Model(&models.OcservGroup{}).Select("name").Find(&ocservGroups).Error
	if err != nil {
		return nil, err
	}

	groups := make([]string, 0, len(ocservGroups))
	for _, ocservGroup := range ocservGroups {
		groups = append(groups, ocservGroup.Name)
	}
	return groups, nil
}

func (o *OcservGroupRepository) GetByID(ctx context.Context, id string) (*models.OcservGroup, error) {
	var ocservGroup models.OcservGroup
	err := o.db.WithContext(ctx).Where("id = ?", id).First(&ocservGroup).Error
	if err != nil {
		return nil, err
	}
	return &ocservGroup, nil
}

func (o *OcservGroupRepository) Create(ctx context.Context, ocservGroup *models.OcservGroup) (*models.OcservGroup, error) {
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ocservGroup).Error; err != nil {
			return err
		}
		if err := o.commonOcservGroupRepo.Create(ocservGroup.Name, ocservGroup.Config); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return ocservGroup, nil
}

func (o *OcservGroupRepository) Update(ctx context.Context, ocservGroup *models.OcservGroup) (*models.OcservGroup, error) {
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(ocservGroup).Save(ocservGroup).Error; err != nil {
			return err
		}
		if err := o.commonOcservGroupRepo.Create(ocservGroup.Name, ocservGroup.Config); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ocservGroup, nil
}

func (o *OcservGroupRepository) Delete(ctx context.Context, id string) (*models.OcservGroup, error) {
	var ocservGroup models.OcservGroup
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).First(&ocservGroup).Error; err != nil {
			return err
		}

		if err := tx.Delete(&ocservGroup).Error; err != nil {
			log.Println("err: ", err)
			return err
		}

		if err := o.commonOcservGroupRepo.Delete(ocservGroup.Name); err != nil {
			return err
		}
		return nil
	})

	return &ocservGroup, err
}

func (o *OcservGroupRepository) DefaultGroup() (*models.OcservGroupConfig, error) {
	defaultsGroup, err := o.commonOcservGroupRepo.DefaultsGroup()
	if err != nil {
		return nil, err
	}
	return defaultsGroup, nil
}

func (o *OcservGroupRepository) UpdateDefaultGroup(groupConfig *models.OcservGroupConfig) error {
	return o.commonOcservGroupRepo.UpdateDefaultsGroup(groupConfig)

}
