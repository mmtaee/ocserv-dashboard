package users

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	userusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/users"
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
	owner := ""
	if isAdmin, ok := c.Get("isAdmin").(bool); !ok || !isAdmin {
		username, ok := c.Get("username").(string)
		if !ok || username == "" {
			return ctl.request.BadRequest(c, errors.New("invalid user id"))
		}
		owner = username
	}
	pagination := ctl.request.Pagination(c)
	result, err := ctl.users.List(c.Request().Context(), userusecase.ListOptions{
		Pagination: pagination, Owner: owner, Query: c.QueryParam("q"), Filter: c.QueryParam("filter"), Group: c.QueryParam("group"),
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
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.users.User(c.Request().Context(), id)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Create creates an OCServ user. @Summary Ocserv User creation
// @Tags Ocserv(Users)
// @Router /ocserv/users [post]
func (ctl *Controller) Create(c *echo.Context) error {
	var input CreateOcservUserData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	owner, _ := c.Get("username").(string)
	result, err := ctl.users.CreateUser(c.Request().Context(), owner, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusCreated, result)
}

// Update updates an OCServ user. @Summary Ocserv User update
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
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
	result, err := ctl.users.UpdateUser(c.Request().Context(), id, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Delete deletes an OCServ user. @Summary Ocserv User delete
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
// @Router /ocserv/users/{id} [delete]
func (ctl *Controller) Delete(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.users.DeleteUser(c.Request().Context(), id); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

// Lock locks an OCServ user. @Summary Ocserv User locking
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
// @Router /ocserv/users/{id}/lock [post]
func (ctl *Controller) Lock(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.users.LockUser(c.Request().Context(), id); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// UnLock unlocks an OCServ user. @Summary Ocserv User unlocking
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
// @Router /ocserv/users/{id}/unlock [post]
func (ctl *Controller) UnLock(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.users.UnlockUser(c.Request().Context(), id); err != nil {
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
	result, err := ctl.users.Statistics(c.Request().Context(), id, input)
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
	owner, _ := c.Get("username").(string)
	result, err := ctl.users.SyncOcpasswd(c.Request().Context(), owner, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// ActivateExpired restores an expired account. @Summary Restore and activate expired Ocserv User accounts
// @Tags Ocserv(Users)
// @Param id path int true "Ocserv User ID"
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
	if err := ctl.users.Activate(c.Request().Context(), id, input); err != nil {
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
	if err := ctl.users.CreateUserCertificate(c.Request().Context(), id); err != nil {
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
	username, path, err := ctl.users.UserCertificate(c.Request().Context(), id)
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
	result, err := ctl.users.SessionLogs(c.Request().Context(), id, pagination, input)
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
	if err := ctl.users.DisconnectUser(c.Param("username")); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// DisconnectSessionById disconnects one session. @Summary Disconnect Ocserv User Session BY ID
// @Tags Ocserv(Users)
// @Router /ocserv/users/{id}/disconnect_by_id [post]
func (ctl *Controller) DisconnectSessionById(c *echo.Context) error {
	if err := ctl.users.DisconnectSession(c.Param("id")); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// Terminate terminates all sessions. @Summary Terminate Ocserv User
// @Tags Ocserv(Users)
// @Router /ocserv/users/{username}/terminate [post]
func (ctl *Controller) Terminate(c *echo.Context) error {
	if err := ctl.users.TerminateUser(c.Param("username")); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// TerminateSessionById terminates one session. @Summary Terminate Ocserv User Session BY ID
// @Tags Ocserv(Users)
// @Router /ocserv/users/{id}/terminate_by_id [post]
func (ctl *Controller) TerminateSessionById(c *echo.Context) error {
	if err := ctl.users.TerminateSession(c.Param("id")); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

func parseID(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid user id")
	}
	return uint(id), nil
}
