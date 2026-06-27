package report

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()
	g := e.Group("/reports", middlewares.AuthMiddleware(), middlewares.AdminPermission())

	g.GET("/session_logs", ctl.SessionLogs)
	g.GET("/statistics", ctl.Statistics)
	g.GET("/users", ctl.OcservUserReport)
	g.GET("/total-bandwidth", ctl.TotalBandwidth)
}
