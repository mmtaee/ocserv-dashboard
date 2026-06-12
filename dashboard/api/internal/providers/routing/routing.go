package routing

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/service/auth"
)

func Register(e *echo.Echo) {
	auth.RegisterRoutes(e)
}
