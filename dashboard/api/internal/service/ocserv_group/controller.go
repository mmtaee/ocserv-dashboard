package ocserv_group

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/request"
)

type OcservGroupController struct {
	groupUC usecase.OcservGroupUseCase
	req     *request.Request
	validator *request.Validator
}

func NewOcservGroupController(groupUC usecase.OcservGroupUseCase) *OcservGroupController {
	return &OcservGroupController{
		groupUC: groupUC,
		req: &request.Request{},
		validator: request.NewValidator(),
	}
}

// ListGroups godoc
// @Summary List Ocserv Groups
// @Description Get all Ocserv groups for current admin (superadmin can access all groups)
// @Tags Ocserv Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {array} models.OcservGroupResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /api/v1/ocserv/groups [get]
func (ctrl *OcservGroupController) ListGroups(c *echo.Context) error {
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	groups, err := ctrl.groupUC.ListGroups(adminID, role)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, groups)
}

// GetGroup godoc
// @Summary Get Ocserv Group
// @Description Get Ocserv group by ID (superadmin can access any group)
// @Tags Ocserv Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Group ID"
// @Success 200 {object} models.OcservGroupResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /api/v1/ocserv/groups/{id} [get]
func (ctrl *OcservGroupController) GetGroup(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid group ID")
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	group, err := ctrl.groupUC.GetGroup(uint(id), adminID, role)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	return c.JSON(http.StatusOK, group)
}

// CreateGroup godoc
// @Summary Create Ocserv Group
// @Description Create new Ocserv group (superadmin can create groups for any admin)
// @Tags Ocserv Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body CreateOcservGroupRequest true "Create group data"
// @Success 201 {object} models.OcservGroup
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 409 {object} request.ErrorResponse
// @Router /api/v1/ocserv/groups [post]
func (ctrl *OcservGroupController) CreateGroup(c *echo.Context) error {
	var req CreateOcservGroupRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	group, err := ctrl.groupUC.CreateGroup(req.Name, req.Config, adminID, role)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	return c.JSON(http.StatusCreated, group)
}

// UpdateGroup godoc
// @Summary Update Ocserv Group
// @Description Update Ocserv group by ID (superadmin can update any group)
// @Tags Ocserv Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Group ID"
// @Param request body UpdateOcservGroupRequest true "Update group data"
// @Success 200 {object} models.OcservGroup
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /api/v1/ocserv/groups/{id} [post]
func (ctrl *OcservGroupController) UpdateGroup(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid group ID")
	}
	var req UpdateOcservGroupRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	group, err := ctrl.groupUC.UpdateGroup(uint(id), req.Config, adminID, role)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	return c.JSON(http.StatusOK, group)
}

// DeleteGroup godoc
// @Summary Delete Ocserv Group
// @Description Delete Ocserv group by ID (superadmin can delete any group)
// @Tags Ocserv Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Group ID"
// @Success 204 {object} nil
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /api/v1/ocserv/groups/{id} [delete]
func (ctrl *OcservGroupController) DeleteGroup(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.BadRequest(c, err, "invalid group ID")
	}
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	if err := ctrl.groupUC.DeleteGroup(uint(id), adminID, role); err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// GroupsLookup godoc
// @Summary Groups Lookup
// @Description Get list of group names (superadmin can access all groups)
// @Tags Ocserv Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {array} string
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /api/v1/ocserv/groups/lookup [get]
func (ctrl *OcservGroupController) GroupsLookup(c *echo.Context) error {
	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	names, err := ctrl.groupUC.GetGroupsLookup(adminID, role)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, names)
}
