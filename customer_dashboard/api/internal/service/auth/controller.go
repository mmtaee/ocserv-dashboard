package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/internal/usecase"
)

type AuthController struct {
	authUC    usecase.AuthUseCase
	req       *request.Request
	validator *request.Validator
}

func NewAuthController(authUC usecase.AuthUseCase) *AuthController {
	return &AuthController{
		authUC:    authUC,
		req:       &request.Request{},
		validator: request.NewValidator(),
	}
}

// Login handles customer login
// @Summary Customer Login
// @Description Login with ocserv username and password to get a token
// @Tags Customer Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /customer/auth/login [post]
func (ctrl *AuthController) Login(c *echo.Context) error {
	var req LoginRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	token, expiresAt, err := ctrl.authUC.Login(req.Username, req.Password)
	if err != nil {
		if err.Error() == "8001" {
			return ctrl.req.ResponseWithCode(c, "8001", nil)
		}
		return ctrl.req.Unauthorized(c, err, err.Error())
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	})
}
