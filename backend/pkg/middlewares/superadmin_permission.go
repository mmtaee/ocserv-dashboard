package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func SuperadminPermission() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			principal, err := Principal(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}
			if err := principal.RequireSuperadmin(); err != nil {
				return echo.NewHTTPError(http.StatusForbidden, err.Error())
			}
			return next(c)
		}
	}
}
