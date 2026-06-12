package middlewares

import (
	"context"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/request"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

// TimeoutMiddleware creates a middleware that cancels the context after a timeout
// Usage: e.Use(middlewares.TimeoutMiddleware(30 * time.Second))
func TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
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
				return c.JSON(http.StatusRequestTimeout, request.MessageResponse{Message: "request timeout"})
			}
		}
	}
}
