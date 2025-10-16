package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/common/pkg/token"
	"strings"
)

func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return UnauthorizedError(c, "missing or invalid Authorization header")
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims, ok := token.Check(tokenStr)
			if !ok {
				return UnauthorizedError(c, "invalid token")
			}

			c.Set("userUID", claims["sub"])
			c.Set("isAdmin", claims["isAdmin"])
			c.Set("username", claims["username"])
			return next(c)
		}
	}
}
