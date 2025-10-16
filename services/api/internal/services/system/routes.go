package system

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()
	e.GET("/system/init", ctl.SystemInit)
	e.POST("/system/setup", ctl.SetupSystem)
	e.POST("/system/users/login", ctl.Login, middlewares.RateLimitMiddleware(2, "m", 3))

	g := e.Group("/system", middlewares.AuthMiddleware())
	g.GET("", ctl.System)
	g.POST("/users/password", ctl.ChangePasswordBySelf)
	g.GET("/users/profile", ctl.Profile)

	g.PATCH("", ctl.SystemUpdate, middlewares.AdminPermission())
	g.POST("/users", ctl.CreateUser, middlewares.AdminPermission())
	g.POST("/users/:uid/password", ctl.ChangeUserPasswordByAdmin, middlewares.AdminPermission())
	g.DELETE("/users/:uid", ctl.DeleteUser, middlewares.AdminPermission())
	g.GET("/users", ctl.Users, middlewares.AdminPermission())
	g.GET("/users/lookup", ctl.UsersLookup, middlewares.AdminPermission())
}
