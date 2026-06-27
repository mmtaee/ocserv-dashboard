package home

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
)

type Controller struct {
	request request.CustomRequestInterface
	homeUC  usecase.HomeUsecaseInterface
}

func New(homeUC usecase.HomeUsecaseInterface) *Controller {
	return &Controller{
		request: request.NewCustomRequest(),
		homeUC:  homeUC,
	}
}

// Home 	     Content of home
//
// @Summary      Content of home
// @Description  Content of home
// @Tags         Home
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} usecase.HomeGetHomeResponse
// @Router       /home [get]
func (ctl *Controller) Home(c echo.Context) error {
	resp, err := ctl.homeUC.Home()
	if err != nil {
		logger.Warn("error in Home handler: %v", err)
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, resp)
}

// OcservStats 	     Content of ocserv server stats
//
// @Summary      Content of ocserv server stats
// @Description  Content of ocserv server stats
// @Tags         Home
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} usecase.HomeOcservStatusResponse
// @Router       /home/ocserv-stats [get]
func (ctl *Controller) OcservStats(c echo.Context) error {
	resp, _ := ctl.homeUC.OcservStats()
	return c.JSON(http.StatusOK, resp)
}

// SystemUsageStats Content of os system usage stats
//
// @Summary      Content of os system usage stats
// @Description  Content of os system usage stats (cpu, ram, swap)
// @Tags         Home
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} usecase.HomeServerStatusResponse
// @Router       /home/system-stats [get]
func (ctl *Controller) SystemUsageStats(c echo.Context) error {
	resp, err := ctl.homeUC.SystemUsageStats()
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, resp)
}

// ContainerUsageStats Content of docker system usage stats
//
// @Summary      Content of docker system usage stats
// @Description  Content of docker system usage stats (cpu, ram, swap)
// @Tags         Home
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} usecase.HomeDockerService
// @Router       /home/container-stats [get]
func (ctl *Controller) ContainerUsageStats(c echo.Context) error {
	resp, err := ctl.homeUC.ContainerUsageStats()
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, resp)
}
