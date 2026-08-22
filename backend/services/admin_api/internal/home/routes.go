package home

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
)

func Routes(e *echo.Group, ctl *Controller) {
	g := e.Group("/home", middlewares.AuthMiddleware())

	g.GET("", ctl.Home)
	g.GET("/ocserv-stats", ctl.OcservStats)
	g.GET("/system-stats", ctl.SystemUsageStats)
	g.GET("/container-stats", ctl.ContainerUsageStats)
}
