package middlewares

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/authz"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/token"
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

			userID, ok := claims["user_id"].(float64)
			if !ok || userID <= 0 || userID != float64(uint(userID)) {
				return UnauthorizedError(c, "invalid user id")
			}

			username, usernameOK := claims["username"].(string)
			superadmin, superadminOK := claims["superadmin"].(bool)
			if !usernameOK || username == "" || !superadminOK {
				return UnauthorizedError(c, "invalid token claims")
			}

			c.Set(principalContextKey, authz.Principal{UserID: uint(userID), Username: username, Superadmin: superadmin})
			return next(c)
		}
	}
}

const principalContextKey = "principal"

func Principal(c *echo.Context) (authz.Principal, error) {
	principal, ok := c.Get(principalContextKey).(authz.Principal)
	if !ok || principal.UserID == 0 || principal.Username == "" {
		return authz.Principal{}, authz.ErrForbidden
	}
	return principal, nil
}
