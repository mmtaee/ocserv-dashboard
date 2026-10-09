package auth

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/auth"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	auth    *auth.Usecase
}

func New(usecase *auth.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), auth: usecase}
}

// Logout revokes the current database-backed session.
// @Summary Logout current session
// @Tags Auth
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 401 {object} request.ErrorResponse
// @Success 204
// @Router /auth/logout [post]
func (ctl *Controller) Logout(c *echo.Context) error {
	principal, err := middlewares.Principal(c)
	if err != nil || principal.SessionID == 0 {
		return middlewares.UnauthorizedError(c, "invalid session")
	}
	if err := ctl.auth.Logout(c.Request().Context(), principal.SessionID); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Sessions lists active dashboard sessions.
// @Summary List active user sessions
// @Tags Auth
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} SessionsResponse
// @Router /auth/sessions [get]
func (ctl *Controller) Sessions(c *echo.Context) error {
	pagination := ctl.request.Pagination(c)
	sessions, total, err := ctl.auth.Sessions(c.Request().Context(), pagination)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, SessionsResponse{
		Meta: request.Meta{Page: pagination.Page, PageSize: pagination.PageSize, TotalRecords: total}, Result: sessions,
	})
}

// Revoke revokes a dashboard session by session ID.
// @Summary Revoke user session
// @Tags Auth
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Session ID"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 204
// @Router /auth/sessions/{id} [delete]
func (ctl *Controller) Revoke(c *echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return ctl.request.BadRequest(c, errors.New("invalid session id"))
	}
	if err := ctl.auth.Revoke(c.Request().Context(), uint(id)); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
