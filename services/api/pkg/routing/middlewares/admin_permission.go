package middlewares

import (
	"github.com/labstack/echo/v4"
)

func AdminPermission() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Get("isAdmin").(bool) {
				return next(c)
			}
			return PermissionDeniedError(c, "Admin permission required")
		}
	}
}
