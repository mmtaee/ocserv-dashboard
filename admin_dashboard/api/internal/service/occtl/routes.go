package occtl

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/middlewares"
)

func RegisterRoutes(e *echo.Echo) {
	controller := NewOcctlController()

	e.GET("/occtl/server_info", controller.ServerInfo)

	g := e.Group("/api/v1/occtl")
	g.Use(middlewares.AuthMiddleware())
	g.GET("/commands", controller.Commands)
}
