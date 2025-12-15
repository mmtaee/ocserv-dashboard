package ocserv_group

import (
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/group"
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
