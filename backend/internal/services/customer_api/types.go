package customerapi

import (
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	customerusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/customer_api"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type LoginData = customerusecase.Credentials
type ChangePasswordInput = customerusecase.ChangePasswordInput
type DateRange = customerusecase.DateRange
type ModelCustomer = customerusecase.Customer
type UsageResponse = customerusecase.Usage
type SummaryResponse = customerusecase.Summary
type LoginResponse = customerusecase.LoginResponse
type CiscoSetupResponse = customerusecase.CiscoSetup

type ActivitiesResponse struct {
	Meta   request.Meta                   `json:"meta" validate:"required"`
	Result *[]models.OcservUserSessionLog `json:"result" validate:"omitempty"`
}
