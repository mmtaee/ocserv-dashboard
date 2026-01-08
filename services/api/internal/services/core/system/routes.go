package system

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()

	// --------------------
	// Public /system routes
	// --------------------
	g := e.Group("/system")

	g.GET("/init", ctl.SystemInit)
	g.GET("/permissions", ctl.AvailablePermissions)

	g.POST("/setup", ctl.SetupSystem, middlewares.RateLimitMiddleware(2, "h", 3))

	// --------------------
	// Authenticated /system routes
	// --------------------
	gAuth := g.Group("", middlewares.AuthMiddleware())

	gAuth.PATCH("", ctl.SystemUpdate, middlewares.SuperAdminPermission())
	gAuth.GET("", ctl.System)
}
