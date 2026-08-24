package adminapi

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
)

func (s *Service) Register(e *echo.Group) {
	s.registerSystemRoutes(e)
	s.registerGroupRoutes(e)
	s.registerUserRoutes(e)
	s.registerOCCTLRoutes(e)
	s.registerDashboardRoutes(e)
	s.registerBackupRoutes(e)
	s.registerReportRoutes(e)
	s.registerSystemdRoutes(e)
	if s.telegramRoutes {
		s.registerTelegramRoutes(e)
	}
}

func (s *Service) registerSystemRoutes(e *echo.Group) {
	public := e.Group("/system")
	public.GET("/release", s.system.DashboardRelease)
	public.GET("/init", s.system.SystemInit)
	public.POST("/setup", s.system.SetupSystem)
	public.POST("/user/reset-password", s.system.ResetAdminPassword, middlewares.RateLimitMiddleware(1, "m", 2))
	public.POST("/users/login", s.system.Login, middlewares.RateLimitMiddleware(2, "m", 3))

	protected := e.Group("/system", middlewares.AuthMiddleware())
	protected.GET("", s.system.System)
	protected.GET("/users/profile", s.system.Profile)
	protected.POST("/users/password", s.system.ChangePasswordBySelf)

	admin := e.Group("/system", middlewares.AuthMiddleware(), middlewares.AdminPermission())
	admin.PATCH("", s.system.SystemUpdate)
	admin.POST("/users", s.system.CreateUser)
	admin.GET("/users", s.system.Users)
	admin.GET("/users/lookup", s.system.UsersLookup)
	admin.POST("/users/:id/password", s.system.ChangeUserPasswordByAdmin)
	admin.DELETE("/users/:id", s.system.DeleteUser)
}

func (s *Service) registerGroupRoutes(e *echo.Group) {
	g := e.Group("/ocserv/groups", middlewares.AuthMiddleware())
	g.GET("", s.groups.OcservGroups)
	g.GET("/lookup", s.groups.OcservGroupsLookup)
	g.GET("/:id", s.groups.OcservGroup)
	g.POST("", s.groups.CreateOcservGroup)
	g.PATCH("/:id", s.groups.UpdateOcservGroup)
	g.DELETE("/:id", s.groups.DeleteOcservGroup)
	g.GET("/defaults", s.groups.GetDefaultsGroup, middlewares.AdminPermission())
	g.PATCH("/defaults", s.groups.UpdateDefaultsGroup, middlewares.AdminPermission())
	g.GET("/unsynced", s.groups.ListUnsyncedGroups, middlewares.AdminPermission())
	g.POST("/sync", s.groups.SyncGroup, middlewares.AdminPermission())
}

func (s *Service) registerUserRoutes(e *echo.Group) {
	g := e.Group("/ocserv/users", middlewares.AuthMiddleware())
	g.GET("", s.users.Users)
	g.GET("/:id", s.users.User)
	g.POST("", s.users.Create)
	g.PATCH("/:id", s.users.Update)
	g.DELETE("/:id", s.users.Delete)
	g.POST("/:username/disconnect", s.users.Disconnect)
	g.POST("/:id/disconnect_by_id", s.users.DisconnectSessionById)
	g.POST("/:username/terminate", s.users.Terminate)
	g.POST("/:id/terminate_by_id", s.users.TerminateSessionById)
	g.POST("/:id/lock", s.users.Lock)
	g.POST("/:id/unlock", s.users.UnLock)
	g.POST("/:id/activate", s.users.ActivateExpired)
	g.POST("/:id/certificate", s.users.CreateCertificate)
	g.GET("/:id/certificate", s.users.DownloadCertificate)
	g.GET("/:id/session_logs", s.users.SessionLogs)
	g.GET("/:id/statistics", s.users.Statistics)
	g.GET("/ocpasswd", s.users.OcpasswdUsers, middlewares.AdminPermission())
	g.POST("/ocpasswd/sync", s.users.SyncToDB, middlewares.AdminPermission())
}

func (s *Service) registerOCCTLRoutes(e *echo.Group) {
	g := e.Group("/occtl")
	g.GET("/server_info", s.occtl.ServerInfo)
	g.GET("/commands", s.occtl.Commands, middlewares.AuthMiddleware())
}

func (s *Service) registerDashboardRoutes(e *echo.Group) {
	g := e.Group("/home", middlewares.AuthMiddleware())
	g.GET("", s.dashboard.Home)
	g.GET("/ocserv-stats", s.dashboard.OcservStats)
	g.GET("/system-stats", s.dashboard.SystemUsageStats)
	g.GET("/container-stats", s.dashboard.ContainerUsageStats)
}

func (s *Service) registerBackupRoutes(e *echo.Group) {
	g := e.Group("/backup", middlewares.AuthMiddleware(), middlewares.AdminPermission())
	g.GET("/ocserv_groups", s.backup.OcservGroupBackup)
	g.POST("/ocserv_groups", s.backup.OcservGroupRestore)
	g.GET("/ocserv_users", s.backup.OcservUserBackup)
	g.POST("/ocserv_users", s.backup.OcservUserRestore)
}

func (s *Service) registerReportRoutes(e *echo.Group) {
	g := e.Group("/reports", middlewares.AuthMiddleware(), middlewares.AdminPermission())
	g.GET("/session_logs", s.reports.SessionLogs)
	g.GET("/statistics", s.reports.Statistics)
	g.GET("/users", s.reports.OcservUserReport)
	g.GET("/total-bandwidth", s.reports.TotalBandwidth)
}

func (s *Service) registerSystemdRoutes(e *echo.Group) {
	g := e.Group("/systemd", middlewares.AuthMiddleware(), middlewares.AdminPermission())
	g.GET("/status", s.systemd.Status)
	g.POST("/restart", s.systemd.Restart, middlewares.RateLimitMiddleware(1, "m", 1))
	g.POST("/disable", s.systemd.Disable, middlewares.RateLimitMiddleware(1, "m", 1))
	g.POST("/enable", s.systemd.Enable, middlewares.RateLimitMiddleware(1, "m", 1))
}

func (s *Service) registerTelegramRoutes(e *echo.Group) {
	g := e.Group("/telegram", middlewares.AuthMiddleware())
	g.GET("/settings", s.telegram.GetSettings, middlewares.AdminPermission())
	g.PATCH("/settings", s.telegram.UpdateSettings, middlewares.AdminPermission())
	g.POST("/test", s.telegram.Test, middlewares.AdminPermission())
	g.GET("/packages", s.telegram.ListPackages)
	g.POST("/packages", s.telegram.CreatePackage, middlewares.AdminPermission())
	g.PATCH("/packages/:id", s.telegram.UpdatePackage, middlewares.AdminPermission())
	g.DELETE("/packages/:id", s.telegram.DeletePackage, middlewares.AdminPermission())
	g.GET("/requests", s.telegram.ListRequests, middlewares.AdminPermission())
	g.GET("/requests/:id", s.telegram.GetRequest, middlewares.AdminPermission())
	g.GET("/requests/:id/receipt", s.telegram.GetReceipt, middlewares.AdminPermission())
	g.POST("/requests/:id/approve", s.telegram.Approve, middlewares.AdminPermission())
	g.POST("/requests/:id/reject", s.telegram.Reject, middlewares.AdminPermission())
	g.POST("/requests/:id/confirm-payment", s.telegram.ConfirmPayment, middlewares.AdminPermission())
	g.DELETE("/requests/:id", s.telegram.DeleteRequest, middlewares.AdminPermission())
	g.GET("/accounts", s.telegram.AccountsForOcservUser)
	g.DELETE("/accounts/:id", s.telegram.DeleteAccount, middlewares.AdminPermission())
}
