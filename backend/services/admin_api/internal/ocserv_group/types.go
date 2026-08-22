package ocserv_group

import (
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/group"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type CreateOcservGroupData struct {
	Name   string                    `json:"name" validate:"required"`
	Config *models.OcservGroupConfig `json:"config" validate:"required"`
}

type UpdateOcservGroupData struct {
	Config *models.OcservGroupConfig `json:"config" validate:"required"`
}

type OcservGroupsResponse struct {
	Meta   request.Meta         `json:"meta" validate:"required"`
	Result []models.OcservGroup `json:"result" validate:"omitempty"`
}

type SyncGroupRequest struct {
	Groups []group.UnsyncedGroup `json:"groups" validate:"required,dive"`
}
