package groups

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	groupusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/groups"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	groups  *groupusecase.Usecase
}

func New(groupUsecase *groupusecase.Usecase) *Controller {
	return &Controller{
		request: request.NewCustomRequest(),
		groups:  groupUsecase,
	}
}

// OcservGroupsLookup 	 List of Ocserv group names
//
// @Summary      List of Ocserv group names
// @Description  List of Ocserv group names
// @Tags         Ocserv(Groups)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200 {array}  string
// @Router       /ocserv/groups/lookup [get]
func (ctl *Controller) OcservGroupsLookup(c *echo.Context) error {
	owner := ""
	val, ok := c.Get("isAdmin").(bool)
	if !ok || !val { // not admin or missing
		usernameVal, ok := c.Get("username").(string)
		if !ok || usernameVal == "" {
			return ctl.request.BadRequest(c, errors.New("invalid user uid"))
		}
		owner = usernameVal
	}
	groups, err := ctl.groups.Lookup(c.Request().Context(), owner)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, groups)
}

// OcservGroups 	 List of Ocserv groups
//
// @Summary      List of Ocserv groups
// @Description  List of Ocserv groups
// @Tags         Ocserv(Groups)
// @Accept       json
// @Produce      json
// @Param 		 page query int false "Page number, starting from 1" minimum(1)
// @Param 		 size query int false "Number of items per page" minimum(1) maximum(100) name(size)
// @Param 		 order query string false "Field to order by"
// @Param 		 sort query string false "Sort order, either ASC or DESC" Enums(ASC, DESC)
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object}  OcservGroupsResponse
// @Router       /ocserv/groups [get]
func (ctl *Controller) OcservGroups(c *echo.Context) error {
	pagination := ctl.request.Pagination(c)

	owner := ""
	if isAdmin := c.Get("isAdmin").(bool); !isAdmin {
		username := c.Get("username").(string)
		if username == "" {
			return ctl.request.BadRequest(c, errors.New("invalid username context"))
		}
		owner = username
	}

	result, err := ctl.groups.List(c.Request().Context(), groupusecase.ListOptions{Pagination: pagination, Owner: owner})
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, OcservGroupsResponse{
		Meta: request.Meta{
			Page:         result.Page,
			PageSize:     result.Size,
			TotalRecords: result.Total,
		},
		Result: result.Groups,
	})
}

// OcservGroup 	 Ocserv group detail
//
// @Summary      Ocserv group detail
// @Description  Ocserv group detail
// @Tags         Ocserv(Groups)
// @Accept       json
// @Produce      json
// @Param 		 id path int true "Ocserv Group ID"
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object}  models.OcservGroup
// @Router       /ocserv/groups/{id} [get]
func (ctl *Controller) OcservGroup(c *echo.Context) error {
	groupID := c.Param("id")
	if groupID == "" {
		return ctl.request.BadRequest(c, errors.New("invalid group id"))
	}

	group, err := ctl.groups.Get(c.Request().Context(), groupID)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, group)
}

// CreateOcservGroup 	     Ocserv Group creation
//
// @Summary      Ocserv Group creation
// @Description  Ocserv Group creation
// @Tags         Ocserv(Groups)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param        request    body  CreateOcservGroupData  true "ocserv group create data"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      201  {object} models.OcservGroup
// @Router       /ocserv/groups [post]
func (ctl *Controller) CreateOcservGroup(c *echo.Context) error {
	var data CreateOcservGroupData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	owner := c.Get("username").(string)
	if owner == "" {
		return ctl.request.BadRequest(c, errors.New("admin or staff username not found"))
	}

	newOcservGroup, err := ctl.groups.Create(c.Request().Context(), owner, data)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusCreated, newOcservGroup)
}

// UpdateOcservGroup 	     Ocserv Group update
//
// @Summary      Ocserv Group update
// @Description  Ocserv Group update
// @Tags         Ocserv(Groups)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 id path int true "Ocserv Group ID"
// @Param        request    body  UpdateOcservGroupData  true "ocserv group create data"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      201  {object} models.OcservGroup
// @Router       /ocserv/groups/{id} [patch]
func (ctl *Controller) UpdateOcservGroup(c *echo.Context) error {
	groupID := c.Param("id")
	if groupID == "" {
		return ctl.request.BadRequest(c, errors.New("invalid group id"))
	}

	var data UpdateOcservGroupData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	updatedOcservGroup, err := ctl.groups.Update(c.Request().Context(), groupID, data)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, updatedOcservGroup)
}

// DeleteOcservGroup 	     Ocserv Group delete
//
// @Summary      Ocserv Group delete
// @Description  Ocserv Group delete
// @Tags         Ocserv(Groups)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 id path int true "Ocserv Group ID"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      204  {object} nil
// @Router       /ocserv/groups/{id} [delete]
func (ctl *Controller) DeleteOcservGroup(c *echo.Context) error {
	groupID := c.Param("id")
	if groupID == "" {
		return ctl.request.BadRequest(c, errors.New("group id is empty"))
	}

	err := ctl.groups.Delete(c.Request().Context(), groupID)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

// GetDefaultsGroup 	     Ocserv Defaults Group config
//
// @Summary      Ocserv Defaults Group config
// @Description  Ocserv Defaults Group config
// @Tags         Ocserv(Groups)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} map[string]interface{}
// @Router       /ocserv/groups/defaults [get]
func (ctl *Controller) GetDefaultsGroup(c *echo.Context) error {
	conf, err := ctl.groups.Defaults()
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, conf)
}

// UpdateDefaultsGroup 	     Ocserv Defaults Group updating
//
// @Summary      Update Ocserv Defaults Group
// @Description  Update Ocserv Defaults Group
// @Tags         Ocserv(Groups)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param        request    body  UpdateOcservGroupData  true "ocserv group default data"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} nil
// @Router       /ocserv/groups/defaults [patch]
func (ctl *Controller) UpdateDefaultsGroup(c *echo.Context) error {
	var data UpdateOcservGroupData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	err := ctl.groups.UpdateDefaults(data.Config)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// ListUnsyncedGroups list of Unsynced Groups from os.dir
//
// @Summary      list of Unsynced Groups
// @Description  list of Unsynced Groups
// @Tags         Ocserv(UnsyncedGroup)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} []group.UnsyncedGroup
// @Router       /ocserv/groups/unsynced [get]
func (ctl *Controller) ListUnsyncedGroups(c *echo.Context) error {
	unsyncedGroups, err := ctl.groups.Unsynced(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, unsyncedGroups)
}

// SyncGroup     Ocserv Groups from file to db
//
// @Summary      Ocserv Groups from file
// @Description  Ocserv Groups from file
// @Tags         Ocserv(UnsyncedGroup)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param        request    body  SyncGroupRequest  true "list of groups with config to sync in db"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200 {object} []string
// @Router       /ocserv/groups/sync [post]
func (ctl *Controller) SyncGroup(c *echo.Context) error {
	owner := c.Get("username").(string)
	if owner == "" {
		return ctl.request.BadRequest(c, errors.New("admin or staff username not found"))
	}

	var data SyncGroupRequest
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	syncGroupNames, err := ctl.groups.Sync(c.Request().Context(), owner, data)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, syncGroupNames)
}
