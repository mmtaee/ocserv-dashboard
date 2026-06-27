package auth

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
)

type AuthController struct {
	adminUseCase usecase.AdminUseCase
	req          *request.Request
	validator    *request.Validator
}

func NewAuthController(adminUseCase usecase.AdminUseCase) *AuthController {
	return &AuthController{
		adminUseCase: adminUseCase,
		req:          &request.Request{},
		validator:    request.NewValidator(),
	}
}

// Login handles admin login
// @Summary Admin Login
// @Description Login with username and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /auth/login [post]
func (ctrl *AuthController) Login(c *echo.Context) error {
	var req LoginRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	token, admin, err := ctrl.adminUseCase.Login(req.Username, req.Password)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		Admin: admin,
	})
}

// GetProfile returns current admin profile
// @Summary Get Admin Profile
// @Description Return authenticated admin's profile
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} models.Administrator
// @Failure 401 {object} request.ErrorResponse
// @Router /auth/profile [get]
func (ctrl *AuthController) GetProfile(c *echo.Context) error {
	adminID := c.Get("id").(uint)

	admin, err := ctrl.adminUseCase.GetProfile(adminID)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}

	return c.JSON(http.StatusOK, admin)
}

// ChangePassword changes admin password
// @Summary Change Admin Password
// @Description Update authenticated admin's password
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body ChangePasswordRequest true "Change Password Data"
// @Success 200 {object} ChangePasswordResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /auth/change-password [post]
func (ctrl *AuthController) ChangePassword(c *echo.Context) error {
	adminID := c.Get("id").(uint)

	var req ChangePasswordRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	token, admin, err := ctrl.adminUseCase.ChangePassword(adminID, req.OldPassword, req.NewPassword)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, ChangePasswordResponse{
		Token: token,
		Admin: admin,
	})
}

// Logout handles admin logout
// @Summary Admin Logout
// @Description Logout current admin and invalidate token
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} request.MessageResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /auth/logout [post]
func (ctrl *AuthController) Logout(c *echo.Context) error {
	token := c.Get("token").(string)

	if err := ctrl.adminUseCase.Logout(token); err != nil {
		return ctrl.req.InternalServerError(c, err)
	}

	return c.JSON(http.StatusOK, request.MessageResponse{
		Message: "Logged out successfully",
	})
}
