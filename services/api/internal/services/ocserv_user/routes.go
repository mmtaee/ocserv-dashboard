package ocserv_user

import (
	"github.com/labstack/echo/v4"
	apiModels "github.com/mmtaee/ocserv-users-management/api/internal/models"
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()

	// Base group (Auth only)
	g := e.Group("/ocserv/users", middlewares.AuthMiddleware())

	// =========================
	// CRUD permissions (staff)
	// =========================
	crudGroup := g.Group("", middlewares.StaffPermissionMiddleware(apiModels.OcservUsersCRUDService))

	crudGroup.GET("", ctl.OcservUsers)
	crudGroup.GET("/:uid", ctl.OcservUser)
	crudGroup.POST("", ctl.CreateOcservUser)
	crudGroup.PATCH("/:uid", ctl.UpdateOcservUser)
	crudGroup.DELETE("/:uid", ctl.DeleteOcservUser)

	// =========================
	// Action permissions (staff)
	// =========================
	actionGroup := g.Group("", middlewares.StaffPermissionMiddleware(apiModels.OcservUsersActionService))

	actionGroup.POST("/:uid/lock", ctl.LockOcservUser)
	actionGroup.POST("/:uid/unlock", ctl.UnLockOcservUser)
	actionGroup.POST("/:username/disconnect", ctl.DisconnectOcservUser)
	actionGroup.POST("/:uid/activate", ctl.ActivateExpiredOcservUsers)

	// =========================
	// Statistics permissions (staff)
	// =========================
	statsGroup := g.Group("", middlewares.StaffPermissionMiddleware(apiModels.OcservUserStatsService))

	statsGroup.GET("/:uid/statistics", ctl.StatisticsOcservUser)

	// =========================
	// Super Admin or Admin only
	// =========================
	adminGroup := g.Group("", middlewares.SuperAdminOrAdminPermission())

	adminGroup.GET("/statistics", ctl.Statistics)
	adminGroup.GET("/total-bandwidth", ctl.TotalBandwidth)

	// =========================
	// SuperAdmin only
	// =========================
	superAdminGroup := g.Group("", middlewares.SuperAdminPermission())

	superAdminGroup.GET("/ocpasswd", ctl.OcpasswdUsers)
	superAdminGroup.POST("/ocpasswd/sync", ctl.SyncToDB)
}
