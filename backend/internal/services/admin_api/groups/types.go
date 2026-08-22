package groups

import (
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	groupusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/groups"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type CreateOcservGroupData = groupusecase.CreateInput

type UpdateOcservGroupData = groupusecase.UpdateInput

type OcservGroupsResponse struct {
	Meta   request.Meta         `json:"meta" validate:"required"`
	Result []models.OcservGroup `json:"result" validate:"omitempty"`
}

type SyncGroupRequest = groupusecase.SyncInput
