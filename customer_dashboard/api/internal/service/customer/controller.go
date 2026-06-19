package customer

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/internal/usecase"
)

type Controller struct {
	req        *request.Request
	customerUC usecase.CustomerUseCase
	validator  *request.Validator
}

func NewController(customerUC usecase.CustomerUseCase) *Controller {
	return &Controller{
		req:        &request.Request{},
		customerUC: customerUC,
		validator:  request.NewValidator(),
	}
}

// GetUser returns the ocserv user
// @Summary Get User
// @Description Get the ocserv user details
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} models.OcservUser
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /customer/user [get]
func (ctrl *Controller) GetUser(c *echo.Context) error {
	username := c.Get("username").(string)
	user, err := ctrl.customerUC.GetUser(username)
	if err != nil {
		if err.Error() == "8001" {
			return ctrl.req.ResponseWithCode(c, "8001", nil)
		}
		return ctrl.req.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, user)
}

// IsOnline checks if user has any online sessions
// @Summary Check If User Is Online
// @Description Check if user is currently online
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /customer/online [get]
func (ctrl *Controller) IsOnline(c *echo.Context) error {
	username := c.Get("username").(string)
	isOnline, err := ctrl.customerUC.IsOnline(username)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, map[string]bool{"is_online": isOnline})
}

// OnlineSessions returns the user's online sessions
// @Summary Get Online Sessions
// @Description Get the user's online sessions
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} OnlineSessionsResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /customer/sessions/online [get]
func (ctrl *Controller) OnlineSessions(c *echo.Context) error {
	username := c.Get("username").(string)
	sessions, err := ctrl.customerUC.OnlineSessions(username)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, OnlineSessionsResponse{
		Result: sessions,
	})
}

// TerminateAllSessions terminates all user sessions
// @Summary Terminate All Sessions
// @Description Terminate all user's sessions
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 202 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /customer/sessions/terminate-all [post]
func (ctrl *Controller) TerminateAllSessions(c *echo.Context) error {
	username := c.Get("username").(string)
	if err := ctrl.customerUC.TerminateAllSessions(username); err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	return c.JSON(http.StatusAccepted, nil)
}

// TerminateSession terminates a specific session
// @Summary Terminate Session By ID
// @Description Terminate a specific user session by ID
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path string true "Session ID"
// @Success 202 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /customer/sessions/terminate/{id} [post]
func (ctrl *Controller) TerminateSession(c *echo.Context) error {
	username := c.Get("username").(string)
	sessionID := c.Param("id")
	if err := ctrl.customerUC.TerminateSession(username, sessionID); err != nil {
		if err.Error() == "8003" {
			return ctrl.req.ResponseWithCode(c, "8003", nil)
		}
		return ctrl.req.BadRequest(c, err)
	}
	return c.JSON(http.StatusAccepted, nil)
}

// UpdatePassword updates user password
// @Summary Update Password
// @Description Update user's ocserv password
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body UpdatePasswordRequest true "Password Update Request"
// @Success 200 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /customer/password [post]
func (ctrl *Controller) UpdatePassword(c *echo.Context) error {
	username := c.Get("username").(string)
	var req UpdatePasswordRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	if err := ctrl.customerUC.UpdatePassword(username, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "8004" {
			return ctrl.req.ResponseWithCode(c, "8004", nil)
		}
		return ctrl.req.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, nil)
}

// CreateCertificate creates certificate if it doesn't exist
// @Summary Create Certificate
// @Description Create user certificate if it doesn't exist
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 201 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /customer/certificate/create [post]
func (ctrl *Controller) CreateCertificate(c *echo.Context) error {
	username := c.Get("username").(string)
	if err := ctrl.customerUC.CreateCertificate(username); err != nil {
		if err.Error() == "8002" {
			return ctrl.req.ResponseWithCode(c, "8002", nil)
		}
		return ctrl.req.BadRequest(c, err)
	}
	return c.JSON(http.StatusCreated, nil)
}

// UserStatistics returns user statistics (RX/TX)
// @Summary Get User Statistics
// @Description Get user bandwidth statistics
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request query UserStatisticsRequest true "Statistics Request"
// @Success 200 {object} UserStatisticsResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /customer/statistics [get]
func (ctrl *Controller) UserStatistics(c *echo.Context) error {
	username := c.Get("username").(string)
	var req UserStatisticsRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	var startDate, endDate *time.Time
	if req.DateStart != "" {
		parsed, err := time.Parse("2006-01-02", req.DateStart)
		if err == nil {
			startDate = &parsed
		}
	}
	if req.DateEnd != "" {
		parsed, err := time.Parse("2006-01-02", req.DateEnd)
		if err == nil {
			endDate = &parsed
		}
	}

	stats, err := ctrl.customerUC.UserStatistics(username, startDate, endDate)
	if err != nil {
		if err.Error() == "8001" {
			return ctrl.req.ResponseWithCode(c, "8001", nil)
		}
		return ctrl.req.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, UserStatisticsResponse{
		Bandwidths: stats,
	})
}

// SessionLogs returns user session logs
// @Summary Get Session Logs
// @Description Get user's session logs
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request query SessionLogsRequest true "Session Logs Request"
// @Success 200 {object} SessionLogsResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /customer/session-logs [get]
func (ctrl *Controller) SessionLogs(c *echo.Context) error {
	username := c.Get("username").(string)
	var req SessionLogsRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	var startDate, endDate *time.Time
	if req.DateStart != "" {
		parsed, err := time.Parse("2006-01-02", req.DateStart)
		if err == nil {
			startDate = &parsed
		}
	}
	if req.DateEnd != "" {
		parsed, err := time.Parse("2006-01-02", req.DateEnd)
		if err == nil {
			endDate = &parsed
		}
	}

	logs, total, err := ctrl.customerUC.SessionLogs(username, req.Page, req.Limit, req.OrderBy, req.Sort, startDate, endDate)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	pagination := request.NewPagination(total, req.Page, req.Limit)

	return c.JSON(http.StatusOK, SessionLogsResponse{
		Meta:   pagination,
		Result: logs,
	})
}

// Summary gets customer account summary
// @Summary Customer Summary
// @Description Get customer account summary
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} SummaryResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /customer/summary [get]
func (ctrl *Controller) Summary(c *echo.Context) error {
	username := c.Get("username").(string)
	user, certStatus, stats, err := ctrl.customerUC.Summary(username)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	dateEnd := time.Now()
	firstOfThisMonth := time.Date(dateEnd.Year(), dateEnd.Month(), 1, 0, 0, 0, 0, dateEnd.Location())
	dateStart := firstOfThisMonth.AddDate(0, -1, 0)

	return c.JSON(http.StatusOK, SummaryResponse{
		OcservUser: ModelCustomer{
			Owner:                user.Owner,
			Username:             user.Username,
			IsLocked:             user.IsLocked,
			ExpireAt:             user.ExpireAt,
			DeactivatedAt:        user.DeactivatedAt,
			TrafficType:          user.TrafficType,
			TrafficSize:          int64(user.TrafficSize),
			Rx:                   user.Rx,
			Tx:                   user.Tx,
			CertificateEnabled:   certStatus.Enabled,
			CertificateAvailable: certStatus.Available,
		},
		Usage: UsageResponse{
			DateStart:  dateStart,
			DateEnd:    dateEnd,
			Bandwidths: stats,
		},
	})
}

// DownloadCertificate downloads customer certificate
// @Summary Download Certificate
// @Description Download customer's certificate
// @Tags Customer
// @Produce application/x-pkcs12
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Success 200 {file} file "user.p12"
// @Router /customer/certificate [get]
func (ctrl *Controller) DownloadCertificate(c *echo.Context) error {
	username := c.Get("username").(string)
	path, err := ctrl.customerUC.DownloadCertificate(username)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/x-pkcs12")
	return c.Attachment(path, username+".p12")
}

// DisconnectSessions disconnects all customer sessions
// @Summary Disconnect Customer Sessions
// @Description Disconnect all active customer sessions
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Success 202 {object} nil
// @Router /customer/disconnect-sessions [post]
func (ctrl *Controller) DisconnectSessions(c *echo.Context) error {
	username := c.Get("username").(string)
	if err := ctrl.customerUC.DisconnectSessions(username); err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	return c.JSON(http.StatusAccepted, nil)
}

// CiscoSetup creates Cisco setup URIs
// @Summary Cisco Setup URIs
// @Description Create Cisco Secure Client setup URIs
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} CiscoSetupResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /customer/setup/cisco [get]
func (ctrl *Controller) CiscoSetup(c *echo.Context) error {
	username := c.Get("username").(string)
	setup, err := ctrl.customerUC.CiscoSetup(username, publicAPIBaseURL(c))
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, CiscoSetupResponse{
		CertificateImportURI: setup.CertificateImportURI,
		ConnectionCreateURI:  setup.ConnectionCreateURI,
		CertificatePassword:  setup.CertificatePassword,
		ConnectionName:       setup.ConnectionName,
		ServerAddress:        setup.ServerAddress,
		ServerPort:           setup.ServerPort,
		ExpiresAt:            setup.ExpiresAt,
	})
}

// DownloadCiscoSetupCertificate downloads certificate for Cisco setup
// @Summary Download Cisco Setup Certificate
// @Description Download certificate via Cisco setup token
// @Tags Customer
// @Produce application/x-pkcs12
// @Param token path string true "Setup Token"
// @Failure 400 {object} request.ErrorResponse
// @Success 200 {file} file "user.p12"
// @Router /customer/setup/cisco/certificate/{token} [get]
func (ctrl *Controller) DownloadCiscoSetupCertificate(c *echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return ctrl.req.BadRequest(c, errors.New("token is required"))
	}

	path, username, err := ctrl.customerUC.DownloadCiscoSetupCertificate(token)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
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
		scheme = strings.TrimSpace(req.URL.Scheme)
	}
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
