package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/errors"
)

func UnauthorizedError(c echo.Context, code ...string) error {
	cd := "1001"
	if len(code) > 0 {
		cd = code[0]
	}
	return errors.RespondError(c, cd)
}

func PermissionDeniedError(c echo.Context, msg ...string) error {
	return errors.RespondError(c, "1002")
}

func TooManyRequestsError(c echo.Context, msg ...string) error {
	return errors.RespondError(c, "1019")
}
