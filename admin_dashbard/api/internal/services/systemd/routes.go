package systemd

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()
	g := e.Group("/systemd", middlewares.AuthMiddleware(), middlewares.AdminPermission())

	g.GET("/status", ctl.Status)
	g.POST("/restart",
		ctl.Restart,
		middlewares.RateLimitMiddleware(1, "m", 1),
	)
	g.POST("/disable",
		ctl.Disable,
		middlewares.RateLimitMiddleware(1, "m", 1),
	)
	g.POST("/enable",
		ctl.Enable,
		middlewares.RateLimitMiddleware(1, "m", 1),
	)
}
