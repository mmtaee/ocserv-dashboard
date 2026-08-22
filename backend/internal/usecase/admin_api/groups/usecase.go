package ocservgroup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/group"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
)

type Usecase struct {
	repository Repository
	users      UserUpdater
	configs    ConfigStore
	reloader   Reloader
}

func New(repo Repository, users UserUpdater, configs ConfigStore, reloader Reloader) *Usecase {
	return &Usecase{repository: repo, users: users, configs: configs, reloader: reloader}
}

func (u *Usecase) Lookup(ctx context.Context, owner string) ([]string, error) {
	groups, err := u.repository.GroupsLookup(ctx, owner)
	if err != nil {
		return nil, err
	}
	return append([]string{"defaults"}, groups...), nil
}

func (u *Usecase) List(ctx context.Context, options ListOptions) (*ListResult, error) {
	groups, total, err := u.repository.Groups(ctx, options.Pagination, options.Owner)
	if err != nil {
		return nil, err
	}
	return &ListResult{Groups: groups, Total: total, Page: options.Pagination.Page, Size: options.Pagination.PageSize}, nil
}

func (u *Usecase) Get(ctx context.Context, id string) (*models.OcservGroup, error) {
	if id == "" {
		return nil, errors.New("invalid group id")
	}
	return u.repository.GetByID(ctx, id)
}

func (u *Usecase) Create(ctx context.Context, owner string, input CreateInput) (*models.OcservGroup, error) {
	if owner == "" {
		return nil, errors.New("admin or staff username not found")
	}
	created, err := u.repository.Create(ctx, &models.OcservGroup{Name: input.Name, Owner: owner, Config: input.Config})
	if err != nil {
		return nil, err
	}
	if err := u.configs.Create(created.Name, created.Config); err != nil {
		_, _ = u.repository.Delete(ctx, fmt.Sprint(created.ID))
		return nil, err
	}
	u.reload()
	return created, nil
}

func (u *Usecase) Update(ctx context.Context, id string, input UpdateInput) (*models.OcservGroup, error) {
	group, err := u.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	previousConfig := group.Config
	group.Config = input.Config
	updated, err := u.repository.Update(ctx, group)
	if err != nil {
		return nil, err
	}
	if err := u.configs.Create(updated.Name, updated.Config); err != nil {
		updated.Config = previousConfig
		_, _ = u.repository.Update(ctx, updated)
		return nil, err
	}
	u.reload()
	return updated, nil
}

func (u *Usecase) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("group id is empty")
	}
	deleted, err := u.repository.Delete(ctx, id)
	if err != nil {
		return err
	}
	if err := u.configs.Delete(deleted.Name); err != nil {
		_, _ = u.repository.Create(ctx, deleted)
		return err
	}
	u.reload()

	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	users, err := u.users.UpdateUsersByDeleteGroup(updateCtx, deleted.Name)
	if err != nil {
		logger.Error("Failed to load users for removed group %s: %v", deleted.Name, err)
		return nil
	}
	for i := range users {
		if err := updateCtx.Err(); err != nil {
			return err
		}
		users[i].Group = "defaults"
		if _, err := u.users.Update(updateCtx, &users[i]); err != nil {
			logger.Warn("DeleteGroup: failed to update user %s: %v", users[i].Username, err)
		}
	}
	return nil
}

func (u *Usecase) Defaults() (*models.OcservGroupConfig, error) {
	return u.configs.DefaultsGroup()
}

// DefaultGroup is retained for the backup usecase while backup orchestration is migrated.
func (u *Usecase) DefaultGroup() (*models.OcservGroupConfig, error) {
	return u.Defaults()
}

func (u *Usecase) UpdateDefaults(config *models.OcservGroupConfig) error {
	if err := u.configs.UpdateDefaultsGroup(config); err != nil {
		return err
	}
	u.reload()
	return nil
}

// UpdateDefaultGroup is retained for the backup usecase while backup orchestration is migrated.
func (u *Usecase) UpdateDefaultGroup(config *models.OcservGroupConfig) error {
	return u.UpdateDefaults(config)
}

func (u *Usecase) Unsynced(ctx context.Context) ([]group.UnsyncedGroup, error) {
	groups, err := u.configs.GroupList(ctx)
	if err != nil || len(groups) == 0 {
		return groups, err
	}
	names := make([]string, 0, len(groups))
	for _, item := range groups {
		names = append(names, item.Name)
	}
	existing, err := u.repository.ExistingNames(ctx, names)
	if err != nil {
		return nil, err
	}
	existingSet := make(map[string]struct{}, len(existing))
	for _, name := range existing {
		existingSet[name] = struct{}{}
	}
	unsynced := make([]group.UnsyncedGroup, 0, len(groups))
	for _, item := range groups {
		if _, found := existingSet[item.Name]; !found {
			unsynced = append(unsynced, item)
		}
	}
	return unsynced, nil
}

func (u *Usecase) Sync(ctx context.Context, owner string, input SyncInput) ([]string, error) {
	if owner == "" {
		return nil, errors.New("admin or staff username not found")
	}
	if len(input.Groups) == 0 {
		return nil, errors.New("no groups found")
	}
	groups := make([]models.OcservGroup, 0, len(input.Groups))
	for _, item := range input.Groups {
		groups = append(groups, models.OcservGroup{Name: item.Name, Config: item.Config, Owner: owner})
	}
	synced, err := u.repository.CreateMany(ctx, groups)
	if err != nil {
		return nil, fmt.Errorf("sync groups: %w", err)
	}
	names := make([]string, 0, len(synced))
	for _, item := range synced {
		names = append(names, item.Name)
	}
	u.reload()
	return names, nil
}

func (u *Usecase) reload() {
	if _, err := u.reloader.Reload(); err != nil {
		logger.Warn("Failed to reload ocserv configuration: %v", err)
	}
}
