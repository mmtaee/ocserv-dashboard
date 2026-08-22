package ciscosetup

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	customerusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/customer_api"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	usecase *customerusecase.Usecase
}

func New(usecase *customerusecase.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), usecase: usecase}
}

func (ctl *Controller) CiscoSetup(c *echo.Context) error {
	var data customerusecase.Credentials
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.usecase.CiscoSetup(c.Request().Context(), data, publicAPIBaseURL(c))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (ctl *Controller) DownloadCiscoSetupCertificate(c *echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return ctl.request.BadRequest(c, errors.New("token is required"))
	}
	path, username, err := ctl.usecase.CiscoCertificatePath(c.Request().Context(), token)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/x-pkcs12")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	return c.Attachment(path, username+".p12")
}

func publicAPIBaseURL(c *echo.Context) string {
	req := c.Request()
	scheme := strings.TrimSpace(req.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		if req.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := strings.TrimSpace(req.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(req.Host)
	}
	return scheme + "://" + host
}
