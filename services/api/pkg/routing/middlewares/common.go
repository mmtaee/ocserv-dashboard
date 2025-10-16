package middlewares

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Unauthorized struct {
	Error string `json:"error"`
}

type PermissionDenied struct {
	Error string `json:"error"`
}

type TooManyRequests struct {
	Error string `json:"error"`
}

func UnauthorizedError(c echo.Context, msg string) error {
	return c.JSON(http.StatusUnauthorized, Unauthorized{Error: msg})
}

func PermissionDeniedError(c echo.Context, msg string) error {
	return c.JSON(http.StatusForbidden, PermissionDenied{Error: msg})
}

func TooManyRequestsError(c echo.Context, msg string) error {
	return c.JSON(http.StatusTooManyRequests, TooManyRequests{Error: msg})
}
