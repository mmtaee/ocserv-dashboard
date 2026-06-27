package home

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()
	g := e.Group("/home", middlewares.AuthMiddleware())

	g.GET("", ctl.Home)
	g.GET("/ocserv-stats", ctl.OcservStats)
	g.GET("/system-stats", ctl.SystemUsageStats)
	g.GET("/container-stats", ctl.ContainerUsageStats)
}
