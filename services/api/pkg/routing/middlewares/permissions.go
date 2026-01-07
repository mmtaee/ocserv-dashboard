package middlewares

import (
	"github.com/labstack/echo/v4"
)

func AdminPermission() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role := c.Get("role").(string)
			if role == "super-admin" || role == "admin" {
				return next(c)
			}
			return PermissionDeniedError(c, "Super Admin or Admin permission required")
		}
	}
}

func SuperAdminPermission() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role := c.Get("role").(string)
			if role == "super-admin" {
				return next(c)
			}
			return PermissionDeniedError(c, "Super Admin permission required")
		}
	}
}
