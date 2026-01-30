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
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing/middlewares"
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

// Login		 Admin users login
//
// @Summary      Admin users login
// @Description  Admin users login with Google captcha(captcha site key required in get config api)
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body LoginData  true "login data"
// @Failure      400 {object} request.ErrorResponse
// @Success      200 {object} UserLoginResponse
// @Router       /users/login [post]
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		ctx = context.WithValue(ctx, "userUID", user.UID)
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

// Profile 		 Get User Profile
//
// @Summary      Get User Profile
// @Description  Get User Profile
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object}  models.User
// @Router       /users/profile [get]
func (ctl *Controller) Profile(c echo.Context) error {
	userUID := c.Get("userUID").(string)
	user, err := ctl.userRepo.GetByUID(c.Request().Context(), userUID)
	if err != nil {
		return middlewares.UnauthorizedError(c, "user not found")
	}
	return c.JSON(http.StatusOK, user)
}

// ChangePasswordBySelf 		 Change user password by self
//
// @Summary      Change user password by self
// @Description  Change user password by self
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body  ChangeUserPasswordBySelf  true "user new password"
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object}  UsersResponse
// @Router       /users/password [post]
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

// ========================================================================
// User Management
// ========================================================================

// Users 		 List of Users
//
// @Summary      List of Admin or staff users
// @Description  List of Admin or staff users
// @Tags         Users
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
// @Router       /users [get]
func (ctl *Controller) Users(c echo.Context) error {
	pagination := ctl.request.Pagination(c)

	var adminId uint
	role := c.Get("role").(apiModels.UserRole)
	if role == apiModels.RoleAdmin {
		adminId = c.Get("ID").(uint)
	}

	users, total, err := ctl.userRepo.Users(c.Request().Context(), pagination, &adminId)
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

// CreateUser	 Create user
//
// @Summary      Create user
// @Description  Create user Admin or simple
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param        request    body  CreateUserData   true "create admin user or staff data"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      201  {object} CreateUserResponse
// @Router       /users [post]
func (ctl *Controller) CreateUser(c echo.Context) error {
	var (
		data    CreateUserData
		adminID uint
		role    apiModels.UserRole
	)

	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	userRole := c.Get("role").(apiModels.UserRole)

	if userRole == apiModels.RoleSuperAdmin {
		if data.Role == nil {
			return ctl.request.BadRequest(c, errors.New("create user with super admin required role value"))
		}
		if *data.Role == apiModels.RoleStaff && data.AdminID == nil {
			return ctl.request.BadRequest(c, errors.New("create staff user with super admin required admin id value"))
		}
		role = *data.Role
		adminID = *data.AdminID
	} else {
		adminID = c.Get("ID").(uint)
		role = apiModels.RoleStaff
	}

	passwd := ctl.cryptoRepo.CreatePassword(data.Password)

	user := &models.User{
		Username: strings.ToLower(data.Username),
		Password: passwd.Hash,
		Salt:     passwd.Salt,
		Role:     role,
		AdminID:  &adminID,
	}

	newUser, err := ctl.userRepo.CreateUser(c.Request().Context(), user)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	var permissions []models.Permission
	if role == apiModels.RoleStaff {
		for _, p := range data.Permissions {
			permissions = append(permissions, models.Permission{
				UserID:  newUser.ID,
				Service: p.Service,
				Action:  p.Action,
			})
		}
		err2 := ctl.userRepo.CreateUserPermission(c.Request().Context(), permissions)
		if err2 != nil {
			return ctl.request.BadRequest(c, err2)
		}
	}

	return c.JSON(http.StatusCreated, CreateUserResponse{
		User:       newUser,
		Permission: permissions,
	})
}

// UpdateUserPermission	 Update staff user permission
//
// @Summary      Update staff user permission
// @Description  Update staff user permission
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 uid path string true "User UID"
// @Param        request    body  UpdatePermissionData   true "staff user permission data"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      200  {object} nil
// @Router       /users/{uid}/permissions [patch]
func (ctl *Controller) UpdateUserPermission(c echo.Context) error {
	userID := c.Param("uid")

	var data UpdatePermissionData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	user, err := ctl.userRepo.GetByUID(c.Request().Context(), userID)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	err2 := ctl.userRepo.RemoveUserPermission(c.Request().Context(), user.ID)
	if err2 != nil {
		return ctl.request.BadRequest(c, err2)
	}

	if len(data.Permissions) > 0 {
		var permissions []models.Permission
		for _, p := range data.Permissions {
			permissions = append(permissions, models.Permission{
				UserID:  user.ID,
				Service: p.Service,
				Action:  p.Action,
			})
		}
		err3 := ctl.userRepo.CreateUserPermission(c.Request().Context(), permissions)
		if err3 != nil {
			return ctl.request.BadRequest(c, err3)
		}
	}
	return c.JSON(http.StatusOK, nil)
}

// ChangeUserPasswordByAdmin 		 Change user password by admin
//
// @Summary      Change user password by admin
// @Description  Change user password by admin
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param 		 uid path string true "User UID"
// @Param        request    body  ChangeUserPassword  true "user new password"
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      200  {object}  UsersResponse
// @Router       /users/{uid}/password [post]
func (ctl *Controller) ChangeUserPasswordByAdmin(c echo.Context) error {
	userTargetID := c.Param("uid")

	var data ChangeUserPassword
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	passwd := ctl.cryptoRepo.CreatePassword(data.Password)

	err := ctl.userRepo.ChangePassword(c.Request().Context(), userTargetID, passwd.Hash, passwd.Salt)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// DeleteUser 	 Delete simple user
//
// @Summary      Delete simple user
// @Description  Delete simple user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param 		 uid path string true "User UID"
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      204  {object}  nil
// @Router       /users/{uid} [delete]
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

// UsersLookup 	 List of Users Lookup
//
// @Summary      List of Users Lookup
// @Description  List of Users Lookup
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      200  {object}  []models.UsersLookup
// @Router       /users/lookup [get]
func (ctl *Controller) UsersLookup(c echo.Context) error {
	users, err := ctl.userRepo.UsersLookup(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, users)
}
