package customer

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
)

type Controller struct {
	request        request.CustomRequestInterface
	systemRepo     repository.SystemRepositoryInterface
	ocservUserRepo repository.OcservUserRepositoryInterface
	occtl          repository.OcctlRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request:        request.NewCustomRequest(),
		systemRepo:     repository.NewSystemRepository(),
		ocservUserRepo: repository.NewtOcservUserRepository(),
		occtl:          repository.NewOcctlRepository(),
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

	if data.Password == "Secret-Ocpasswd" {
		return ctl.request.BadRequest(c, errors.New("invalid username or password"))
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
			Owner:                user.Owner,
			Username:             user.Username,
			IsLocked:             user.IsLocked,
			CertificateEnabled:   user.CertificateEnabled,
			CertificateAvailable: user.CertificateAvailable,
			ExpireAt:             user.ExpireAt,
			DeactivatedAt:        user.DeactivatedAt,
			TrafficType:          user.TrafficType,
			TrafficSize:          user.TrafficSize,
			Rx:                   user.Rx,
			Tx:                   user.Tx,
		},
		Usage: UsageResponse{
			DateStart:  dateStart,
			DateEnd:    dateEnd,
			Bandwidths: usage,
		},
	})
}

// DownloadCertificate downloads the customer's own PKCS#12 certificate bundle.
//
// @Summary      Download customer certificate
// @Description  Download customer's own .p12 certificate bundle using ocserv username/password
// @Tags         Customers
// @Accept       json
// @Produce      application/x-pkcs12
// @Param        request body  SummaryData  true "customer username and password (same ocserv account)."
// @Failure      400 {object} request.ErrorResponse
// @Failure      429 {object} middlewares.TooManyRequests
// @Success      200 {file} file "user.p12"
// @Router       /customers/certificate [post]
func (ctl *Controller) DownloadCertificate(c echo.Context) error {
	var data SummaryData

	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	if data.Password == "Secret-Ocpasswd" {
		return ctl.request.BadRequest(c, errors.New("invalid username or password"))
	}

	user, err := ctl.ocservUserRepo.GetByUsername(c.Request().Context(), data.Username)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	if user.Password != data.Password {
		return ctl.request.BadRequest(c, errors.New("invalid username or password"))
	}

	path, err := ctl.ocservUserRepo.CertificatePathByUsername(c.Request().Context(), user.Username)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/x-pkcs12")
	return c.Attachment(path, user.Username+".p12")
}

// DisconnectSessions
//
// @Summary      Disconnect all online sessions of a customer
// @Description  disconnects all online sessions for a customer
// @Tags         Customers
// @Accept       json
// @Produce      json
// @Param        request body  SummaryData  true "customer username and password (same ocserv account)."
// @Failure      400 {object} request.ErrorResponse
// @Failure      429 {object} middlewares.TooManyRequests
// @Success      202  {object} nil
// @Router       /customers/disconnect_sessions [post]
func (ctl *Controller) DisconnectSessions(c echo.Context) error {
	var data SummaryData

	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	if data.Password == "Secret-Ocpasswd" {
		return ctl.request.BadRequest(c, errors.New("invalid username or password"))
	}

	user, err := ctl.ocservUserRepo.GetByUsername(c.Request().Context(), data.Username)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	if user.Password != data.Password {
		return ctl.request.BadRequest(c, errors.New("invalid username or password"))
	}

	_, _ = ctl.occtl.Disconnect(user.Username)

	return c.JSON(http.StatusAccepted, nil)
}
