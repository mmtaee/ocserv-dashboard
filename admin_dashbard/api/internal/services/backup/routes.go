package backup

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()
	g := e.Group("/backup", middlewares.AuthMiddleware(), middlewares.AdminPermission())

	g.GET("/ocserv_groups", ctl.OcservGroupBackup)
	g.POST("/ocserv_groups", ctl.OcservGroupRestore)

	g.GET("/ocserv_users", ctl.OcservUserBackup)
	g.POST("/ocserv_users", ctl.OcservUserRestore)
}
