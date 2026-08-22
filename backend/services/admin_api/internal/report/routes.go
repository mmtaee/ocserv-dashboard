package report

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
)

func Routes(e *echo.Group, ctl *Controller) {
	g := e.Group("/reports", middlewares.AuthMiddleware(), middlewares.AdminPermission())

	g.GET("/session_logs", ctl.SessionLogs)
	g.GET("/statistics", ctl.Statistics)
	g.GET("/users", ctl.OcservUserReport)
	g.GET("/total-bandwidth", ctl.TotalBandwidth)
}
