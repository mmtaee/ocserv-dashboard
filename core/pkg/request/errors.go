package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/labstack/echo/v5"
)

// Request provides methods for handling requests and responses
type Request struct{}

var (
	errorMap map[string]ErrorDetail
	once     sync.Once
)

// ResponseWithCode handles error responses using predefined error codes from config/errors.json
func (r *Request) ResponseWithCode(c *echo.Context, errorCode interface{}, errs ...interface{}) error {
	loadErrorConfig()

	// Handle both int and string error codes
	var codeStr string
	switch v := errorCode.(type) {
	case int:
		codeStr = fmt.Sprintf("%d", v)
	case string:
		codeStr = v
	}

	// If the first error is an AppError, use its code
	if len(errs) > 0 {
		if appErr, ok := errs[0].(*AppError); ok {
			codeStr = fmt.Sprintf("%d", appErr.Code)
		}
	}

	detail, ok := errorMap[codeStr]
	if !ok {
		// Default to 500 if code not found
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    500,
			Message: "Internal Server Error",
		})
	}

	response := ErrorResponse{
		Code:    0, // Will set later
		Message: detail.FA,
	}
	fmt.Sscanf(codeStr, "%d", &response.Code)

	for _, err := range errs {
		switch v := err.(type) {
		case error:
			response.Error = append(response.Error, v.Error())
		case string:
			response.Error = append(response.Error, v)
		case []string:
			response.Error = append(response.Error, v...)
		case nil:
			// Do nothing
		default:
			response.Error = append(response.Error, fmt.Sprintf("%v", v))
		}
	}

	return c.JSON(detail.Status, response)
}

// BadRequest handles 400 Bad Request responses with a detailed error format
func (r *Request) BadRequest(c *echo.Context, err interface{}, msg ...string) error {
	return r.ResponseWithCode(c, 1004, err)
}

// InternalServerError handles 500 Internal Server Error responses
func (r *Request) InternalServerError(c *echo.Context, err interface{}, msg ...string) error {
	return r.ResponseWithCode(c, 5001, err)
}

// Unauthorized handles 401 Unauthorized responses
func (r *Request) Unauthorized(c *echo.Context, err interface{}, msg ...string) error {
	return r.ResponseWithCode(c, 2001, err)
}

// Forbidden handles 403 Forbidden responses
func (r *Request) Forbidden(c *echo.Context, err interface{}, msg ...string) error {
	return r.ResponseWithCode(c, 2002, err)
}

func loadErrorConfig() {
	once.Do(func() {
		errorMap = make(map[string]ErrorDetail)
		// Assuming the file is in config/errors.json relative to project root
		// In a real scenario, this path might need to be more robust
		data, err := os.ReadFile("config/errors.json")
		if err != nil {
			fmt.Printf("Warning: could not load error config: %v\n", err)
			return
		}
		if err := json.Unmarshal(data, &errorMap); err != nil {
			fmt.Printf("Warning: could not unmarshal error config: %v\n", err)
		}
	})
}
