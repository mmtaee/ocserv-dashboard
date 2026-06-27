package system

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()

	// =========================
	// Public system routes
	// =========================
	system := e.Group("/system")

	system.GET("/release", ctl.DashboardRelease)
	system.GET("/init", ctl.SystemInit)
	system.POST("/setup", ctl.SetupSystem)
	system.POST(
		"/user/reset-password",
		ctl.ResetAdminPassword,
		middlewares.RateLimitMiddleware(1, "m", 2),
	)
	system.POST(
		"/users/login",
		ctl.Login,
		middlewares.RateLimitMiddleware(2, "m", 3),
	)

	// =========================
	// Protected system routes
	// =========================
	protected := e.Group("/system", middlewares.AuthMiddleware())

	protected.GET("", ctl.System)
	protected.GET("/users/profile", ctl.Profile)
	protected.POST("/users/password", ctl.ChangePasswordBySelf)

	// =========================
	// Admin-only routes
	// =========================
	admin := e.Group("/system", middlewares.AuthMiddleware(), middlewares.AdminPermission())

	admin.PATCH("", ctl.SystemUpdate)

	admin.POST("/users", ctl.CreateUser)
	admin.GET("/users", ctl.Users)
	admin.GET("/users/lookup", ctl.UsersLookup)

	admin.POST("/users/:uid/password", ctl.ChangeUserPasswordByAdmin)
	admin.DELETE("/users/:uid", ctl.DeleteUser)
}
