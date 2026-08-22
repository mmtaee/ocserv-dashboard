package customerapi

import (
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	ocservaccount "github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/user"
	platformocserv "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/ocserv"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	ciscosetup "github.com/mmtaee/ocserv-dashboard/backend/internal/services/customer_api/cisco_setup"
	customerusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/customer_api"
)

type Service struct {
	controller *CustomerController
	ciscoSetup *ciscosetup.Controller
}

// New constructs the Customer API dependency graph.
func New(cfg *config.Config) *Service {
	usecase := customerusecase.New(
		repository.NewSystemRepository(),
		repository.NewtOcservUserRepository(),
		platformocserv.NewClient(),
		cfg.SecretKey,
		ocservaccount.NewOcservUser(),
	)
	return &Service{controller: newCustomerController(usecase), ciscoSetup: ciscosetup.New(usecase)}
}
