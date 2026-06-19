package routing

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/internal/service/auth"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/internal/service/customer"
)

func Register(e *echo.Echo) {
	auth.RegisterRoutes(e)
	customer.RegisterRoutes(e)
}
