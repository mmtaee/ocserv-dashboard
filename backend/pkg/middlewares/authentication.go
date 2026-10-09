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
			tokenValue, err := BearerToken(c)
			if err != nil {
				return UnauthorizedError(c, "missing or invalid Authorization header")
			}
			session, err := authenticator.FindActiveByToken(c.Request().Context(), tokenValue)
			if err != nil || session.ID == 0 || session.User.ID == 0 || session.User.Username == "" {
				return UnauthorizedError(c, "invalid token")
			}

			SetPrincipal(c, authz.Principal{
				SessionID: session.ID, UserID: session.User.ID,
				Username: session.User.Username, Superadmin: session.User.Superadmin,
			})
			return next(c)
		}
	}
}

const principalContextKey = "principal"

func BearerToken(c *echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", authz.ErrForbidden
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" {
		return "", authz.ErrForbidden
	}
	return token, nil
}

func SetPrincipal(c *echo.Context, principal authz.Principal) {
	c.Set(principalContextKey, principal)
}

func Principal(c *echo.Context) (authz.Principal, error) {
	principal, ok := c.Get(principalContextKey).(authz.Principal)
	if !ok || principal.UserID == 0 || principal.Username == "" {
		return authz.Principal{}, authz.ErrForbidden
	}
	return principal, nil
}
