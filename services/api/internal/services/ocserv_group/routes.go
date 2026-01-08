package ocserv_group

import (
	"github.com/labstack/echo/v4"
	apiModels "github.com/mmtaee/ocserv-users-management/api/internal/models"
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()

	// Base group (auth only)
	g := e.Group("/ocserv/groups", middlewares.AuthMiddleware())

	// =========================
	// CRUD permissions (staff)
	// =========================
	actionGroup := g.Group("", middlewares.StaffPermissionMiddleware(apiModels.OcservGroupsCRUDService))

	actionGroup.GET("", ctl.OcservGroups)
	actionGroup.GET("/lookup", ctl.OcservGroupsLookup)
	actionGroup.GET("/:id", ctl.OcservGroup)
	actionGroup.POST("", ctl.CreateOcservGroup)
	actionGroup.PATCH("/:id", ctl.UpdateOcservGroup)
	actionGroup.DELETE("/:id", ctl.DeleteOcservGroup)

	// =========================
	// SuperAdmin only
	// =========================
	superadminGroup := g.Group("", middlewares.SuperAdminPermission())

	superadminGroup.GET("/defaults", ctl.GetDefaultsGroup)
	superadminGroup.PATCH("/defaults", ctl.UpdateDefaultsGroup)
	superadminGroup.GET("/unsynced", ctl.ListUnsyncedGroups)
	superadminGroup.POST("/sync", ctl.SyncGroup)
}
