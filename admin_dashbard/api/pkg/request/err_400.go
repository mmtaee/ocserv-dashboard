package request

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/errors"
)

func (r *Request) BadRequest(c echo.Context, err interface{}, code ...string) error {
	cd := "1000"
	if len(code) > 0 {
		cd = code[0]
	}
	return errors.RespondError(c, cd)
}
