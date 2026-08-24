package middlewares

import (
	"context"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/authz"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

type TokenAuthenticator interface {
	FindActiveByToken(ctx context.Context, token string) (*models.UserToken, error)
}

// AuthMiddleware authenticates a database-backed bearer session.
// Usage: e.Use(middlewares.AuthMiddleware(repository.NewUserTokenRepository()))
func AuthMiddleware(authenticator TokenAuthenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return UnauthorizedError(c, "missing or invalid Authorization header")
			}

			tokenValue := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if tokenValue == "" {
				return UnauthorizedError(c, "invalid token")
			}
			session, err := authenticator.FindActiveByToken(c.Request().Context(), tokenValue)
			if err != nil || session.ID == 0 || session.User.ID == 0 || session.User.Username == "" {
				return UnauthorizedError(c, "invalid token")
			}

			c.Set(principalContextKey, authz.Principal{
				SessionID: session.ID, UserID: session.User.ID,
				Username: session.User.Username, Superadmin: session.User.Superadmin,
			})
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
