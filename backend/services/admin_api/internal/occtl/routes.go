package occtl

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
)

func Routes(e *echo.Group, ctl *Controller) {
	g := e.Group("/occtl")
	g.GET("/server_info", ctl.ServerInfo)
	g.GET("/commands", ctl.Commands, middlewares.AuthMiddleware())
}
