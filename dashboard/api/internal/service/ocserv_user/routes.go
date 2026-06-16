package ocserv_user

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/middlewares"
)

func RegisterRoutes(e *echo.Echo) {
	userRepo := repository.NewOcservUserRepository(infra.DB)
	userUC := usecase.NewOcservUserUseCase(userRepo)
	userController := NewOcservUserController(userUC)

	g := e.Group("/api/v1/ocserv/users")
	g.Use(middlewares.AuthMiddleware())

	g.GET("", userController.ListUsers)
	g.GET("/:id", userController.GetUser)
	g.POST("", userController.CreateUser)
	g.POST("/:id", userController.UpdateUser)
	g.DELETE("/:id", userController.DeleteUser)
	g.POST("/:id/lock", userController.LockUser)
	g.POST("/:id/unlock", userController.UnlockUser)
	g.POST("/online-sessions", userController.GetOnlineSessions)
}
