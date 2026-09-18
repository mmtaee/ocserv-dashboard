package middlewares

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

func TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if isLongRunningBackupRestore(c) {
				return next(c)
			}

			ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
			defer cancel()

			c.SetRequest(c.Request().WithContext(ctx))

			done := make(chan error, 1)
			go func() {
				done <- next(c)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return echo.NewHTTPError(http.StatusRequestTimeout, "Request Timeout")
			}
		}
	}
}

func isLongRunningBackupRestore(c echo.Context) bool {
	if c.Request().Method != http.MethodPost {
		return false
	}
	path := c.Request().URL.Path
	return strings.HasPrefix(path, "/api/backup/ocserv_users") ||
		strings.HasPrefix(path, "/api/backup/ocserv_groups")
}
