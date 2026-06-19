package ocserv_user

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/occtl"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
)

type OcservUserController struct {
	userUC    usecase.OcservUserUseCase
	req      *request.Request
	validator *request.Validator
	occtl     *occtl.OcservOcctl
}

func NewOcservUserController(userUC usecase.OcservUserUseCase) *OcservUserController {
	return &OcservUserController{
		userUC:    userUC,
		req:      &request.Request{},
		validator: request.NewValidator(),
		occtl:    occtl.NewOcservOcctl(),
	}
}

// ListUsers godoc
// @Summary List Ocserv Users
// @Description Get all Ocserv users for current admin (superadmin can access all users)
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param page query int false "Page number, starting from 1" minimum(1)
// @Param limit query int false "Number of items per page" minimum(1) maximum(100)
// @Param q query string false "ocserv username q search" minLength(2)
// @Param filter query string false "filter ocserv user by statues" Enums(active, deactivated, locked)
// @Param group query string false "filter ocserv user by group name"
// @Param order_by query string false "Field to order by" Enums(id, created_at)
// @Param sort query string false "Sort order" Enums(asc, desc)
// @Success 200 {object} OcservUsersResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /ocserv/users [get]
func (ctrl *OcservUserController) ListUsers(c *echo.Context) error {
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)

	// Get query params
	page, limit := request.GetPaginationParams(c)
	q := c.QueryParam("q")
	filter := c.QueryParam("filter")
	group := c.QueryParam("group")
	orderBy := c.QueryParam("order_by")
	sort := c.QueryParam("sort")

	// Validate filter
	validFilters := map[string]bool{"active": true, "deactivated": true, "locked": true}
	if filter != "" && !validFilters[filter] {
		filter = ""
	}

	// Get users from use case
	users, total, err := ctrl.userUC.ListUsersPaginated(adminID, role, page, limit, q, filter, group, orderBy, sort)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}

	// Build response
	response := OcservUsersResponse{
		Meta:   request.NewPagination(total, page, limit),
		Result: users,
	}

	return c.JSON(http.StatusOK, response)
}

// GetOnlineSessions godoc
// @Summary Get Online Sessions for Specific Usernames
// @Description Get list of online sessions filtered by provided usernames
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body GetOnlineSessionsRequest true "List of usernames to filter online sessions"
// @Success 200 {array} models.OnlineUserSession
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /ocserv/users/online-sessions [post]
func (ctrl *OcservUserController) GetOnlineSessions(c *echo.Context) error {
	var req GetOnlineSessionsRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	// Create a map for quick lookup
	usernameMap := make(map[string]bool)
	for _, u := range req.Usernames {
		usernameMap[u] = true
	}

	// Get all online sessions
	sessionsPtr, err := ctrl.occtl.OnlineSessions()
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}

	// Filter sessions
	var filteredSessions []models.OnlineUserSession
	if sessionsPtr != nil {
		for _, session := range *sessionsPtr {
			if usernameMap[session.Username] {
				filteredSessions = append(filteredSessions, session)
			}
		}
	}

	return c.JSON(http.StatusOK, filteredSessions)
}

// GetUser godoc
// @Summary Get Ocserv User
// @Description Get Ocserv user by ID
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "User ID"
// @Success 200 {object} models.OcservUser
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /ocserv/users/{id} [get]
func (ctrl *OcservUserController) GetUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid user ID")
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	user, err := ctrl.userUC.GetUser(uint(id), adminID, role)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	return c.JSON(http.StatusOK, user)
}

// CreateUser godoc
// @Summary Create Ocserv User
// @Description Create new Ocserv user
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body CreateOcservUserRequest true "Create user data"
// @Success 201 {object} models.OcservUser
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 409 {object} request.ErrorResponse
// @Router /ocserv/users [post]
func (ctrl *OcservUserController) CreateUser(c *echo.Context) error {
	var req CreateOcservUserRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}
	adminID := c.Get("id").(uint)

	var expireAt *time.Time
	if req.Unlimited {
		expireAt = nil
	} else if req.ExpireAt != "" {
		parsedTime, err := time.Parse("2006-01-02", req.ExpireAt)
		if err == nil {
			expireAt = &parsedTime
		}
	}

	user, err := ctrl.userUC.CreateUser(
		req.Username,
		req.Password,
		req.Group,
		req.TrafficType,
		req.TrafficSize,
		req.Description,
		req.Config,
		adminID,
		expireAt,
	)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	return c.JSON(http.StatusCreated, user)
}

// UpdateUser godoc
// @Summary Update Ocserv User
// @Description Update Ocserv user by ID (superadmin can update any user)
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "User ID"
// @Param request body UpdateOcservUserRequest true "Update user data"
// @Success 200 {object} models.OcservUser
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /ocserv/users/{id} [post]
func (ctrl *OcservUserController) UpdateUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid user ID")
	}
	var req UpdateOcservUserRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)

	var expireAt *time.Time
	if req.Unlimited {
		expireAt = nil
	} else if req.ExpireAt != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.ExpireAt)
		if err == nil {
			expireAt = &parsedTime
		}
	}

	user, err := ctrl.userUC.UpdateUser(
		uint(id),
		adminID,
		role,
		req.Group,
		req.Password,
		expireAt,
		req.Unlimited,
		req.TrafficType,
		req.TrafficSize,
		req.Description,
		req.Config,
	)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	return c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary Delete Ocserv User
// @Description Delete Ocserv user by ID (superadmin can delete any user)
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "User ID"
// @Success 204 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /ocserv/users/{id} [delete]
func (ctrl *OcservUserController) DeleteUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid user ID")
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	if err := ctrl.userUC.DeleteUser(uint(id), adminID, role); err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// LockUser godoc
// @Summary Lock Ocserv User
// @Description Lock Ocserv user by ID (superadmin can lock any user)
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "User ID"
// @Success 200 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /ocserv/users/{id}/lock [post]
func (ctrl *OcservUserController) LockUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid user ID")
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	if err := ctrl.userUC.LockUser(uint(id), adminID, role); err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// UnlockUser godoc
// @Summary Unlock Ocserv User
// @Description Unlock Ocserv user by ID (superadmin can unlock any user)
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "User ID"
// @Success 200 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /ocserv/users/{id}/unlock [post]
func (ctrl *OcservUserController) UnlockUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid user ID")
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	if err := ctrl.userUC.UnlockUser(uint(id), adminID, role); err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// DisconnectUser godoc
// @Summary Disconnect Ocserv User (All Sessions)
// @Description Disconnect all sessions for an Ocserv user by username
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param username path string true "Username"
// @Success 200 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /ocserv/users/{username}/disconnect [post]
func (ctrl *OcservUserController) DisconnectUser(c *echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return ctrl.req.BadRequest(c, nil, "username is required")
	}
	_, err := ctrl.occtl.DisconnectUser(username)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// DisconnectSession godoc
// @Summary Disconnect Ocserv User Session by ID
// @Description Disconnect a specific Ocserv user session by session ID
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path string true "Session ID"
// @Success 200 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /ocserv/users/{id}/disconnect-by-id [post]
func (ctrl *OcservUserController) DisconnectSession(c *echo.Context) error {
	sessionID := c.Param("id")
	if sessionID == "" {
		return ctrl.req.BadRequest(c, nil, "session ID is required")
	}
	_, err := ctrl.occtl.DisconnectSession(sessionID)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// TerminateUser godoc
// @Summary Terminate Ocserv User (All Sessions)
// @Description Terminate all sessions for an Ocserv user by username
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param username path string true "Username"
// @Success 200 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /ocserv/users/{username}/terminate [post]
func (ctrl *OcservUserController) TerminateUser(c *echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return ctrl.req.BadRequest(c, nil, "username is required")
	}
	_, err := ctrl.occtl.TerminateUser(username)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// TerminateSession godoc
// @Summary Terminate Ocserv User Session by ID
// @Description Terminate a specific Ocserv user session by session ID
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path string true "Session ID"
// @Success 200 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /ocserv/users/{id}/terminate-by-id [post]
func (ctrl *OcservUserController) TerminateSession(c *echo.Context) error {
	sessionID := c.Param("id")
	if sessionID == "" {
		return ctrl.req.BadRequest(c, nil, "session ID is required")
	}
	_, err := ctrl.occtl.TerminateSession(sessionID)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// SessionLogs godoc
// @Summary Ocserv User Session Logs
// @Description Get Ocserv user session logs
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "User ID"
// @Param page query int false "Page number, starting from 1" minimum(1)
// @Param limit query int false "Number of items per page" minimum(1) maximum(100)
// @Param order_by query string false "Field to order by" Enums(id, created_at)
// @Param sort query string false "Sort order" Enums(asc, desc)
// @Param date_start query string false "Start date (YYYY-MM-DD)"
// @Param date_end query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} SessionLogsResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /ocserv/users/{id}/session-logs [get]
func (ctrl *OcservUserController) SessionLogs(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid user ID")
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	user, err := ctrl.userUC.GetUser(uint(id), adminID, role)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	page, limit := request.GetPaginationParams(c)
	orderBy := c.QueryParam("order_by")
	sort := c.QueryParam("sort")

	var startDate, endDate *time.Time
	dStart := c.QueryParam("date_start")
	if dStart != "" {
		parsed, err := time.Parse("2006-01-02", dStart)
		if err == nil {
			startDate = &parsed
		}
	}
	dEnd := c.QueryParam("date_end")
	if dEnd != "" {
		parsed, err := time.Parse("2006-01-02", dEnd)
		if err == nil {
			e := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &e
		}
	}

	logs, total, err := ctrl.userUC.UserSessionLogs(user.Username, page, limit, orderBy, sort, startDate, endDate)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	response := SessionLogsResponse{
		Meta:   request.NewPagination(total, page, limit),
		Result: logs,
	}
	return c.JSON(http.StatusOK, response)
}

// Statistics godoc
// @Summary Ocserv User Statistics
// @Description Get Ocserv user statistics
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "User ID"
// @Param date_start query string false "Start date (YYYY-MM-DD)"
// @Param date_end query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} StatisticsResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /ocserv/users/{id}/statistics [get]
func (ctrl *OcservUserController) Statistics(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid user ID")
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)

	var startDate, endDate *time.Time
	dStart := c.QueryParam("date_start")
	if dStart != "" {
		parsed, err := time.Parse("2006-01-02", dStart)
		if err == nil {
			startDate = &parsed
		}
	}
	dEnd := c.QueryParam("date_end")
	if dEnd != "" {
		parsed, err := time.Parse("2006-01-02", dEnd)
		if err == nil {
			e := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &e
		}
	}

	stats, err := ctrl.userUC.UserStatistics(uint(id), adminID, role, startDate, endDate)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	var totalRx, totalTx float64
	for _, s := range stats {
		totalRx += s.Rx
		totalTx += s.Tx
	}
	response := StatisticsResponse{
		Statistics: stats,
		TotalBandwidths: TotalBandwidths{
			Rx: totalRx,
			Tx: totalTx,
		},
	}
	return c.JSON(http.StatusOK, response)
}

// ActivateExpired godoc
// @Summary Activate Expired Ocserv User
// @Description Activate an expired Ocserv user
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "User ID"
// @Param request body ActivateUserData true "Activate user data"
// @Success 200 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /ocserv/users/{id}/activate [post]
func (ctrl *OcservUserController) ActivateExpired(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid user ID")
	}
	var req ActivateUserData
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	var expireAt *time.Time
	if req.ExpireAt != nil {
		parsed, err := time.Parse("2006-01-02", *req.ExpireAt)
		if err == nil {
			expireAt = &parsed
		}
	}
	if err := ctrl.userUC.ActivateExpired(uint(id), adminID, role, expireAt); err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// CreateCertificate godoc
// @Summary Create certificate for ocserv user
// @Description Create certificate for an existing ocserv user
// @Tags Ocserv Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "User ID"
// @Success 200 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /ocserv/users/{id}/certificate [post]
func (ctrl *OcservUserController) CreateCertificate(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid user ID")
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	if err := ctrl.userUC.CreateCertificate(uint(id), adminID, role); err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// DownloadCertificate godoc
// @Summary Download ocserv user certificate
// @Description Download the user's PKCS#12 certificate bundle
// @Tags Ocserv Users
// @Produce application/x-pkcs12
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "User ID"
// @Success 200 {file} file "user.p12"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /ocserv/users/{id}/certificate [get]
func (ctrl *OcservUserController) DownloadCertificate(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid user ID")
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	username, path, err := ctrl.userUC.DownloadCertificate(uint(id), adminID, role)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/x-pkcs12")
	return c.Attachment(path, username+".p12")
}
