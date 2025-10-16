package request

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"net/http"
)

type ErrorResponse struct {
	Error   []string `json:"error" validate:"required"`
	Message []string `json:"message" validate:"required"`
}

func (r *Request) BadRequest(c echo.Context, err interface{}, msg ...string) error {
	var response ErrorResponse

	switch err.(type) {
	case error:
		var pqErr *pgconn.PgError
		if errors.As(err.(error), &pqErr) {
			response.Error = append(response.Error, pqErr.Code)
		} else {
			response.Error = append(response.Error, err.(error).Error())
		}
	case string:
		response.Error = append(response.Error, err.(string))
	case map[string]interface{}:
		errs := err.(map[string]interface{})["error"]
		if errSlice, ok := errs.([]string); ok && len(errSlice) > 0 {
			response.Error = append(response.Error, errSlice...)
		}
	default:
		response.Error = append(response.Error, err.(string))
	}
	response.Message = append(response.Message, msg...)
	return c.JSON(http.StatusBadRequest, response)
}
