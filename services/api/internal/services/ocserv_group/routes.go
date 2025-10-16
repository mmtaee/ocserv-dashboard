package ocserv_group

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()
	g := e.Group("/ocserv/groups", middlewares.AuthMiddleware())
	g.GET("", ctl.OcservGroups)
	g.GET("/lookup", ctl.OcservGroupsLookup)
	g.GET("/:id", ctl.OcservGroup)
	g.POST("", ctl.CreateOcservGroup)
	g.PATCH("/:id", ctl.UpdateOcservGroup)
	g.DELETE("/:id", ctl.DeleteOcservGroup)
	g.GET("/defaults", ctl.GetDefaultsGroup)
	g.PATCH("/defaults", ctl.UpdateDefaultsGroup)
}
