package system

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/internal/models"
	apiModels "github.com/mmtaee/ocserv-users-management/api/internal/models"
	"github.com/mmtaee/ocserv-users-management/api/internal/repository"
	"github.com/mmtaee/ocserv-users-management/api/pkg/captcha"
	"github.com/mmtaee/ocserv-users-management/api/pkg/crypto"
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type Controller struct {
	request         request.CustomRequestInterface
	systemRepo      repository.SystemRepositoryInterface
	userRepo        repository.UserRepositoryInterface
	captchaVerifier captcha.GoogleCaptchaInterface
	cryptoRepo      crypto.CustomPasswordInterface
}

func New() *Controller {
	return &Controller{
		request:         request.NewCustomRequest(),
		systemRepo:      repository.NewSystemRepository(),
		userRepo:        repository.NewUserRepository(),
		captchaVerifier: captcha.NewGoogleVerifier(),
		cryptoRepo:      crypto.NewCustomPassword(),
	}
}

// SetupSystem
// @Summary      Setup user and system config
// @Description  Setup user and system config
// @Tags         System
// @Accept       json
// @Produce      json
// @Param        request  body  SetupSystem   true "system setup data"
// @Failure      400 {object} request.ErrorResponse
// @Success      201  {object}  SetupSystemResponse
// @Router       /system/setup [post]
func (ctl *Controller) SetupSystem(c echo.Context) error {
	if _, err := ctl.systemRepo.System(c.Request().Context()); err == nil {
		return ctl.request.BadRequest(c, errors.New("the system is already configured"))
	}

	var data SetupSystem
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	passwd := ctl.cryptoRepo.CreatePassword(data.Password)

	user := &models.User{
		Username: strings.ToLower(data.Username),
		Password: passwd.Hash,
		Salt:     passwd.Salt,
		Role:     apiModels.RoleSuperAdmin,
	}

	system := &models.System{
		GoogleCaptchaSiteKey:   data.GoogleCaptchaSiteKey,
		GoogleCaptchaSecretKey: data.GoogleCaptchaSecretKey,
	}
	newUser, newSystem, err := ctl.systemRepo.SystemSetup(c.Request().Context(), user, system)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	token, err := ctl.userRepo.CreateToken(c.Request().Context(), newUser, true)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(
		http.StatusCreated,
		SetupSystemResponse{
			User:   *newUser,
			System: *newSystem,
			Token:  token,
		},
	)
}

// SystemInit
// @Summary      Get panel System init Config
// @Description  Get panel System init Config
// @Tags         System
// @Accept       json
// @Produce      json
// @Failure      400 {object} request.ErrorResponse
// @Success      200  {object}  GetSystemInitResponse
// @Router       /system/init [get]
func (ctl *Controller) SystemInit(c echo.Context) error {
	config, err := ctl.systemRepo.System(c.Request().Context())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusOK, nil)
		}
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, GetSystemInitResponse{
		GoogleCaptchaSiteKey: config.GoogleCaptchaSiteKey,
	})
}

// System
// @Summary      Get panel System Config
// @Description  Get panel System Config
// @Tags         System
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object}  GetSystemResponse
// @Router       /system [get]
func (ctl *Controller) System(c echo.Context) error {
	config, err := ctl.systemRepo.System(c.Request().Context())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusOK, nil)
		}
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, GetSystemResponse{
		GoogleCaptchaSiteKey:   config.GoogleCaptchaSiteKey,
		GoogleCaptchaSecretKey: config.GoogleCaptchaSecretKey,
	})
}

// SystemUpdate
// @Summary      Update panel System Config
// @Description  Update panel System Config
// @Tags         System
// @Accept       json
// @Produce      json
// @Param        request    body  PatchSystemUpdateData   true "update system config data"
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object}  GetSystemResponse
// @Router       /system [patch]
func (ctl *Controller) SystemUpdate(c echo.Context) error {
	userUID := c.Param("userUID")

	var data PatchSystemUpdateData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	system := models.System{}

	if data.GoogleCaptchaSiteKey != nil {
		system.GoogleCaptchaSiteKey = *data.GoogleCaptchaSiteKey
	}
	if data.GoogleCaptchaSecretKey != nil {
		system.GoogleCaptchaSecretKey = *data.GoogleCaptchaSecretKey
	}

	ctx := context.WithValue(c.Request().Context(), "userUID", userUID)
	updatedConfig, err := ctl.systemRepo.SystemUpdate(ctx, &system)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, GetSystemResponse{
		GoogleCaptchaSiteKey:   updatedConfig.GoogleCaptchaSiteKey,
		GoogleCaptchaSecretKey: updatedConfig.GoogleCaptchaSecretKey,
	})
}

// AvailablePermissions 	list of available permissions
//
// @Summary      list of available permissions
// @Description  list of available permissions
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.User
// @Router       /system/permissions [get]
func (ctl *Controller) AvailablePermissions(c echo.Context) error {
	permissions := []string{
		"ocserv-groups.crud",
		"ocserv-users.crud",
		"ocserv-users.action",
		"ocserv-users.stats",
	}
	return c.JSON(http.StatusOK, permissions)
}
