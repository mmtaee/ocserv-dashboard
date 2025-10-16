package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"unicode"
)

func (r *Request) DoValidate(c echo.Context, data interface{}) interface{} {
	if err := c.Bind(&data); err != nil {
		return errorWrapper(err)
	}

	if err := r.validator.Struct(data); err != nil {
		return errorWrapper(err)
	}
	return nil
}

func formatSnakeCase(s string) string {
	var result []rune
	for i, char := range s {
		if unicode.IsUpper(char) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(char))
	}
	return string(result)
}

func formatError(err validator.FieldError) string {
	field := formatSnakeCase(err.Field())
	switch err.Tag() {
	case "required":
		return field + " is required"
	case "min":
		return field + " must be at least " + err.Param() + " characters long"
	case "max":
		return field + " must be at most " + err.Param() + " characters long"
	default:
		return "Invalid input"
	}
}

func errorWrapper(err error) interface{} {
	var (
		invalidValidationError *validator.InvalidValidationError
		httpErr                *echo.HTTPError
		validationErrors       []string
	)

	if errors.As(err, &invalidValidationError) {
		return map[string]interface{}{
			"error": "validation error",
		}
	}

	if errors.As(err, &httpErr) {
		var internalErr *json.UnmarshalTypeError
		if errors.As(err, &internalErr) {
			return errors.New(fmt.Sprintf(
				"field %v expected %v got %v",
				internalErr.Field,
				internalErr.Type,
				internalErr.Value,
			))
		}
		return errors.New("json parse error")
	}

	for _, err := range err.(validator.ValidationErrors) {
		validationErrors = append(validationErrors, formatError(err))
	}
	return map[string]interface{}{
		"error": validationErrors,
	}
}
