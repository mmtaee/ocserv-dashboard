package middlewares

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/token"
	"strconv"
	"strings"
)

func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return UnauthorizedError(c, "missing or invalid Authorization header")
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims, ok := token.Check(tokenStr)
			if !ok {
				return UnauthorizedError(c, "invalid token")
			}

			subject, ok := claims["sub"].(string)
			if !ok {
				return UnauthorizedError(c, "invalid user id")
			}
			userID, err := strconv.ParseUint(subject, 10, 64)
			if err != nil || userID == 0 {
				return UnauthorizedError(c, "invalid user id")
			}

			c.Set("userID", uint(userID))
			c.Set("isAdmin", claims["isAdmin"])
			c.Set("username", claims["username"])
			return next(c)
		}
	}
}
