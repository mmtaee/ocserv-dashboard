package ocserv_group

import "github.com/mmtaee/ocserv-dashboard/core/models"

type CreateOcservGroupRequest struct {
	Name   string             `json:"name" validate:"required"`
	Config *models.OcservGroupConfig `json:"config"`
}

type UpdateOcservGroupRequest struct {
	Config *models.OcservGroupConfig `json:"config"`
}
