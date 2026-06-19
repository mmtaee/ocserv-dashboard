package report

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/middlewares"
)

func RegisterRoutes(e *echo.Echo) {
	reportRepo := repository.NewReportRepository(infra.DB)
	reportUC := usecase.NewReportUseCase(reportRepo)
	reportController := NewReportController(reportUC)

	g := e.Group("/api/v1/reports")
	g.Use(middlewares.AuthMiddleware())

	g.GET("/session_logs", reportController.SessionLogs)
	g.GET("/statistics", reportController.Statistics)
	g.GET("/users", reportController.OcservUserReport)
	g.GET("/total-bandwidth", reportController.TotalBandwidth)
}
