package telegram

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/middlewares"
)

func RegisterRoutes(e *echo.Echo, cfg config.TelegramConfig) {
	telegramRepo := repository.NewTelegramRepository(infra.DB)
	telegramUC := usecase.NewTelegramUseCase(telegramRepo)

	userRepo := repository.NewOcservUserRepository(infra.DB)
	userUC := usecase.NewOcservUserUseCase(userRepo)

	telegramController := NewTelegramController(telegramUC, userUC, userRepo, cfg)

	api := e.Group("/api/v1/telegram")
	api.Use(middlewares.AuthMiddleware())

	// Settings
	api.GET("/settings", telegramController.GetSettings, middlewares.AdminPermissionMiddleware())
	api.PATCH("/settings", telegramController.UpdateSettings, middlewares.AdminPermissionMiddleware())
	api.POST("/test", telegramController.Test, middlewares.AdminPermissionMiddleware())

	// Packages
	api.GET("/packages", telegramController.ListPackages)
	api.POST("/packages", telegramController.CreatePackage, middlewares.AdminPermissionMiddleware())
	api.PATCH("/packages/:id", telegramController.UpdatePackage, middlewares.AdminPermissionMiddleware())
	api.DELETE("/packages/:id", telegramController.DeletePackage, middlewares.AdminPermissionMiddleware())

	// Requests
	api.GET("/requests", telegramController.ListRequests, middlewares.AdminPermissionMiddleware())
	api.GET("/requests/:id", telegramController.GetRequest, middlewares.AdminPermissionMiddleware())
	api.GET("/requests/:id/receipt", telegramController.GetReceipt, middlewares.AdminPermissionMiddleware())
	api.POST("/requests/:id/approve", telegramController.Approve, middlewares.AdminPermissionMiddleware())
	api.POST("/requests/:id/reject", telegramController.Reject, middlewares.AdminPermissionMiddleware())
	api.POST("/requests/:id/confirm-payment", telegramController.ConfirmPayment, middlewares.AdminPermissionMiddleware())
	api.DELETE("/requests/:id", telegramController.DeleteRequest, middlewares.AdminPermissionMiddleware())

	// Linked accounts
	api.GET("/accounts", telegramController.AccountsForOcservUser)
	api.DELETE("/accounts/:id", telegramController.DeleteAccount, middlewares.AdminPermissionMiddleware())
}
