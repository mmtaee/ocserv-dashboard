package customerapi

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
)

func (s *Service) Register(e *echo.Group) {
	g := e.Group("/customers")
	g.POST("/summary", s.controller.Summary, middlewares.RateLimitMiddleware(2, "m", 5))
	g.POST("/certificate", s.controller.DownloadCertificate, middlewares.RateLimitMiddleware(2, "m", 5))
	g.POST("/setup/cisco", s.ciscoSetup.CiscoSetup, middlewares.RateLimitMiddleware(2, "m", 5))
	g.GET("/setup/cisco/certificate/:token", s.ciscoSetup.DownloadCiscoSetupCertificate, middlewares.RateLimitMiddleware(10, "m", 20))
	g.POST("/disconnect_sessions", s.controller.DisconnectSessions, middlewares.RateLimitMiddleware(1, "m", 2))
}
