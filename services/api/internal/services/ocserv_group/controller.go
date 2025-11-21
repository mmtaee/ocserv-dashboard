package ocserv_group

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/internal/repository"
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"net/http"
	"sync"
	"time"
)

type Controller struct {
	request         request.CustomRequestInterface
	ocservGroupRepo repository.OcservGroupRepositoryInterface
	ocservUserRepo  repository.OcservUserRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request:         request.NewCustomRequest(),
		ocservGroupRepo: repository.NewOcservGroupRepository(),
		ocservUserRepo:  repository.NewtOcservUserRepository(),
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
func (ctl *Controller) OcservGroupsLookup(c echo.Context) error {
	owner := ""
	if isAdmin := c.Get("isAdmin").(bool); !isAdmin {
		username := c.Get("username").(string)
		if username == "" {
			return ctl.request.BadRequest(c, errors.New("invalid username context"))
		}
		owner = username
	}

	groups, err := ctl.ocservGroupRepo.GroupsLookup(c.Request().Context(), owner)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	groups = append([]string{"defaults"}, groups...)
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
func (ctl *Controller) OcservGroups(c echo.Context) error {
	pagination := ctl.request.Pagination(c)

	owner := ""
	if isAdmin := c.Get("isAdmin").(bool); !isAdmin {
		username := c.Get("username").(string)
		if username == "" {
			return ctl.request.BadRequest(c, errors.New("invalid username context"))
		}
		owner = username
	}

	ocservGroup, total, err := ctl.ocservGroupRepo.Groups(c.Request().Context(), pagination, owner)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, OcservGroupsResponse{
		Meta: request.Meta{
			Page:         pagination.Page,
			PageSize:     pagination.PageSize,
			TotalRecords: total,
		},
		Result: ocservGroup,
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
func (ctl *Controller) OcservGroup(c echo.Context) error {
	groupID := c.Param("id")
	if groupID == "" {
		return ctl.request.BadRequest(c, errors.New("invalid group id"))
	}

	group, err := ctl.ocservGroupRepo.GetByID(c.Request().Context(), groupID)
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
func (ctl *Controller) CreateOcservGroup(c echo.Context) error {
	var data CreateOcservGroupData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	owner := c.Get("username").(string)
	if owner == "" {
		return ctl.request.BadRequest(c, errors.New("admin or staff username not found"))
	}

	ocservGroup := models.OcservGroup{
		Name:   data.Name,
		Owner:  owner,
		Config: data.Config,
	}

	newOcservGroup, err := ctl.ocservGroupRepo.Create(c.Request().Context(), &ocservGroup)
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
func (ctl *Controller) UpdateOcservGroup(c echo.Context) error {
	groupID := c.Param("id")
	if groupID == "" {
		return ctl.request.BadRequest(c, errors.New("invalid group id"))
	}

	var data UpdateOcservGroupData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	ocservGroup, err := ctl.ocservGroupRepo.GetByID(c.Request().Context(), groupID)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	ocservGroup.Config = data.Config
	updatedOcservGroup, err := ctl.ocservGroupRepo.Update(c.Request().Context(), ocservGroup)
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
func (ctl *Controller) DeleteOcservGroup(c echo.Context) error {
	groupID := c.Param("id")
	if groupID == "" {
		return ctl.request.BadRequest(c, errors.New("group id is empty"))
	}

	group, err := ctl.ocservGroupRepo.Delete(c.Request().Context(), groupID)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	go func(groupName string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var wg sync.WaitGroup

		users, err := ctl.ocservUserRepo.UpdateUsersByDeleteGroup(ctx, groupName)
		if err != nil {
			logger.Error("Failed to load users for removed group %s: %v", groupName, err)
			return
		}

		for _, u := range *users {
			// create local copy for goroutine
			ocservUser := u
			ocservUser.Group = "defaults"

			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, err2 := ctl.ocservUserRepo.Update(ctx, &ocservUser); err2 != nil {
					logger.Warn("DeleteGroup: failed to update user %s: %v", ocservUser.Username, err2)
				}
			}()
		}

		wg.Wait()
	}(group.Name)

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
func (ctl *Controller) GetDefaultsGroup(c echo.Context) error {
	conf, err := ctl.ocservGroupRepo.DefaultGroup()
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
func (ctl *Controller) UpdateDefaultsGroup(c echo.Context) error {
	var data UpdateOcservGroupData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	err := ctl.ocservGroupRepo.UpdateDefaultGroup(data.Config)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}
