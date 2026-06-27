package ocserv_user

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()
	g := e.Group("/ocserv/users", middlewares.AuthMiddleware())

	g.GET("", ctl.Users)
	g.GET("/:uid", ctl.User)

	g.POST("", ctl.Create)
	g.PATCH("/:uid", ctl.Update)
	g.DELETE("/:uid", ctl.Delete)

	g.POST("/:username/disconnect", ctl.Disconnect)
	g.POST("/:id/disconnect_by_id", ctl.DisconnectSessionById)

	g.POST("/:username/terminate", ctl.Terminate)
	g.POST("/:id/terminate_by_id", ctl.TerminateSessionById)

	g.POST("/:uid/lock", ctl.Lock)
	g.POST("/:uid/unlock", ctl.UnLock)
	g.POST("/:uid/activate", ctl.ActivateExpired)

	g.POST("/:uid/certificate", ctl.CreateCertificate)
	g.GET("/:uid/certificate", ctl.DownloadCertificate)

	g.GET("/:uid/session_logs", ctl.SessionLogs)
	g.GET("/:uid/statistics", ctl.Statistics)

	g.GET("/ocpasswd", ctl.OcpasswdUsers, middlewares.AdminPermission())
	g.POST("/ocpasswd/sync", ctl.SyncToDB, middlewares.AdminPermission())
}
