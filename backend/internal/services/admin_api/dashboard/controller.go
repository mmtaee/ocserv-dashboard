package dashboard

import (
	"net/http"

	"github.com/labstack/echo/v5"
	dashboardusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/dashboard"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request   request.CustomRequestInterface
	dashboard *dashboardusecase.Usecase
}

func New(usecase *dashboardusecase.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), dashboard: usecase}
}

// Home returns the admin dashboard snapshot.
//
// @Summary Content of home
// @Description Content of home
// @Tags Home
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Success 200 {object} GetHomeResponse
// @Router /home [get]
func (ctl *Controller) Home(c *echo.Context) error {
	result, err := ctl.dashboard.Home(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// OcservStats returns Ocserv runtime statistics.
//
// @Summary Content of ocserv server stats
// @Description Content of ocserv server stats
// @Tags Home
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Success 200 {object} OcservStatusResponse
// @Router /home/ocserv-stats [get]
func (ctl *Controller) OcservStats(c *echo.Context) error {
	return c.JSON(http.StatusOK, ctl.dashboard.OcservStats())
}

// SystemUsageStats returns host resource usage.
//
// @Summary Content of os system usage stats
// @Description Content of os system usage stats (cpu, ram, swap)
// @Tags Home
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Success 200 {object} ServerStatusResponse
// @Router /home/system-stats [get]
func (ctl *Controller) SystemUsageStats(c *echo.Context) error {
	result, err := ctl.dashboard.SystemUsage(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// ContainerUsageStats returns selected container resource usage.
//
// @Summary Content of docker system usage stats
// @Description Content of docker system usage stats (cpu, ram, swap)
// @Tags Home
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} middlewares.Unauthorized
// @Success 200 {object} DockerService
// @Router /home/container-stats [get]
func (ctl *Controller) ContainerUsageStats(c *echo.Context) error {
	result, err := ctl.dashboard.ContainerUsage(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}
