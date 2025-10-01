package ocserv_user

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()
	g := e.Group("/ocserv/users", middlewares.AuthMiddleware())

	g.GET("", ctl.OcservUsers)
	g.GET("/:uid", ctl.OcservUser)
	g.POST("", ctl.CreateOcservUser)
	g.PATCH("/:uid", ctl.UpdateOcservUser)
	g.DELETE("/:uid", ctl.DeleteOcservUser)
	g.POST("/:uid/lock", ctl.LockOcservUser)
	g.POST("/:uid/unlock", ctl.UnLockOcservUser)
	g.POST("/:username/disconnect", ctl.DisconnectOcservUser)
	g.GET("/:uid/statistics", ctl.StatisticsOcservUser)
	g.GET("/statistics", ctl.Statistics)
	g.GET("/total-bandwidth", ctl.TotalBandwidth)
}
