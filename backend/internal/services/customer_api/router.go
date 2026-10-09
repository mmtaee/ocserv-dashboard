package customerapi

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
)

func (s *Service) Register(e *echo.Group) {
	public := e.Group("/customers")
	public.POST("/login", s.controller.Login, middlewares.RateLimitMiddleware(2, "m", 5))

	g := e.Group("/customers", s.authenticate)
	g.GET("/summary", s.controller.Summary)
	g.GET("/sessions", s.controller.Sessions)
	g.POST("/disconnect_sessions", s.controller.DisconnectSessions, middlewares.RateLimitMiddleware(1, "m", 2))
	g.POST("/terminate_sessions", s.controller.TerminateSessions, middlewares.RateLimitMiddleware(1, "m", 2))
	g.POST("/password", s.controller.ChangePassword, middlewares.RateLimitMiddleware(2, "m", 5))
	g.GET("/stats", s.controller.Stats)
	g.GET("/activities", s.controller.Activities)
	g.GET("/bandwidth", s.controller.Bandwidth)
	g.GET("/certificate", s.controller.DownloadCertificate, middlewares.RateLimitMiddleware(2, "m", 5))
	g.GET("/setup/cisco", s.controller.CiscoSetup, middlewares.RateLimitMiddleware(2, "m", 5))
	g.GET("/setup/cisco/certificate", s.controller.DownloadCiscoSetupCertificate, middlewares.RateLimitMiddleware(10, "m", 20))
}
