package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"net/http"
	"time"
)

func RequestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			req := c.Request()
			res := c.Response()

			logger.Info(
				"%s %s | %s | %d %s | %.3fs",
				req.Method,
				req.URL.Path,
				c.RealIP(),
				res.Status,
				http.StatusText(res.Status),
				time.Since(start).Seconds(),
			)

			return err
		}
	}
}
