package auth

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/user"
)

// RegisterRoutes registers all customer auth routes
func RegisterRoutes(e *echo.Echo) {
	ocservUserRepo := repository.NewOcservUserRepository(infra.DB, user.NewOcservUser())
	authUC := usecase.NewAuthUseCase(ocservUserRepo)
	authController := NewAuthController(authUC)

	// Public routes
	public := e.Group("/api/v1/customer/auth")
	public.POST("/login", authController.Login)
}
