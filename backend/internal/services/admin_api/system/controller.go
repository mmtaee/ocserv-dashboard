package system

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	systemusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/system"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	system  *systemusecase.Usecase
}

func New(usecase *systemusecase.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), system: usecase}
}

// DashboardRelease returns current/latest dashboard releases.
// @Summary Get Dashboard the current and latest release
// @Tags System
// @Produce json
// @Failure 400 {object} request.ErrorResponse
// @Success 200 {object} DashboardRelease
// @Router /system/release [get]
func (ctl *Controller) DashboardRelease(c *echo.Context) error {
	result, err := ctl.system.Release(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// ResetAdminPassword resets an admin password using the application secret.
// @Summary Reset admin password by secret key
// @Tags System(User)
// @Accept json
// @Produce json
// @Param request body ResetAdminPassword true "Reset admin password data"
// @Failure 400 {object} request.ErrorResponse
// @Success 200 {object} ResetPasswordResponse
// @Router /system/user/reset-password [post]
func (ctl *Controller) ResetAdminPassword(c *echo.Context) error {
	var input ResetAdminPassword
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.system.ResetPassword(c.Request().Context(), input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// SystemInit returns public initialization settings.
// @Summary Get panel System init Config
// @Tags System
// @Produce json
// @Failure 400 {object} request.ErrorResponse
// @Success 200 {object} GetSystemInitResponse
// @Router /system/init [get]
func (ctl *Controller) SystemInit(c *echo.Context) error {
	result, err := ctl.system.Init(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// System returns administrative system settings.
// @Summary Get panel System Config
// @Tags System
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Success 200 {object} GetSystemResponse
// @Router /system [get]
func (ctl *Controller) System(c *echo.Context) error {
	result, err := ctl.system.Settings(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// SystemUpdate patches administrative system settings.
// @Summary Update panel System Config
// @Tags System
// @Accept json
// @Produce json
// @Param request body PatchSystemUpdateData true "update system config data"
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Success 200 {object} GetSystemResponse
// @Router /system [patch]
func (ctl *Controller) SystemUpdate(c *echo.Context) error {
	var input PatchSystemUpdateData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	userID, err := currentUserID(c)
	if err != nil {
		return middlewares.UnauthorizedError(c, err.Error())
	}
	result, err := ctl.system.Update(c.Request().Context(), userID, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Login authenticates an administrative user.
// @Summary Admin users login
// @Tags System(Users)
// @Accept json
// @Produce json
// @Param request body LoginData true "login data"
// @Failure 400 {object} request.ErrorResponse
// @Success 200 {object} UserLoginResponse
// @Router /system/users/login [post]
func (ctl *Controller) Login(c *echo.Context) error {
	var input LoginData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.system.Login(c.Request().Context(), input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// CreateUser creates a non-admin panel user.
// @Summary Create user
// @Tags System(Users)
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body CreateUserData true "create user data"
// @Failure 400 {object} request.ErrorResponse
// @Success 201 {object} models.User
// @Router /system/users [post]
func (ctl *Controller) CreateUser(c *echo.Context) error {
	var input CreateUserData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.system.CreateUser(c.Request().Context(), input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusCreated, result)
}

// Users lists non-admin panel users.
// @Summary List of Admin or simple users
// @Tags System(Users)
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} UsersResponse
// @Router /system/users [get]
func (ctl *Controller) Users(c *echo.Context) error {
	pagination := ctl.request.Pagination(c)
	users, total, err := ctl.system.Users(c.Request().Context(), pagination)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, UsersResponse{Meta: request.Meta{Page: pagination.Page, PageSize: pagination.PageSize, TotalRecords: total}, Result: users})
}

// ChangeUserPasswordByAdmin changes a target user's password.
// @Summary Change user password by admin
// @Tags System(Users)
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body ChangeUserPassword true "user new password"
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} nil
// @Router /system/users/{id}/password [post]
func (ctl *Controller) ChangeUserPasswordByAdmin(c *echo.Context) error {
	var input ChangeUserPassword
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.system.ChangeUserPassword(c.Request().Context(), id, input.Password); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// DeleteUser deletes a non-admin panel user.
// @Summary Delete simple user
// @Tags System(Users)
// @Produce json
// @Param id path int true "User ID"
// @Param Authorization header string true "Bearer TOKEN"
// @Success 204 {object} nil
// @Router /system/users/{id} [delete]
func (ctl *Controller) DeleteUser(c *echo.Context) error {
	actorID, err := currentUserID(c)
	if err != nil {
		return middlewares.UnauthorizedError(c, err.Error())
	}
	targetID, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.system.DeleteUser(c.Request().Context(), actorID, targetID); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

// ChangePasswordBySelf changes the current user's password.
// @Summary Change user password by self
// @Tags System(Users)
// @Accept json
// @Produce json
// @Param request body ChangeUserPasswordBySelf true "user new password"
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} nil
// @Router /system/users/password [post]
func (ctl *Controller) ChangePasswordBySelf(c *echo.Context) error {
	var input ChangeUserPasswordBySelf
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	userID, err := currentUserID(c)
	if err != nil {
		return middlewares.UnauthorizedError(c, err.Error())
	}
	if err := ctl.system.ChangeOwnPassword(c.Request().Context(), userID, input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// Profile returns the current panel user.
// @Summary Get User Profile
// @Tags System(Users)
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} models.User
// @Router /system/users/profile [get]
func (ctl *Controller) Profile(c *echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return middlewares.UnauthorizedError(c, err.Error())
	}
	result, err := ctl.system.Profile(c.Request().Context(), userID)
	if err != nil {
		return middlewares.UnauthorizedError(c, "user not found")
	}
	return c.JSON(http.StatusOK, result)
}

func currentUserID(c *echo.Context) (uint, error) {
	principal, err := middlewares.Principal(c)
	if err != nil {
		return 0, errors.New("invalid user id")
	}
	return principal.UserID, nil
}

func parseID(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid user id")
	}
	return uint(id), nil
}

// UsersLookup lists panel users for selectors.
// @Summary List of Users Lookup
// @Tags System(Users)
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} []models.UsersLookup
// @Router /system/users/lookup [get]
func (ctl *Controller) UsersLookup(c *echo.Context) error {
	result, err := ctl.system.UsersLookup(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}
