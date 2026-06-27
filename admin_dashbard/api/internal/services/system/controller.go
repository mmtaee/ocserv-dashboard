package system

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing/middlewares"
	"gorm.io/gorm"
)

type Controller struct {
	request request.CustomRequestInterface
	usecase usecase.SystemUsecaseInterface
}

func New(systemUsecase usecase.SystemUsecaseInterface) *Controller {
	return &Controller{
		request: request.NewCustomRequest(),
		usecase: systemUsecase,
	}
}

func (ctl *Controller) DashboardRelease(c echo.Context) error {
	current, latest, err := ctl.usecase.DashboardRelease()
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(http.StatusOK, usecase.DashboardReleaseResponse{
		Current: current,
		Latest:  latest,
	})
}

func (ctl *Controller) SetupSystem(c echo.Context) error {
	var data usecase.SetupSystemData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err, "1003")
	}

	user, system, token, err := ctl.usecase.SetupSystem(c.Request().Context(), data)
	if err != nil {
		if err.Error() == "the system is already configured" {
			return ctl.request.BadRequest(c, err, "1015")
		}
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(
		http.StatusCreated,
		usecase.SetupSystemResponse{
			User:   *user,
			System: *system,
			Token:  token,
		},
	)
}

func (ctl *Controller) ResetAdminPassword(c echo.Context) error {
	var data usecase.ResetAdminPasswordData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err, "1003")
	}

	user, token, err := ctl.usecase.ResetAdminPassword(c.Request().Context(), data)
	if err != nil {
		if err.Error() == "the secret key is invalid" {
			return ctl.request.BadRequest(c, err, "1016")
		}
		if err.Error() == "username not found" {
			return ctl.request.BadRequest(c, err, "1017")
		}
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(http.StatusOK, usecase.ResetPasswordResponse{
		User:  user,
		Token: token,
	})
}

func (ctl *Controller) SystemInit(c echo.Context) error {
	googleCaptchaSiteKey, telegramBotEnabled, err := ctl.usecase.SystemInit(c.Request().Context())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusOK, nil)
		}
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(http.StatusOK, usecase.GetSystemInitResponse{
		GoogleCaptchaSiteKey: googleCaptchaSiteKey,
		TelegramBotEnabled:   telegramBotEnabled,
	})
}

func (ctl *Controller) System(c echo.Context) error {
	res, err := ctl.usecase.System(c.Request().Context())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusOK, nil)
		}
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, res)
}

func (ctl *Controller) SystemUpdate(c echo.Context) error {
	userUID := c.Param("userUID")

	var data usecase.PatchSystemUpdateData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err, "1003")
	}

	res, err := ctl.usecase.SystemUpdate(c.Request().Context(), data, userUID)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, res)
}

func (ctl *Controller) Login(c echo.Context) error {
	var data usecase.LoginData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err, "1003")
	}

	user, token, err := ctl.usecase.Login(c.Request().Context(), data)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1014")
	}

	return c.JSON(http.StatusOK, usecase.UserLoginResponse{
		User:  user,
		Token: token,
	})
}

func (ctl *Controller) CreateUser(c echo.Context) error {
	var data usecase.CreateUserData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err, "1003")
	}

	newUser, err := ctl.usecase.CreateUser(c.Request().Context(), data)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusCreated, newUser)
}

func (ctl *Controller) Users(c echo.Context) error {
	pagination := ctl.request.Pagination(c)

	users, total, err := ctl.usecase.Users(c.Request().Context(), pagination)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(http.StatusOK, usecase.UsersResponse{
		Meta: request.Meta{
			Page:         pagination.Page,
			TotalRecords: total,
			PageSize:     pagination.PageSize,
		},
		Result: users,
	})
}

func (ctl *Controller) ChangeUserPasswordByAdmin(c echo.Context) error {
	userTargetID := c.Param("uid")

	var data usecase.ChangeUserPasswordData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err, "1003")
	}
	err := ctl.usecase.ChangeUserPasswordByAdmin(c.Request().Context(), userTargetID, data)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, nil)
}

func (ctl *Controller) DeleteUser(c echo.Context) error {
	deleteUserID := c.Param("uid")
	userUID := c.Param("userUID")

	err := ctl.usecase.DeleteUser(c.Request().Context(), deleteUserID, userUID)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusNoContent, nil)
}

func (ctl *Controller) ChangePasswordBySelf(c echo.Context) error {
	userUID := c.Get("userUID").(string)

	var data usecase.ChangePasswordBySelfData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err, "1003")
	}

	err := ctl.usecase.ChangePasswordBySelf(c.Request().Context(), userUID, data)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, nil)
}

func (ctl *Controller) Profile(c echo.Context) error {
	userUID := c.Get("userUID").(string)
	user, err := ctl.usecase.Profile(c.Request().Context(), userUID)
	if err != nil {
		return middlewares.UnauthorizedError(c, "1018")
	}
	return c.JSON(http.StatusOK, user)
}

func (ctl *Controller) UsersLookup(c echo.Context) error {
	users, err := ctl.usecase.UsersLookup(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, users)
}
