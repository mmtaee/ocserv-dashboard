package super_admin

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/middlewares"
)

// RegisterRoutes registers all super admin routes
func RegisterRoutes(e *echo.Echo) {
	adminRepo := repository.NewAdminRepository(infra.DB)
	superAdminUC := usecase.NewSuperAdminUseCase(adminRepo)
	superAdminController := NewSuperAdminController(superAdminUC)

	// Protected super admin routes
	protected := e.Group("/api/v1/super-admin")
	protected.Use(middlewares.AuthMiddleware())
	protected.Use(middlewares.SuperAdminMiddleware())

	adminGroup := protected.Group("/admins")
	adminGroup.POST("", superAdminController.CreateAdmin)
	adminGroup.GET("", superAdminController.ListAdmins)
	adminGroup.GET("/:id", superAdminController.GetAdmin)
	adminGroup.POST("/:id", superAdminController.UpdateAdmin)
	adminGroup.POST("/:id/change-password", superAdminController.ChangeAdminPassword)
	adminGroup.POST("/:id/suspend", superAdminController.SuspendAdmin)
	adminGroup.POST("/:id/unsuspend", superAdminController.UnsuspendAdmin)
}
