package customerapi

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	customerusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/customer"
	customer "github.com/mmtaee/ocserv-dashboard/backend/services/customer_api/internal/controller"
)

type Service struct {
	controller *customer.Controller
}

func New(cfg *config.Config) *Service {
	usecase := customerusecase.New(
		repository.NewSystemRepository(),
		repository.NewtOcservUserRepository(),
		repository.NewOcctlRepository(),
		cfg.SecretKey,
	)
	return &Service{controller: customer.New(usecase)}
}

func (s *Service) Register(group *echo.Group) {
	customer.Routes(group, s.controller)
}
