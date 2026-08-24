package ocserv_user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/authz"
	userusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/users"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	users   *userusecase.Usecase
}

func New(usecase *userusecase.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), users: usecase}
}

// Users lists OCServ users. @Summary List of Ocserv Users
// @Tags Ocserv(Users)
// @Produce json
// @Router /ocserv/users [get]
func (ctl *Controller) Users(c *echo.Context) error {
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	pagination := ctl.request.Pagination(c)
	result, err := ctl.users.List(c.Request().Context(), userusecase.ListOptions{
		Pagination: pagination, Principal: principal, Query: c.QueryParam("q"), Filter: c.QueryParam("filter"), Group: c.QueryParam("group"),
	})
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, OcservUsersResponse{Meta: request.Meta{Page: pagination.Page, TotalRecords: result.Total, PageSize: pagination.PageSize}, Result: result.Users})
}

// User returns an OCServ user. @Summary Ocserv user detail
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
// @Router /ocserv/users/{id} [get]
func (ctl *Controller) User(c *echo.Context) error {
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.users.User(c.Request().Context(), principal, id)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Create creates an OCServ user. @Summary Ocserv User creation
// @Tags Ocserv(Users)
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body CreateOcservUserData true "OCServ user and expiry configuration"
// @Success 201 {object} models.OcservUser
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Router /ocserv/users [post]
func (ctl *Controller) Create(c *echo.Context) error {
	var input CreateOcservUserData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	result, err := ctl.users.CreateUser(c.Request().Context(), principal, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusCreated, result)
}

// Update updates an OCServ user. @Summary Ocserv User update
// @Tags Ocserv(Users)
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Ocserv User ID"
// @Param request body UpdateOcservUserData true "OCServ user and expiry changes"
// @Success 200 {object} models.OcservUser
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Failure 403 {object} middlewares.PermissionDenied
// @Router /ocserv/users/{id} [patch]
func (ctl *Controller) Update(c *echo.Context) error {
	var input UpdateOcservUserData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	result, err := ctl.users.UpdateUser(c.Request().Context(), principal, id, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// BulkUpdate updates multiple OCServ users atomically.
// @Summary Bulk update OCServ users
// @Tags Ocserv(Users)
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body BulkUpdateRequest true "Bulk user updates"
// @Success 200 {object} BulkUsersResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Failure 403 {object} middlewares.PermissionDenied
// @Router /ocserv/users/bulk [patch]
func (ctl *Controller) BulkUpdate(c *echo.Context) error {
	var input BulkUpdateRequest
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	result, err := ctl.users.BulkUpdate(c.Request().Context(), principal, input)
	if err != nil {
		return ctl.respondError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// BulkDelete deletes multiple OCServ users atomically.
// @Summary Bulk delete OCServ users
// @Tags Ocserv(Users)
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body BulkIDsRequest true "OCServ user IDs"
// @Success 200 {object} BulkDeleteResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Failure 403 {object} middlewares.PermissionDenied
// @Router /ocserv/users/bulk [delete]
func (ctl *Controller) BulkDelete(c *echo.Context) error {
	var input BulkIDsRequest
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	result, err := ctl.users.BulkDelete(c.Request().Context(), principal, input)
	if err != nil {
		return ctl.respondError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// BulkSetEnabled enables or disables multiple OCServ users atomically.
// @Summary Bulk enable or disable OCServ users
// @Tags Ocserv(Users)
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body BulkStatusRequest true "OCServ user IDs and enabled state"
// @Success 200 {object} BulkUsersResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Failure 403 {object} middlewares.PermissionDenied
// @Router /ocserv/users/bulk/status [patch]
func (ctl *Controller) BulkSetEnabled(c *echo.Context) error {
	var input BulkStatusRequest
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	result, err := ctl.users.BulkSetEnabled(c.Request().Context(), principal, input)
	if err != nil {
		return ctl.respondError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// BulkSetGroup assigns a group, or removes assignment when group is empty.
// @Summary Bulk assign or remove an OCServ group
// @Tags Ocserv(Users)
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body BulkGroupRequest true "OCServ user IDs and group; empty group removes assignment"
// @Success 200 {object} BulkUsersResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Failure 403 {object} middlewares.PermissionDenied
// @Router /ocserv/users/bulk/group [patch]
func (ctl *Controller) BulkSetGroup(c *echo.Context) error {
	var input BulkGroupRequest
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	result, err := ctl.users.BulkSetGroup(c.Request().Context(), principal, input)
	if err != nil {
		return ctl.respondError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Delete deletes an OCServ user. @Summary Ocserv User delete
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
// @Router /ocserv/users/{id} [delete]
func (ctl *Controller) Delete(c *echo.Context) error {
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.users.DeleteUser(c.Request().Context(), principal, id); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

// Lock locks an OCServ user. @Summary Ocserv User locking
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
// @Router /ocserv/users/{id}/lock [post]
func (ctl *Controller) Lock(c *echo.Context) error {
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.users.LockUser(c.Request().Context(), principal, id); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// UnLock unlocks an OCServ user. @Summary Ocserv User unlocking
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
// @Router /ocserv/users/{id}/unlock [post]
func (ctl *Controller) UnLock(c *echo.Context) error {
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.users.UnlockUser(c.Request().Context(), principal, id); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// Statistics returns OCServ user traffic statistics. @Summary Ocserv User Statistics
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
// @Router /ocserv/users/{id}/statistics [get]
func (ctl *Controller) Statistics(c *echo.Context) error {
	var input StatisticsData
	if err := c.Bind(&input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	result, err := ctl.users.Statistics(c.Request().Context(), principal, id, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// OcpasswdUsers lists users from ocpasswd. @Summary Ocserv Users from ocpasswd file
// @Tags Ocserv(Ocpasswd)
// @Router /ocserv/users/ocpasswd [get]
func (ctl *Controller) OcpasswdUsers(c *echo.Context) error {
	pagination := ctl.request.Pagination(c)
	result, err := ctl.users.ListOcpasswd(c.Request().Context(), pagination)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, OcservUsersSyncResponse{Meta: request.Meta{Page: pagination.Page, TotalRecords: int64(result.Total), PageSize: pagination.PageSize}, Result: result.Users})
}

// SyncToDB imports ocpasswd users. @Summary Ocserv Users from ocpasswd file to db
// @Tags Ocserv(Ocpasswd)
// @Router /ocserv/users/ocpasswd/sync [post]
func (ctl *Controller) SyncToDB(c *echo.Context) error {
	var input SyncOcpasswdRequest
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	result, err := ctl.users.SyncOcpasswd(c.Request().Context(), principal, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// ActivateExpired restores an expired account. @Summary Restore and activate expired Ocserv User accounts
// @Tags Ocserv(Users)
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Ocserv User ID"
// @Param request body ActivateUserData true "New expiry configuration; omitted values reset to unlimited"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Failure 403 {object} middlewares.PermissionDenied
// @Router /ocserv/users/{id}/activate [post]
func (ctl *Controller) ActivateExpired(c *echo.Context) error {
	var input ActivateUserData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	if err := ctl.users.Activate(c.Request().Context(), principal, id, input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// CreateCertificate creates an OCServ user certificate. @Summary Create certificate for ocserv user
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
// @Router /ocserv/users/{id}/certificate [post]
func (ctl *Controller) CreateCertificate(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	if err := ctl.users.CreateUserCertificate(c.Request().Context(), principal, id); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// DownloadCertificate downloads an OCServ user certificate. @Summary Download ocserv user certificate
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
// @Router /ocserv/users/{id}/certificate [get]
func (ctl *Controller) DownloadCertificate(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	username, path, err := ctl.users.UserCertificate(c.Request().Context(), principal, id)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/x-pkcs12")
	return c.Attachment(path, username+".p12")
}

// SessionLogs returns an OCServ user's session logs. @Summary Ocserv User session logs
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
// @Router /ocserv/users/{id}/session_logs [get]
func (ctl *Controller) SessionLogs(c *echo.Context) error {
	var input SessionLogsData
	if err := c.Bind(&input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	pagination := ctl.request.Pagination(c)
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	result, err := ctl.users.SessionLogs(c.Request().Context(), principal, id, pagination, input)
	if errors.Is(err, userusecase.ErrUserNotFound) {
		return c.JSON(http.StatusNotFound, nil)
	}
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, SessionLogsResponse{Meta: request.Meta{Page: pagination.Page, TotalRecords: result.Total, PageSize: pagination.PageSize}, Result: result.Logs})
}

// Disconnect disconnects all sessions. @Summary Disconnect Ocserv User
// @Tags Ocserv(Users)
// @Router /ocserv/users/{username}/disconnect [post]
func (ctl *Controller) Disconnect(c *echo.Context) error {
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	if err := ctl.users.DisconnectUser(c.Request().Context(), principal, c.Param("username")); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// DisconnectSessionById disconnects one session. @Summary Disconnect Ocserv User Session BY ID
// @Tags Ocserv(Users)
// @Router /ocserv/users/{id}/disconnect_by_id [post]
func (ctl *Controller) DisconnectSessionById(c *echo.Context) error {
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	if err := ctl.users.DisconnectSession(c.Request().Context(), principal, c.Param("id")); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// Terminate terminates all sessions. @Summary Terminate Ocserv User
// @Tags Ocserv(Users)
// @Router /ocserv/users/{username}/terminate [post]
func (ctl *Controller) Terminate(c *echo.Context) error {
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	if err := ctl.users.TerminateUser(c.Request().Context(), principal, c.Param("username")); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// TerminateSessionById terminates one session. @Summary Terminate Ocserv User Session BY ID
// @Tags Ocserv(Users)
// @Router /ocserv/users/{id}/terminate_by_id [post]
func (ctl *Controller) TerminateSessionById(c *echo.Context) error {
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.respondError(c, err)
	}
	if err := ctl.users.TerminateSession(c.Request().Context(), principal, c.Param("id")); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

func (ctl *Controller) respondError(c *echo.Context, err error) error {
	if errors.Is(err, authz.ErrForbidden) {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}
	return ctl.request.BadRequest(c, err)
}

func parseID(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid user id")
	}
	return uint(id), nil
}
