package systemd

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/middlewares"
)

func RegisterRoutes(e *echo.Echo) {
	systemdRepo := repository.NewSystemdRepository("ocserv")
	systemdUC := usecase.NewSystemdUseCase(systemdRepo)
	systemdController := NewSystemdController(systemdUC)

	protected := e.Group("/api/v1/systemd")
	protected.Use(middlewares.AuthMiddleware())
	protected.Use(middlewares.AdminPermissionMiddleware())

	protected.GET("/status", systemdController.Status)
	protected.POST("/restart", systemdController.Restart)
	protected.POST("/disable", systemdController.Disable)
	protected.POST("/enable", systemdController.Enable)
	protected.GET("/main-config", systemdController.GetMainConfig)
	protected.PUT("/main-config", systemdController.UpdateMainConfig)
}
