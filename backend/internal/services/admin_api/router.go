package adminapi

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
)

func (s *Service) Register(e *echo.Group) {
	s.registerAuthRoutes(e)
	s.registerSystemRoutes(e)
	s.registerGroupRoutes(e)
	s.registerUserRoutes(e)
	s.registerOCCTLRoutes(e)
	s.registerDashboardRoutes(e)
	s.registerBackupRoutes(e)
	s.registerReportRoutes(e)
	s.registerRuntimeRoutes(e)
	if s.telegramRoutes {
		s.registerTelegramRoutes(e)
	}
	if !s.agentNode {
		s.registerAgentRoutes(e)
	}
}

func (s *Service) registerAuthRoutes(e *echo.Group) {
	g := e.Group("/auth", s.authenticate)
	g.POST("/logout", s.auth.Logout)
	g.GET("/sessions", s.auth.Sessions, middlewares.SuperadminPermission())
	g.DELETE("/sessions/:id", s.auth.Revoke, middlewares.SuperadminPermission())
}

func (s *Service) registerSystemRoutes(e *echo.Group) {
	public := e.Group("/system")
	public.GET("/release", s.system.DashboardRelease)
	public.GET("/init", s.system.SystemInit)
	public.POST("/users/login", s.system.Login, middlewares.RateLimitMiddleware(2, "m", 3))

	protected := e.Group("/system", s.authenticate)
	protected.GET("", s.system.System)
	protected.GET("/users/profile", s.system.Profile)
	protected.POST("/users/password", s.system.ChangePasswordBySelf)
	protected.POST("/user/reset-password", s.system.ResetAdminPassword, middlewares.SuperadminPermission(), middlewares.RateLimitMiddleware(1, "m", 2))

	admin := e.Group("/system", s.authenticate, middlewares.SuperadminPermission())
	admin.PATCH("", s.system.SystemUpdate)
	admin.POST("/users", s.system.CreateUser)
	admin.GET("/users", s.system.Users)
	admin.GET("/users/lookup", s.system.UsersLookup)
	admin.POST("/users/:id/password", s.system.ChangeUserPasswordByAdmin)
	admin.DELETE("/users/:id", s.system.DeleteUser)
	admin.GET("/ocserv-config", s.runtime.Config)
	admin.PATCH("/ocserv-config", s.runtime.UpdateConfig, middlewares.RateLimitMiddleware(1, "m", 1))
}

func (s *Service) registerGroupRoutes(e *echo.Group) {
	g := e.Group("/ocserv/groups", s.authenticate)
	g.GET("", s.groups.OcservGroups)
	g.GET("/lookup", s.groups.OcservGroupsLookup)
	g.GET("/:id", s.groups.OcservGroup)
	g.POST("", s.groups.CreateOcservGroup, middlewares.SuperadminPermission())
	g.PATCH("/:id", s.groups.UpdateOcservGroup, middlewares.SuperadminPermission())
	g.DELETE("/:id", s.groups.DeleteOcservGroup, middlewares.SuperadminPermission())
	g.GET("/defaults", s.groups.GetDefaultsGroup)
	g.PATCH("/defaults", s.groups.UpdateDefaultsGroup, middlewares.SuperadminPermission())
	g.GET("/unsynced", s.groups.ListUnsyncedGroups, middlewares.SuperadminPermission())
	g.POST("/sync", s.groups.SyncGroup, middlewares.SuperadminPermission())
}

func (s *Service) registerUserRoutes(e *echo.Group) {
	g := e.Group("/ocserv/users", s.authenticate)
	g.GET("", s.users.Users)
	g.PATCH("/bulk", s.users.BulkUpdate)
	g.DELETE("/bulk", s.users.BulkDelete)
	g.PATCH("/bulk/status", s.users.BulkSetEnabled)
	g.PATCH("/bulk/group", s.users.BulkSetGroup)
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
	g.POST("/:id/reset-usage", s.users.ResetUsage)
	g.POST("/:id/activate", s.users.ActivateExpired)
	g.POST("/:id/certificate", s.users.CreateCertificate)
	g.GET("/:id/certificate", s.users.DownloadCertificate)
	g.GET("/:id/session_logs", s.users.SessionLogs)
	g.GET("/:id/statistics", s.users.Statistics)
	g.GET("/ocpasswd", s.users.OcpasswdUsers, middlewares.SuperadminPermission())
	g.POST("/ocpasswd/sync", s.users.SyncToDB, middlewares.SuperadminPermission())
}

func (s *Service) registerOCCTLRoutes(e *echo.Group) {
	g := e.Group("/occtl")
	g.GET("/server_info", s.occtl.ServerInfo)
	g.GET("/commands", s.occtl.Commands, s.authenticate, middlewares.SuperadminPermission())
}

func (s *Service) registerDashboardRoutes(e *echo.Group) {
	g := e.Group("/home", s.authenticate, middlewares.SuperadminPermission())
	g.GET("", s.dashboard.Home)
	g.GET("/ocserv-stats", s.dashboard.OcservStats)
	g.GET("/system-stats", s.dashboard.SystemUsageStats)
	g.GET("/container-stats", s.dashboard.ContainerUsageStats)
}

func (s *Service) registerBackupRoutes(e *echo.Group) {
	g := e.Group("/backup", s.authenticate, middlewares.SuperadminPermission())
	g.GET("/ocserv_groups", s.backup.OcservGroupBackup)
	g.POST("/ocserv_groups", s.backup.OcservGroupRestore)
	g.GET("/ocserv_users", s.backup.OcservUserBackup)
	g.POST("/ocserv_users", s.backup.OcservUserRestore)
}

func (s *Service) registerReportRoutes(e *echo.Group) {
	g := e.Group("/reports", s.authenticate, middlewares.SuperadminPermission())
	g.GET("/session_logs", s.reports.SessionLogs)
	g.GET("/statistics", s.reports.Statistics)
	g.GET("/users", s.reports.OcservUserReport)
	g.GET("/total-bandwidth", s.reports.TotalBandwidth)
}

func (s *Service) registerRuntimeRoutes(e *echo.Group) {
	g := e.Group("/systemd", s.authenticate, middlewares.SuperadminPermission())
	g.GET("/status", s.runtime.Status)
	g.POST("/restart", s.runtime.Restart, middlewares.RateLimitMiddleware(1, "m", 1))
	g.POST("/disable", s.runtime.Disable, middlewares.RateLimitMiddleware(1, "m", 1))
	g.POST("/enable", s.runtime.Enable, middlewares.RateLimitMiddleware(1, "m", 1))
}

func (s *Service) registerTelegramRoutes(e *echo.Group) {
	g := e.Group("/telegram", s.authenticate)
	g.GET("/settings", s.telegram.GetSettings)
	g.PATCH("/settings", s.telegram.UpdateSettings, middlewares.SuperadminPermission())
	g.POST("/test", s.telegram.Test, middlewares.SuperadminPermission())
	g.GET("/packages", s.telegram.ListPackages)
	g.POST("/packages", s.telegram.CreatePackage, middlewares.SuperadminPermission())
	g.PATCH("/packages/:id", s.telegram.UpdatePackage, middlewares.SuperadminPermission())
	g.DELETE("/packages/:id", s.telegram.DeletePackage, middlewares.SuperadminPermission())
	g.GET("/requests", s.telegram.ListRequests, middlewares.SuperadminPermission())
	g.GET("/requests/:id", s.telegram.GetRequest, middlewares.SuperadminPermission())
	g.GET("/requests/:id/receipt", s.telegram.GetReceipt, middlewares.SuperadminPermission())
	g.POST("/requests/:id/approve", s.telegram.Approve, middlewares.SuperadminPermission())
	g.POST("/requests/:id/reject", s.telegram.Reject, middlewares.SuperadminPermission())
	g.POST("/requests/:id/confirm-payment", s.telegram.ConfirmPayment, middlewares.SuperadminPermission())
	g.DELETE("/requests/:id", s.telegram.DeleteRequest, middlewares.SuperadminPermission())
	g.GET("/accounts", s.telegram.AccountsForOcservUser)
	g.DELETE("/accounts/:id", s.telegram.DeleteAccount, middlewares.SuperadminPermission())
}

func (s *Service) registerAgentRoutes(e *echo.Group) {
	g := e.Group("/ocserv/agents", s.authenticate, middlewares.SuperadminPermission())
	g.GET("", s.agents.List)
	g.GET("/:id", s.agents.Get)
	g.POST("", s.agents.Create)
	g.PATCH("/:id", s.agents.Update)
	g.DELETE("/:id", s.agents.Delete)
}
