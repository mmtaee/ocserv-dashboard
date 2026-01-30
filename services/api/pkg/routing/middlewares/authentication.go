package middlewares

import (
	"github.com/labstack/echo/v4"
	apiModels "github.com/mmtaee/ocserv-users-management/api/internal/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
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
				logger.Error("invalid token with claims", claims)
				return UnauthorizedError(c, "invalid token")
			}

			subIDFloat, ok := claims["sub-id"].(float64)
			if !ok {
				logger.Error("invalid token with sub-id", claims)
				return UnauthorizedError(c, "invalid token")
			}
			subID := uint(subIDFloat)

			c.Set("ID", subID)
			c.Set("userUID", claims["sub"])
			c.Set("role", apiModels.UserRole(claims["role"].(string)))
			c.Set("username", claims["username"])
			return next(c)
		}
	}
}
