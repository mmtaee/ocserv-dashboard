package system

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/middlewares"
)

// RegisterRoutes registers all system routes
func RegisterRoutes(e *echo.Echo) {
	systemRepo := repository.NewSystemRepository(infra.DB)
	systemUC := usecase.NewSystemUseCase(systemRepo)
	systemController := NewSystemController(systemUC)

	// Protected routes
	protected := e.Group("/api/v1/system")
	protected.Use(middlewares.AuthMiddleware())
	protected.GET("", systemController.GetSystem)
	protected.POST("", systemController.UpdateSystem)
}
