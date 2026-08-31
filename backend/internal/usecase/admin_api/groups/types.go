package ocservgroup

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/group"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Repository interface {
	repository.OcservGroupRepositoryInterface
}

type ConfigStore interface {
	Create(name string, config *models.OcservGroupConfig) error
	Delete(name string) error
	DefaultsGroup() (*models.OcservGroupConfig, error)
	UpdateDefaultsGroup(config *models.OcservGroupConfig) error
	GroupList(ctx context.Context) ([]group.UnsyncedGroup, error)
}

type Reloader interface {
	Reload() (string, error)
}

type UserUpdater interface {
	UpdateUsersByDeleteGroup(ctx context.Context, groupName string) ([]models.OcservUser, error)
	Update(ctx context.Context, user *models.OcservUser) (*models.OcservUser, error)
}

type CreateInput struct {
	Name   string                    `json:"name" validate:"required"`
	Config *models.OcservGroupConfig `json:"config" validate:"required"`
}

type UpdateInput struct {
	Config *models.OcservGroupConfig `json:"config" validate:"required"`
}

type SyncInput struct {
	Groups []group.UnsyncedGroup `json:"groups" validate:"required,dive"`
}

type ListResult struct {
	Groups []models.OcservGroupWithTraffic
	Total  int64
	Page   int
	Size   int
}

type ListOptions struct {
	Pagination *request.Pagination
}
