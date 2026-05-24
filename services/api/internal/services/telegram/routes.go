package telegram

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing/middlewares"
)

// Routes registers the admin-side Telegram management endpoints. All routes
// require authentication; mutations and request handling require admin role.
func Routes(e *echo.Group) {
	ctl := New()

	g := e.Group("/telegram", middlewares.AuthMiddleware())

	// Settings
	g.GET("/settings", ctl.GetSettings, middlewares.AdminPermission())
	g.PATCH("/settings", ctl.UpdateSettings, middlewares.AdminPermission())
	g.POST("/test", ctl.Test, middlewares.AdminPermission())

	// Packages
	g.GET("/packages", ctl.ListPackages)
	g.POST("/packages", ctl.CreatePackage, middlewares.AdminPermission())
	g.PATCH("/packages/:id", ctl.UpdatePackage, middlewares.AdminPermission())
	g.DELETE("/packages/:id", ctl.DeletePackage, middlewares.AdminPermission())

	// Requests
	g.GET("/requests", ctl.ListRequests, middlewares.AdminPermission())
	g.GET("/requests/:id", ctl.GetRequest, middlewares.AdminPermission())
	g.GET("/requests/:id/receipt", ctl.GetReceipt, middlewares.AdminPermission())
	g.POST("/requests/:id/approve", ctl.Approve, middlewares.AdminPermission())
	g.POST("/requests/:id/reject", ctl.Reject, middlewares.AdminPermission())
	g.POST("/requests/:id/confirm-payment", ctl.ConfirmPayment, middlewares.AdminPermission())
	g.DELETE("/requests/:id", ctl.DeleteRequest, middlewares.AdminPermission())

	// Linked accounts
	g.GET("/accounts", ctl.AccountsForOcservUser)
	g.DELETE("/accounts/:id", ctl.DeleteAccount, middlewares.AdminPermission())
}
