package request

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// Validator is a wrapper around go-playground/validator
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates a new instance of Validator
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validator for Iranian National Code
	_ = v.RegisterValidation("national_code", func(fl validator.FieldLevel) bool {
		code := fl.Field().String()
		if matched, _ := regexp.MatchString(`^\d{10}$`, code); !matched {
			return false
		}

		check, _ := strconv.Atoi(string(code[9]))
		sum := 0
		for i := 0; i < 9; i++ {
			num, _ := strconv.Atoi(string(code[i]))
			sum += num * (10 - i)
		}

		remainder := sum % 11
		if remainder < 2 {
			return check == remainder
		}
		return check == 11-remainder
	})

	return &Validator{
		validator: v,
	}
}

// Validate validates the given data struct and returns an error if validation fails
// It also binds the data from the request context if necessary
func (v *Validator) Validate(c *echo.Context, data interface{}) error {
	// Bind data from the request
	if err := c.Bind(data); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Validate the struct
	if err := v.validator.Struct(data); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	return nil
}
