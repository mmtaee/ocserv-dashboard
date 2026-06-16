package ocserv_group

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/middlewares"
)

// RegisterRoutes registers all Ocserv Group routes
func RegisterRoutes(e *echo.Echo) {
	groupRepo := repository.NewOcservGroupRepository(infra.DB)
	userRepo := repository.NewOcservUserRepository(infra.DB)
	groupUC := usecase.NewOcservGroupUseCase(groupRepo, userRepo)
	groupController := NewOcservGroupController(groupUC)

	group := e.Group("/api/v1/ocserv/groups")
	group.Use(middlewares.AuthMiddleware())

	group.GET("", groupController.ListGroups)
	group.GET("/lookup", groupController.GroupsLookup)
	group.GET("/:id", groupController.GetGroup)
	group.POST("", groupController.CreateGroup)
	group.POST("/:id", groupController.UpdateGroup)
	group.DELETE("/:id", groupController.DeleteGroup)
}
