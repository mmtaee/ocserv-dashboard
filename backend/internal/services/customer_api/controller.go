package customerapi

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	customerusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/customer_api"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type CustomerController struct {
	request request.CustomRequestInterface
	usecase *customerusecase.Usecase
}

func newCustomerController(usecase *customerusecase.Usecase) *CustomerController {
	return &CustomerController{request: request.NewCustomRequest(), usecase: usecase}
}

// Login authenticates an Ocserv user and returns a seven-hour JWT.
// @Summary Customer login
// @Tags Customers
// @Accept json
// @Produce json
// @Param request body LoginData true "Ocserv credentials"
// @Failure 400 {object} request.ErrorResponse
// @Success 200 {object} LoginResponse
// @Router /customers/login [post]
func (ctl *CustomerController) Login(c *echo.Context) error {
	var input LoginData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.usecase.Login(c.Request().Context(), input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// @Summary Customer account summary
// @Tags Customers
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 401 {object} request.ErrorResponse
// @Success 200 {object} SummaryResponse
// @Router /customers/summary [get]
func (ctl *CustomerController) Summary(c *echo.Context) error {
	username, err := customerUsername(c)
	if err != nil {
		return err
	}
	result, err := ctl.usecase.Summary(c.Request().Context(), username)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// @Summary Customer online sessions
// @Tags Customers
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 401 {object} request.ErrorResponse
// @Success 200 {array} models.OnlineUserSession
// @Router /customers/sessions [get]
func (ctl *CustomerController) Sessions(c *echo.Context) error {
	username, err := customerUsername(c)
	if err != nil {
		return err
	}
	result, err := ctl.usecase.Sessions(c.Request().Context(), username)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// @Summary Download customer certificate
// @Tags Customers
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 401 {object} request.ErrorResponse
// @Success 200 {file} binary
// @Router /customers/certificate [get]
func (ctl *CustomerController) DownloadCertificate(c *echo.Context) error {
	username, err := customerUsername(c)
	if err != nil {
		return err
	}
	path, username, err := ctl.usecase.CertificatePath(c.Request().Context(), username)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/x-pkcs12")
	return c.Attachment(path, username+".p12")
}

// @Summary Disconnect all customer sessions
// @Tags Customers
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 401 {object} request.ErrorResponse
// @Success 202
// @Router /customers/disconnect_sessions [post]
func (ctl *CustomerController) DisconnectSessions(c *echo.Context) error {
	username, err := customerUsername(c)
	if err != nil {
		return err
	}
	if err := ctl.usecase.Disconnect(c.Request().Context(), username); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusAccepted, nil)
}

// @Summary Terminate all customer sessions
// @Tags Customers
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 401 {object} request.ErrorResponse
// @Success 202
// @Router /customers/terminate_sessions [post]
func (ctl *CustomerController) TerminateSessions(c *echo.Context) error {
	username, err := customerUsername(c)
	if err != nil {
		return err
	}
	if err := ctl.usecase.Terminate(c.Request().Context(), username); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusAccepted, nil)
}

// @Summary Change customer Ocserv password
// @Tags Customers
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body ChangePasswordInput true "New password"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Success 204
// @Router /customers/password [post]
func (ctl *CustomerController) ChangePassword(c *echo.Context) error {
	username, err := customerUsername(c)
	if err != nil {
		return err
	}
	var input ChangePasswordInput
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.usecase.ChangePassword(c.Request().Context(), username, input.Password); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// @Summary Customer traffic statistics
// @Tags Customers
// @Param Authorization header string true "Bearer TOKEN"
// @Param date_start query string false "Start date" Format(date)
// @Param date_end query string false "End date" Format(date)
// @Failure 401 {object} request.ErrorResponse
// @Success 200 {array} models.DailyTraffic
// @Router /customers/stats [get]
func (ctl *CustomerController) Stats(c *echo.Context) error {
	username, err := customerUsername(c)
	if err != nil {
		return err
	}
	var input DateRange
	if err := c.Bind(&input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.usecase.Statistics(c.Request().Context(), username, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// @Summary Customer activity list
// @Tags Customers
// @Param Authorization header string true "Bearer TOKEN"
// @Param page query int false "Page number" minimum(1)
// @Param size query int false "Page size" minimum(1) maximum(100)
// @Param date_start query string false "Start date" Format(date)
// @Param date_end query string false "End date" Format(date)
// @Failure 401 {object} request.ErrorResponse
// @Success 200 {object} ActivitiesResponse
// @Router /customers/activities [get]
func (ctl *CustomerController) Activities(c *echo.Context) error {
	username, err := customerUsername(c)
	if err != nil {
		return err
	}
	var input DateRange
	if err := c.Bind(&input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	pagination := ctl.request.Pagination(c)
	result, err := ctl.usecase.Activities(c.Request().Context(), username, pagination, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, ActivitiesResponse{Meta: request.Meta{Page: pagination.Page, PageSize: pagination.PageSize, TotalRecords: result.Total}, Result: result.Logs})
}

// @Summary Customer bandwidth
// @Tags Customers
// @Param Authorization header string true "Bearer TOKEN"
// @Param date_start query string false "Start date" Format(date)
// @Param date_end query string false "End date" Format(date)
// @Failure 401 {object} request.ErrorResponse
// @Success 200 {object} repository.TotalBandwidths
// @Router /customers/bandwidth [get]
func (ctl *CustomerController) Bandwidth(c *echo.Context) error {
	username, err := customerUsername(c)
	if err != nil {
		return err
	}
	var input DateRange
	if err := c.Bind(&input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.usecase.Bandwidth(c.Request().Context(), username, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// @Summary Build Cisco Secure Client setup data
// @Tags Customers
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 401 {object} request.ErrorResponse
// @Success 200 {object} customerusecase.CiscoSetup
// @Router /customers/setup/cisco [get]
func (ctl *CustomerController) CiscoSetup(c *echo.Context) error {
	username, err := customerUsername(c)
	if err != nil {
		return err
	}
	result, err := ctl.usecase.CiscoSetup(c.Request().Context(), username, publicAPIBaseURL(c))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// @Summary Download Cisco setup certificate
// @Tags Customers
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 401 {object} request.ErrorResponse
// @Success 200 {file} binary
// @Router /customers/setup/cisco/certificate [get]
func (ctl *CustomerController) DownloadCiscoSetupCertificate(c *echo.Context) error {
	username, err := customerUsername(c)
	if err != nil {
		return err
	}
	path, username, err := ctl.usecase.CiscoCertificatePath(c.Request().Context(), username)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/x-pkcs12")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	return c.Attachment(path, username+".p12")
}

func customerUsername(c *echo.Context) (string, error) {
	principal, err := middlewares.Principal(c)
	if err != nil {
		return "", middlewares.UnauthorizedError(c, "invalid token")
	}
	return principal.Username, nil
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
