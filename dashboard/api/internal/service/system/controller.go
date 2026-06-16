package system

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/request"
)

type SystemController struct {
	systemUseCase usecase.SystemUseCase
	req           *request.Request
	validator     *request.Validator
}

func NewSystemController(systemUseCase usecase.SystemUseCase) *SystemController {
	return &SystemController{
		systemUseCase: systemUseCase,
		req:           &request.Request{},
		validator:     request.NewValidator(),
	}
}

// GetSystem returns the system configuration
// @Summary Get System Config
// @Description Return the system configuration
// @Tags System
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} models.System
// @Failure 401 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /system [get]
func (ctrl *SystemController) GetSystem(c *echo.Context) error {
	system, err := ctrl.systemUseCase.Get()
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, system)
}

// UpdateSystem updates the system configuration
// @Summary Update System Config
// @Description Update the system configuration
// @Tags System
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body UpdateSystemRequest true "System Config Data"
// @Success 200 {object} models.System
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /system [post]
func (ctrl *SystemController) UpdateSystem(c *echo.Context) error {
	var req UpdateSystemRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	system, err := ctrl.systemUseCase.Get()
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	// Update fields
	system.GoogleCaptchaSecretKey = req.GoogleCaptchaSecretKey
	system.GoogleCaptchaSiteKey = req.GoogleCaptchaSiteKey
	system.AutoDeleteInactiveUsers = req.AutoDeleteInactiveUsers
	system.KeepInactiveUserDays = req.KeepInactiveUserDays
	system.ClientProfileServerAddress = req.ClientProfileServerAddress
	system.ClientProfileServerPort = req.ClientProfileServerPort
	system.ClientProfileConnectionName = req.ClientProfileConnectionName

	updatedSystem, err := ctrl.systemUseCase.Update(system)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, updatedSystem)
}
