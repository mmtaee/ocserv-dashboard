package customer

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/internal/repository"
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
	"net/http"
	"strconv"
	"time"
)

type Controller struct {
	request        request.CustomRequestInterface
	ocservUserRepo repository.OcservUserRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request:        request.NewCustomRequest(),
		ocservUserRepo: repository.NewtOcservUserRepository(),
	}
}

// Summary 	     Customer summary account
//
// @Summary      Customer summary account
// @Description  Customer summary account
// @Tags         Customers
// @Accept       json
// @Produce      json
// @Param        request body  SummaryData  true "customer username and password (same ocserv account)."
// @Failure      400 {object} request.ErrorResponse
// @Failure      429 {object} middlewares.TooManyRequests
// @Success      200  {object} SummaryResponse
// @Router       /customers/summary [post]
func (ctl *Controller) Summary(c echo.Context) error {
	var data SummaryData

	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	user, err := ctl.ocservUserRepo.GetByUsername(c.Request().Context(), data.Username)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	if user.Password != data.Password {
		return ctl.request.BadRequest(c, errors.New("invalid username or password"))
	}

	dateEnd := time.Now()
	firstOfThisMonth := time.Date(dateEnd.Year(), dateEnd.Month(), 1, 0, 0, 0, 0, dateEnd.Location())
	dateStart := firstOfThisMonth.AddDate(0, -1, 0)

	usage, err := ctl.ocservUserRepo.TotalBandwidthUserDateRange(
		c.Request().Context(),
		strconv.Itoa(int(user.ID)),
		&dateStart,
		&dateEnd,
	)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, SummaryResponse{
		OcservUser: ModelCustomer{
			Owner:         user.Owner,
			Username:      user.Username,
			IsLocked:      user.IsLocked,
			ExpireAt:      user.ExpireAt,
			DeactivatedAt: user.DeactivatedAt,
			TrafficType:   user.TrafficType,
			TrafficSize:   user.TrafficSize,
			Rx:            user.Rx,
			Tx:            user.Tx,
		},
		Usage: UsageResponse{
			DateStart:  dateStart,
			DateEnd:    dateEnd,
			Bandwidths: usage,
		},
	})
}
