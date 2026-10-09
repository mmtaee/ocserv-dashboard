package occtl

import platformocserv "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/ocserv"

// Repository is the persistence boundary required by this usecase.
type Repository interface {
	platformocserv.ClientInterface
}

type CommandInput struct {
	Action int    `query:"action" validate:"required,min=1,max=16"`
	Value  string `query:"value" validate:"omitempty"`
}
