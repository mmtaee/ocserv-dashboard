package super_admin

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/request"
)

type SuperAdminController struct {
	superAdminUC usecase.SuperAdminUseCase
	req          *request.Request
	validator    *request.Validator
}

func NewSuperAdminController(superAdminUC usecase.SuperAdminUseCase) *SuperAdminController {
	return &SuperAdminController{
		superAdminUC: superAdminUC,
		req:          &request.Request{},
		validator:    request.NewValidator(),
	}
}

// CreateAdmin handles creating a new admin
// @Summary Create Admin
// @Description Create a new admin
// @Tags Super Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body CreateAdminRequest true "Create Admin"
// @Success 200 {object} models.Administrator
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 409 {object} request.ErrorResponse
// @Router /super-admin/admins [post]
func (ctrl *SuperAdminController) CreateAdmin(c *echo.Context) error {
	var req CreateAdminRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	admin, err := ctrl.superAdminUC.CreateAdmin(req.Username, req.Password)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, admin)
}

// ListAdmins handles listing all admins
// @Summary List Admins
// @Description List all admins
// @Tags Super Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {array} models.Administrator
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /super-admin/admins [get]
func (ctrl *SuperAdminController) ListAdmins(c *echo.Context) error {
	admins, err := ctrl.superAdminUC.ListAdmins()
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}

	return c.JSON(http.StatusOK, admins)
}

// GetAdmin handles getting an admin by ID
// @Summary Get Admin
// @Description Get an admin by ID
// @Tags Super Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Admin ID"
// @Success 200 {object} models.Administrator
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /super-admin/admins/{id} [get]
func (ctrl *SuperAdminController) GetAdmin(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid admin ID")
	}

	admin, err := ctrl.superAdminUC.GetAdmin(uint(id))
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, admin)
}

// UpdateAdmin handles updating an admin
// @Summary Update Admin
// @Description Update an admin
// @Tags Super Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Admin ID"
// @Param request body UpdateAdminRequest true "Update Admin"
// @Success 200 {object} models.Administrator
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Failure 409 {object} request.ErrorResponse
// @Router /super-admin/admins/{id} [post]
func (ctrl *SuperAdminController) UpdateAdmin(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid admin ID")
	}

	var req UpdateAdminRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	admin, err := ctrl.superAdminUC.UpdateAdmin(uint(id), req.Username)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, admin)
}

// ChangeAdminPassword handles changing an admin's password
// @Summary Change Admin Password
// @Description Change an admin's password
// @Tags Super Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Admin ID"
// @Param request body ChangeAdminPasswordRequest true "Change Password"
// @Success 200 {object} request.MessageResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /super-admin/admins/{id}/change-password [post]
func (ctrl *SuperAdminController) ChangeAdminPassword(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid admin ID")
	}

	var req ChangeAdminPasswordRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	err = ctrl.superAdminUC.ChangeAdminPassword(uint(id), req.Password)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, request.MessageResponse{Message: "Password changed successfully"})
}

// SuspendAdmin handles suspending an admin
// @Summary Suspend Admin
// @Description Suspend an admin
// @Tags Super Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Admin ID"
// @Param request body SuspendAdminRequest true "Suspend Admin"
// @Success 200 {object} request.MessageResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /super-admin/admins/{id}/suspend [post]
func (ctrl *SuperAdminController) SuspendAdmin(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid admin ID")
	}

	var req SuspendAdminRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	err = ctrl.superAdminUC.SuspendAdmin(uint(id), req.Reason)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, request.MessageResponse{Message: "Admin suspended successfully"})
}

// UnsuspendAdmin handles unsuspending an admin
// @Summary Unsuspend Admin
// @Description Unsuspend an admin
// @Tags Super Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Admin ID"
// @Success 200 {object} request.MessageResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /super-admin/admins/{id}/unsuspend [post]
func (ctrl *SuperAdminController) UnsuspendAdmin(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid admin ID")
	}

	err = ctrl.superAdminUC.UnsuspendAdmin(uint(id))
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, request.MessageResponse{Message: "Admin unsuspended successfully"})
}
