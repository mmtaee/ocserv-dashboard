package system

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/internal/models"
	"github.com/mmtaee/ocserv-users-management/api/internal/repository"
	"github.com/mmtaee/ocserv-users-management/api/pkg/captcha"
	"github.com/mmtaee/ocserv-users-management/api/pkg/crypto"
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing/middlewares"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
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
		IsAdmin:  true,
	}

	inactiveDays := data.KeepInactiveUserDays
	if inactiveDays < 1 {
		inactiveDays = 1
	}

	system := &models.System{
		GoogleCaptchaSiteKey:    data.GoogleCaptchaSiteKey,
		GoogleCaptchaSecretKey:  data.GoogleCaptchaSecretKey,
		AutoDeleteInactiveUsers: data.AutoDeleteInactiveUsers,
		KeepInactiveUserDays:    inactiveDays,
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
		GoogleCaptchaSiteKey:    config.GoogleCaptchaSiteKey,
		GoogleCaptchaSecretKey:  config.GoogleCaptchaSecretKey,
		AutoDeleteInactiveUsers: config.AutoDeleteInactiveUsers,
		KeepInactiveUserDays:    config.KeepInactiveUserDays,
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
	if data.AutoDeleteInactiveUsers != nil {
		system.AutoDeleteInactiveUsers = *data.AutoDeleteInactiveUsers
	}
	if data.KeepInactiveUserDays != nil {
		inactiveDays := *data.KeepInactiveUserDays
		if inactiveDays < 1 {
			inactiveDays = 1
		}
		system.KeepInactiveUserDays = inactiveDays
	}

	ctx := context.WithValue(c.Request().Context(), "userUID", userUID)
	updatedConfig, err := ctl.systemRepo.SystemUpdate(ctx, &system)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, GetSystemResponse{
		GoogleCaptchaSiteKey:    updatedConfig.GoogleCaptchaSiteKey,
		GoogleCaptchaSecretKey:  updatedConfig.GoogleCaptchaSecretKey,
		AutoDeleteInactiveUsers: updatedConfig.AutoDeleteInactiveUsers,
		KeepInactiveUserDays:    updatedConfig.KeepInactiveUserDays,
	})
}

// Login		 Admin users login
//
// @Summary      Admin users login
// @Description  Admin users login with Google captcha(captcha site key required in get config api)
// @Tags         System(Users)
// @Accept       json
// @Produce      json
// @Param        request body LoginData  true "login data"
// @Failure      400 {object} request.ErrorResponse
// @Success      200 {object} UserLoginResponse
// @Router       /system/users/login [post]
func (ctl *Controller) Login(c echo.Context) error {
	var data LoginData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	system, err := ctl.systemRepo.System(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	if secretKey := system.GoogleCaptchaSecretKey; secretKey != "" {
		ctl.captchaVerifier.SetSecretKey(secretKey)
		ctl.captchaVerifier.Verify(data.Token)
		if !ctl.captchaVerifier.IsValid() {
			return ctl.request.BadRequest(c, errors.New("captcha challenge failed"))
		}
	}

	user, err := ctl.userRepo.GetByUsername(c.Request().Context(), data.Username)
	if err != nil {
		return ctl.request.BadRequest(c, errors.New("invalid username or password"))
	}

	if ok := ctl.cryptoRepo.CheckPassword(data.Password, user.Password, user.Salt); !ok {
		return ctl.request.BadRequest(c, errors.New("invalid username or password"))
	}

	token, err := ctl.userRepo.CreateToken(c.Request().Context(), user, data.RememberMe)
	if err != nil {
		return ctl.request.BadRequest(c, err, "user created")
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), "userUID", user.UID), 10*time.Second)
		ctx = context.WithValue(ctx, "username", user.Username)
		defer cancel()

		now := time.Now()
		user.LastLogin = &now
		_ = ctl.userRepo.UpdateLastLogin(ctx, user)
	}()

	return c.JSON(http.StatusOK, UserLoginResponse{
		User:  user,
		Token: token,
	})
}

// CreateUser	 Create user
//
// @Summary      Create user
// @Description  Create user Admin or simple
// @Tags         System(Users)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param        request    body  CreateUserData   true "create user data"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      201  {object}  models.User
// @Router       /system/users [post]
func (ctl *Controller) CreateUser(c echo.Context) error {
	//userUID := c.Param("userUID")

	var data CreateUserData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	passwd := ctl.cryptoRepo.CreatePassword(data.Password)

	user := &models.User{
		Username: strings.ToLower(data.Username),
		Password: passwd.Hash,
		Salt:     passwd.Salt,
		IsAdmin:  false,
	}

	//ctx := context.WithValue(c.Request().Context(), "userUID", userUID)
	newUser, err := ctl.userRepo.CreateUser(c.Request().Context(), user)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusCreated, newUser)
}

// Users 		 List of Users
//
// @Summary      List of Admin or simple users
// @Description  List of Admin or simple users
// @Tags         System(Users)
// @Accept       json
// @Produce      json
// @Param 		 page query int false "Page number, starting from 1" minimum(1)
// @Param 		 size query int false "Number of items per page" minimum(1) maximum(100) name(size)
// @Param 		 order query string false "Field to order by"
// @Param 		 sort query string false "Sort order, either ASC or DESC" Enums(ASC, DESC)
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      200  {object}  UsersResponse
// @Router       /system/users [get]
func (ctl *Controller) Users(c echo.Context) error {
	pagination := ctl.request.Pagination(c)

	users, total, err := ctl.userRepo.Users(c.Request().Context(), pagination)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, UsersResponse{
		Meta: request.Meta{
			Page:         pagination.Page,
			PageSize:     pagination.PageSize,
			TotalRecords: total,
		},
		Result: users,
	})
}

// ChangeUserPasswordByAdmin 		 Change user password by admin
//
// @Summary      Change user password by admin
// @Description  Change user password by admin
// @Tags         System(Users)
// @Accept       json
// @Produce      json
// @Param 		 uid path string true "User UID"
// @Param        request    body  ChangeUserPassword  true "user new password"
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      200  {object}  UsersResponse
// @Router       /system/users/{uid}/password [post]
func (ctl *Controller) ChangeUserPasswordByAdmin(c echo.Context) error {
	userTargetID := c.Param("uid")
	userUID := c.Param("userUID")

	var data ChangeUserPassword
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	passwd := ctl.cryptoRepo.CreatePassword(data.Password)

	ctx := context.WithValue(c.Request().Context(), "userUID", userUID)

	err := ctl.userRepo.ChangePassword(ctx, userTargetID, passwd.Hash, passwd.Salt)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// DeleteUser 	 Delete simple user
//
// @Summary      Delete simple user
// @Description  Delete simple user
// @Tags         System(Users)
// @Accept       json
// @Produce      json
// @Param 		 uid path string true "User UID"
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      204  {object}  nil
// @Router       /system/users/{uid} [delete]
func (ctl *Controller) DeleteUser(c echo.Context) error {
	deleteUserID := c.Param("uid")
	userUID := c.Param("userUID")

	ctx := context.WithValue(c.Request().Context(), "userUID", userUID)
	err := ctl.userRepo.DeleteUser(ctx, deleteUserID)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

// ChangePasswordBySelf 		 Change user password by self
//
// @Summary      Change user password by self
// @Description  Change user password by self
// @Tags         System(Users)
// @Accept       json
// @Produce      json
// @Param        request body  ChangeUserPasswordBySelf  true "user new password"
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object}  UsersResponse
// @Router       /system/users/password [post]
func (ctl *Controller) ChangePasswordBySelf(c echo.Context) error {
	userUID := c.Get("userUID").(string)

	var data ChangeUserPasswordBySelf
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	user, _ := ctl.userRepo.GetByUID(c.Request().Context(), userUID)
	if ok := ctl.cryptoRepo.CheckPassword(data.OldPassword, user.Password, user.Salt); !ok {
		return ctl.request.BadRequest(c, errors.New("invalid old password"))
	}

	passwd := ctl.cryptoRepo.CreatePassword(data.NewPassword)
	err := ctl.userRepo.ChangePassword(c.Request().Context(), userUID, passwd.Hash, passwd.Salt)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// Profile 		 Get User Profile
//
// @Summary      Get User Profile
// @Description  Get User Profile
// @Tags         System(Users)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object}  models.User
// @Router       /system/users/profile [get]
func (ctl *Controller) Profile(c echo.Context) error {
	userUID := c.Get("userUID").(string)
	user, err := ctl.userRepo.GetByUID(c.Request().Context(), userUID)
	if err != nil {
		return middlewares.UnauthorizedError(c, "user not found")
	}
	return c.JSON(http.StatusOK, user)
}

// UsersLookup 	 List of Users Lookup
//
// @Summary      List of Users Lookup
// @Description  List of Users Lookup
// @Tags         System(Users)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      200  {object}  []models.UsersLookup
// @Router       /system/users/lookup [get]
func (ctl *Controller) UsersLookup(c echo.Context) error {
	users, err := ctl.userRepo.UsersLookup(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, users)
}
