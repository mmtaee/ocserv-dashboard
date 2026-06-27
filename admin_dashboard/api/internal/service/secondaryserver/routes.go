package secondaryserver

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/middlewares"
)

// RegisterRoutes registers all secondary server routes
func RegisterRoutes(e *echo.Echo) {
	serverRepo := repository.NewSecondaryServerRepository(infra.DB)
	serverUC := usecase.NewSecondaryServerUseCase(serverRepo)
	serverController := NewSecondaryServerController(serverUC)

	protected := e.Group("/api/v1/secondary-servers")
	protected.Use(middlewares.AuthMiddleware(), middlewares.SuperAdminMiddleware())
	protected.GET("", serverController.List)
	protected.GET("/:id", serverController.Get)
	protected.POST("", serverController.Create)
	protected.PUT("/:id", serverController.Update)
	protected.DELETE("/:id", serverController.Delete)
}
