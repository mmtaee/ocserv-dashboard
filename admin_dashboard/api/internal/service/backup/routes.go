package backup

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/middlewares"
)

func RegisterRoutes(e *echo.Echo) {
	backupRepo := repository.NewBackupRepository(infra.DB)
	backupUC := usecase.NewBackupUseCase(backupRepo)

	ocservGroupRepo := repository.NewOcservGroupRepository(infra.DB)
	ocservGroupUC := usecase.NewOcservGroupUseCase(ocservGroupRepo)

	backupController := NewBackupController(backupUC, ocservGroupUC, ocservGroupRepo)

	protected := e.Group("/api/v1/backup")
	protected.Use(middlewares.AuthMiddleware())
	protected.Use(middlewares.AdminPermissionMiddleware())

	protected.GET("/ocserv_groups", backupController.OcservGroupBackup)
	protected.POST("/ocserv_groups", backupController.OcservGroupRestore)

	protected.GET("/ocserv_users", backupController.OcservUserBackup)
	protected.POST("/ocserv_users", backupController.OcservUserRestore)
}
