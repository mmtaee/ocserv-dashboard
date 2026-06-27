package request

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Request struct {
	validator *validator.Validate
}

type CustomRequestInterface interface {
	DoValidate(echo.Context, interface{}) interface{}
	BadRequest(c echo.Context, err interface{}, msg ...string) error
	Pagination(c echo.Context) *Pagination
	//Response(c echo.Context, p *Pagination, total int64, result interface{}) error
}

func NewCustomRequest() *Request {
	return &Request{
		validator: validator.New(),
	}
}
