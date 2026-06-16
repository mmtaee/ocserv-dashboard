package group

import "github.com/mmtaee/ocserv-dashboard/core/models"

type UnsyncedGroup struct {
	Name   string                    `json:"name" validate:"required"`
	Path   string                    `json:"path" validate:"omitempty"`
	Config *models.OcservGroupConfig `json:"config" validate:"required"`
}
