package telegram

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing/middlewares"
)

func Routes(e *echo.Group, ctl *Controller) {
	g := e.Group("/telegram", middlewares.AuthMiddleware())

	g.GET("/settings", ctl.GetSettings, middlewares.AdminPermission())
	g.PATCH("/settings", ctl.UpdateSettings, middlewares.AdminPermission())
	g.POST("/test", ctl.Test, middlewares.AdminPermission())

	g.GET("/packages", ctl.ListPackages)
	g.POST("/packages", ctl.CreatePackage, middlewares.AdminPermission())
	g.PATCH("/packages/:id", ctl.UpdatePackage, middlewares.AdminPermission())
	g.DELETE("/packages/:id", ctl.DeletePackage, middlewares.AdminPermission())

	g.GET("/requests", ctl.ListRequests, middlewares.AdminPermission())
	g.GET("/requests/:id", ctl.GetRequest, middlewares.AdminPermission())
	g.GET("/requests/:id/receipt", ctl.GetReceipt, middlewares.AdminPermission())
	g.POST("/requests/:id/approve", ctl.Approve, middlewares.AdminPermission())
	g.POST("/requests/:id/reject", ctl.Reject, middlewares.AdminPermission())
	g.POST("/requests/:id/confirm-payment", ctl.ConfirmPayment, middlewares.AdminPermission())
	g.DELETE("/requests/:id", ctl.DeleteRequest, middlewares.AdminPermission())

	g.GET("/accounts", ctl.AccountsForOcservUser)
	g.DELETE("/accounts/:id", ctl.DeleteAccount, middlewares.AdminPermission())
}
