package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
)

func RequestLoggerMiddleware(service string) echo.MiddlewareFunc {
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
			if err != nil && status < http.StatusBadRequest {
				var httpErr *echo.HTTPError
				if errors.As(err, &httpErr) {
					status = httpErr.Code
				} else {
					status = http.StatusInternalServerError
				}
			}
			path := c.Path()
			if path == "" {
				path = req.URL.Path
			}
			logger.Request(service, c.RealIP(), fmt.Sprintf("%s %s", req.Method, path), status, time.Since(start))

			return err
		}
	}
}
