package middlewares

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	"net/http"
	"time"
)

func RequestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()

			err := next(c)

			req := c.Request()
			res := c.Response()
			status := http.StatusOK
			if echoResponse, unwrapErr := echo.UnwrapResponse(res); unwrapErr == nil {
				status = echoResponse.Status
			}

			logger.Info(
				"%s %s | %s | %d %s | %.3fs",
				req.Method,
				req.URL.Path,
				c.RealIP(),
				status,
				http.StatusText(status),
				time.Since(start).Seconds(),
			)

			return err
		}
	}
}
