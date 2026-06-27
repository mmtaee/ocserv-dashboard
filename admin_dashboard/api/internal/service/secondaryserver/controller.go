package secondaryserver

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
)

type SecondaryServerController struct {
	serverUseCase usecase.SecondaryServerUseCase
	req           *request.Request
	validator     *request.Validator
}

func NewSecondaryServerController(serverUseCase usecase.SecondaryServerUseCase) *SecondaryServerController {
	return &SecondaryServerController{
		serverUseCase: serverUseCase,
		req:           &request.Request{},
		validator:     request.NewValidator(),
	}
}

// List handles getting all secondary servers
// @Summary List secondary servers
// @Description Get all secondary servers
// @Tags SecondaryServer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {array} models.SecondaryServer
// @Failure 401 {object} request.ErrorResponse
// @Router /secondary-servers [get]
func (ctrl *SecondaryServerController) List(c *echo.Context) error {
	servers, err := ctrl.serverUseCase.GetAll()
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, servers)
}

// Get handles getting a single secondary server by ID
// @Summary Get secondary server
// @Description Get secondary server by ID
// @Tags SecondaryServer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Secondary Server ID"
// @Success 200 {object} models.SecondaryServer
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /secondary-servers/{id} [get]
func (ctrl *SecondaryServerController) Get(c *echo.Context) error {
	idStr := c.PathParam("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.ResponseWithCode(c, 9002, err)
	}

	server, err := ctrl.serverUseCase.GetByID(uint(id))
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, server)
}

// Create handles creating a new secondary server
// @Summary Create secondary server
// @Description Create a new secondary server
// @Tags SecondaryServer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body SecondaryServerRequest true "Secondary Server Data"
// @Success 201 {object} models.SecondaryServer
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Router /secondary-servers [post]
func (ctrl *SecondaryServerController) Create(c *echo.Context) error {
	var req SecondaryServerRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	server, err := ctrl.serverUseCase.Create(req.Title, req.IP, req.Port, req.Token)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusCreated, server)
}

// Update handles updating a secondary server
// @Summary Update secondary server
// @Description Update an existing secondary server
// @Tags SecondaryServer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Secondary Server ID"
// @Param request body SecondaryServerRequest true "Updated Secondary Server Data"
// @Success 200 {object} models.SecondaryServer
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /secondary-servers/{id} [put]
func (ctrl *SecondaryServerController) Update(c *echo.Context) error {
	idStr := c.PathParam("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.ResponseWithCode(c, 9002, err)
	}

	var req SecondaryServerRequest
	if err := ctrl.validator.Validate(c, &req); err != nil {
		return err
	}

	server, err := ctrl.serverUseCase.Update(uint(id), req.Title, req.IP, req.Port, req.Token)
	if err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, server)
}

// Delete handles deleting a secondary server
// @Summary Delete secondary server
// @Description Delete a secondary server by ID
// @Tags SecondaryServer
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Secondary Server ID"
// @Success 200 {object} request.MessageResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 404 {object} request.ErrorResponse
// @Router /secondary-servers/{id} [delete]
func (ctrl *SecondaryServerController) Delete(c *echo.Context) error {
	idStr := c.PathParam("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctrl.req.ResponseWithCode(c, 9002, err)
	}

	if err := ctrl.serverUseCase.Delete(uint(id)); err != nil {
		code, parseErr := strconv.Atoi(err.Error())
		if parseErr != nil {
			return ctrl.req.InternalServerError(c, err)
		}
		return ctrl.req.ResponseWithCode(c, code, err)
	}

	return c.JSON(http.StatusOK, request.MessageResponse{
		Message: "Secondary server deleted successfully",
	})
}
