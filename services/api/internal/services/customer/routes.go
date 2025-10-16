package customer

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()
	g := e.Group("/customers")
	g.POST("/summary", ctl.Summary, middlewares.RateLimitMiddleware(2, "m", 5))
}
