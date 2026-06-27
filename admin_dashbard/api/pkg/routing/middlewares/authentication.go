package middlewares

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/token"
)

func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return UnauthorizedError(c, "1012")
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims, ok := token.Check(tokenStr)
			if !ok {
				return UnauthorizedError(c, "1013")
			}

			c.Set("userUID", claims["sub"])
			c.Set("isAdmin", claims["isAdmin"])
			c.Set("username", claims["username"])
			return next(c)
		}
	}
}
