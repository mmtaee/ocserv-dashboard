package system

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group) {
	ctl := New()

	// --------------------
	// Public users routes
	// --------------------
	g := e.Group("/users")

	g.POST("/login", ctl.Login, middlewares.RateLimitMiddleware(2, "m", 3))

	// --------------------
	// Authenticated /users routes
	// --------------------
	gAuth := g.Group("", middlewares.AuthMiddleware())
	gAuth.POST("/password", ctl.ChangePasswordBySelf)
	gAuth.GET("/profile", ctl.Profile)

	// --------------------
	// Admin / SuperAdmin users routes
	// --------------------
	adminGroup := gAuth.Group("", middlewares.SuperAdminOrAdminPermission())

	adminGroup.POST("", ctl.CreateUser)
	adminGroup.POST("/:uid/password", ctl.ChangeUserPasswordByAdmin)
	adminGroup.DELETE("/:uid", ctl.DeleteUser)
	adminGroup.GET("", ctl.Users)
	adminGroup.GET("/lookup", ctl.UsersLookup)
	adminGroup.PATCH("/:uid/permissions", ctl.UpdateUserPermission)
}
