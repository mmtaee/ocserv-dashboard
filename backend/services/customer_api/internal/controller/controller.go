package customer

import (
	"net/http"

	"github.com/labstack/echo/v5"
	customerusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/customer"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	usecase *customerusecase.Usecase
}

func New(usecase *customerusecase.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), usecase: usecase}
}

func (ctl *Controller) Summary(c *echo.Context) error {
	var data SummaryData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.usecase.Summary(c.Request().Context(), data)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (ctl *Controller) DownloadCertificate(c *echo.Context) error {
	var data SummaryData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	path, username, err := ctl.usecase.CertificatePath(c.Request().Context(), data)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/x-pkcs12")
	return c.Attachment(path, username+".p12")
}

func (ctl *Controller) DisconnectSessions(c *echo.Context) error {
	var data SummaryData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.usecase.Disconnect(c.Request().Context(), data); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusAccepted, nil)
}
