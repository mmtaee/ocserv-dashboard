package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/ocserv/group"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
)

type OcservGroupUsecaseInterface interface {
	GroupsLookup(ctx context.Context, owner string) ([]string, error)
	Groups(ctx context.Context, pagination *request.Pagination, owner string) ([]models.OcservGroup, int64, error)
	GetByID(ctx context.Context, groupID string) (*models.OcservGroup, error)
	Create(ctx context.Context, ocservGroup *models.OcservGroup) (*models.OcservGroup, error)
	Update(ctx context.Context, ocservGroup *models.OcservGroup) (*models.OcservGroup, error)
	Delete(ctx context.Context, groupID string) (*models.OcservGroup, error)
	UpdateUsersByDeleteGroup(ctx context.Context, groupName string) ([]models.OcservUser, error)
	DefaultGroup() (*models.OcservGroupConfig, error)
	UpdateDefaultGroup(config *models.OcservGroupConfig) error
	ListUnsyncedGroups(ctx context.Context) ([]group.UnsyncedGroup, error)
	GroupSyncToDB(ctx context.Context, groups []models.OcservGroup) ([]models.OcservGroup, error)
}

type OcservGroupUsecase struct {
	ocservGroupRepo repository.OcservGroupRepositoryInterface
	ocservUserRepo  repository.OcservUserRepositoryInterface
}

func NewOcservGroupUsecase(
	ocservGroupRepo repository.OcservGroupRepositoryInterface,
	ocservUserRepo repository.OcservUserRepositoryInterface,
) *OcservGroupUsecase {
	return &OcservGroupUsecase{
		ocservGroupRepo: ocservGroupRepo,
		ocservUserRepo:  ocservUserRepo,
	}
}

func (uc *OcservGroupUsecase) GroupsLookup(ctx context.Context, owner string) ([]string, error) {
	return uc.ocservGroupRepo.GroupsLookup(ctx, owner)
}

func (uc *OcservGroupUsecase) Groups(ctx context.Context, pagination *request.Pagination, owner string) ([]models.OcservGroup, int64, error) {
	return uc.ocservGroupRepo.Groups(ctx, pagination, owner)
}

func (uc *OcservGroupUsecase) GetByID(ctx context.Context, groupID string) (*models.OcservGroup, error) {
	return uc.ocservGroupRepo.GetByID(ctx, groupID)
}

func (uc *OcservGroupUsecase) Create(ctx context.Context, ocservGroup *models.OcservGroup) (*models.OcservGroup, error) {
	return uc.ocservGroupRepo.Create(ctx, ocservGroup)
}

func (uc *OcservGroupUsecase) Update(ctx context.Context, ocservGroup *models.OcservGroup) (*models.OcservGroup, error) {
	return uc.ocservGroupRepo.Update(ctx, ocservGroup)
}

func (uc *OcservGroupUsecase) Delete(ctx context.Context, groupID string) (*models.OcservGroup, error) {
	group, err := uc.ocservGroupRepo.Delete(ctx, groupID)
	if err != nil {
		return nil, err
	}

	go func(groupName string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var wg sync.WaitGroup

		users, err := uc.ocservUserRepo.UpdateUsersByDeleteGroup(ctx, groupName)
		if err != nil {
			logger.Error("Failed to load users for removed group %s: %v", groupName, err)
			return
		}

		for _, u := range users {
			ocservUser := u
			ocservUser.Group = "defaults"

			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, err2 := uc.ocservUserRepo.Update(ctx, &ocservUser); err2 != nil {
					logger.Warn("DeleteGroup: failed to update user %s: %v", ocservUser.Username, err2)
				}
			}()
		}

		wg.Wait()
	}(group.Name)

	return group, nil
}

func (uc *OcservGroupUsecase) UpdateUsersByDeleteGroup(ctx context.Context, groupName string) ([]models.OcservUser, error) {
	return uc.ocservUserRepo.UpdateUsersByDeleteGroup(ctx, groupName)
}

func (uc *OcservGroupUsecase) DefaultGroup() (*models.OcservGroupConfig, error) {
	return uc.ocservGroupRepo.DefaultGroup()
}

func (uc *OcservGroupUsecase) UpdateDefaultGroup(config *models.OcservGroupConfig) error {
	return uc.ocservGroupRepo.UpdateDefaultGroup(config)
}

func (uc *OcservGroupUsecase) ListUnsyncedGroups(ctx context.Context) ([]group.UnsyncedGroup, error) {
	return uc.ocservGroupRepo.ListUnsyncedGroups(ctx)
}

func (uc *OcservGroupUsecase) GroupSyncToDB(ctx context.Context, groups []models.OcservGroup) ([]models.OcservGroup, error) {
	return uc.ocservGroupRepo.GroupSyncToDB(ctx, groups)
}
