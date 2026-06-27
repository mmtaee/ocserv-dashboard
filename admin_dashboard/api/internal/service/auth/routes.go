package auth

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/middlewares"
)

// RegisterRoutes registers all auth routes
func RegisterRoutes(e *echo.Echo) {
	adminRepo := repository.NewAdminRepository(infra.DB)
	adminUC := usecase.NewAdminUseCase(adminRepo)
	authController := NewAuthController(adminUC)

	// Public routes
	public := e.Group("/api/v1/auth")
	public.POST("/login", authController.Login)

	// Protected routes
	protected := e.Group("/api/v1/auth")
	protected.Use(middlewares.AuthMiddleware())
	protected.GET("/profile", authController.GetProfile)
	protected.POST("/change-password", authController.ChangePassword)
	protected.POST("/logout", authController.Logout)
}
