package customer

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()
	g := e.Group("/customers")
	g.POST("/summary", ctl.Summary, middlewares.RateLimitMiddleware(2, "m", 5))
	g.POST("/certificate", ctl.DownloadCertificate, middlewares.RateLimitMiddleware(2, "m", 5))
	g.POST("/setup/cisco", ctl.CiscoSetup, middlewares.RateLimitMiddleware(2, "m", 5))
	g.GET("/setup/cisco/certificate/:token", ctl.DownloadCiscoSetupCertificate, middlewares.RateLimitMiddleware(10, "m", 20))
	g.POST("/disconnect_sessions", ctl.DisconnectSessions, middlewares.RateLimitMiddleware(1, "m", 2))
}
