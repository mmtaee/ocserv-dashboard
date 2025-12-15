package repository

import (
	"context"
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/group"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/occtl"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"gorm.io/gorm"
)

type OcservGroupRepository struct {
	db                    *gorm.DB
	commonOcservGroupRepo group.OcservGroupInterface
	commonOcservOcctlRepo occtl.OcservOcctlInterface
}

type OcservGroupCRUD interface {
	Groups(ctx context.Context, pagination *request.Pagination, owner string) ([]models.OcservGroup, int64, error)
	GroupsLookup(ctx context.Context, owner string) ([]string, error)
	GetByID(ctx context.Context, id string) (*models.OcservGroup, error)
	Create(ctx context.Context, ocservGroup *models.OcservGroup) (*models.OcservGroup, error)
	Update(ctx context.Context, ocservGroup *models.OcservGroup) (*models.OcservGroup, error)
	Delete(ctx context.Context, id string) (*models.OcservGroup, error)
}

type OcservDefaultGroup interface {
	DefaultGroup() (*models.OcservGroupConfig, error)
	UpdateDefaultGroup(groupConfig *models.OcservGroupConfig) error
}

type OcservGroupSync interface {
	ListUnsyncedGroups(ctx context.Context) ([]group.UnsyncedGroup, error)
	GroupSyncToDB(ctx context.Context, groups []models.OcservGroup) ([]models.OcservGroup, error)
}

type OcservGroupRepositoryInterface interface {
	OcservGroupCRUD
	OcservDefaultGroup
	OcservGroupSync
}

func NewOcservGroupRepository() *OcservGroupRepository {
	return &OcservGroupRepository{
		db:                    database.GetConnection(),
		commonOcservGroupRepo: group.NewOcservGroup(),
		commonOcservOcctlRepo: occtl.NewOcservOcctl(),
	}
}

func (o *OcservGroupRepository) Groups(ctx context.Context, pagination *request.Pagination, owner string) (
	[]models.OcservGroup, int64, error,
) {
	var totalRecords int64

	totalQuery := o.db.WithContext(ctx).Model(&models.OcservGroup{})
	if owner != "" {
		totalQuery = totalQuery.Where("owner = ?", owner)
	}
	err := totalQuery.Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}

	var ocservGroups []models.OcservGroup
	txPaginator := request.Paginator(ctx, o.db, pagination)

	query := txPaginator.Model(&ocservGroups)
	if owner != "" {
		query = query.Where("owner = ?", owner)
	}
	err = query.Find(&ocservGroups).Error
	if err != nil {
		return nil, 0, err
	}
	return ocservGroups, totalRecords, nil
}

func (o *OcservGroupRepository) GroupsLookup(ctx context.Context, owner string) ([]string, error) {
	var ocservGroups []models.OcservGroup

	query := o.db.WithContext(ctx).Model(&models.OcservGroup{})
	if owner != "" {
		query = query.Where("owner = ?", owner)
	}

	err := query.Select("name").Find(&ocservGroups).Error
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

	go func() {
		_, _ = o.commonOcservOcctlRepo.ReloadConfigs()
	}()

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

	go func() {
		_, _ = o.commonOcservOcctlRepo.ReloadConfigs()
	}()

	return ocservGroup, nil
}

func (o *OcservGroupRepository) Delete(ctx context.Context, id string) (*models.OcservGroup, error) {
	var ocservGroup models.OcservGroup
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).First(&ocservGroup).Error; err != nil {
			return err
		}

		if err := tx.Delete(&ocservGroup).Error; err != nil {
			return err
		}

		if err := o.commonOcservGroupRepo.Delete(ocservGroup.Name); err != nil {
			return err
		}

		go func() {
			_, _ = o.commonOcservOcctlRepo.ReloadConfigs()
		}()

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
	err := o.commonOcservGroupRepo.UpdateDefaultsGroup(groupConfig)
	if err != nil {
		return err
	}

	go func() {
		_, _ = o.commonOcservOcctlRepo.ReloadConfigs()
	}()
	return nil
}

func (o *OcservGroupRepository) ListUnsyncedGroups(ctx context.Context) ([]group.UnsyncedGroup, error) {
	groups, err := o.commonOcservGroupRepo.GroupList(ctx)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return []group.UnsyncedGroup{}, nil
	}

	var groupsName []string
	for _, g := range groups {
		groupsName = append(groupsName, g.Name)
	}

	var existing []string
	if err = o.db.WithContext(ctx).
		Model(&models.OcservGroup{}).
		Where("name IN ?", groupsName).
		Pluck("name", &existing).Error; err != nil {
		return nil, err
	}

	existingMap := make(map[string]struct{}, len(existing))
	for _, name := range existing {
		existingMap[name] = struct{}{}
	}

	var unsynced []group.UnsyncedGroup

	for _, g := range groups {
		if _, found := existingMap[g.Name]; !found {
			unsynced = append(unsynced, group.UnsyncedGroup{
				Name:   g.Name,
				Path:   g.Path,
				Config: g.Config,
			})
		}
	}

	return unsynced, nil
}

func (o *OcservGroupRepository) GroupSyncToDB(ctx context.Context, groups []models.OcservGroup) ([]models.OcservGroup, error) {
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&groups).Error; err != nil {
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

	return groups, nil
}
