package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/labstack/echo/v5"
)

// ErrorDetail represents the detailed error information from json config
type ErrorDetail struct {
	Status int    `json:"status"`
	FA     string `json:"fa"`
	EN     string `json:"en"`
	IT     string `json:"it"`
	RU     string `json:"ru"`
	ZhCn   string `json:"zh-cn"`
	ZhTw   string `json:"zh-tw"`
}

// ErrorResponse represents the structure of error responses
type ErrorResponse struct {
	Code    int      `json:"code" validate:"required"`
	Message string   `json:"message" validate:"required"`
	Error   []string `json:"error,omitempty"`
}

// AppError is a custom error type that includes an error code
type AppError struct {
	Code int
	Err  error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("error code %d", e.Code)
}

func NewAppError(code int, err error) error {
	return &AppError{
		Code: code,
		Err:  err,
	}
}

// Request provides methods for handling requests and responses
type Request struct{}

var (
	errorMap map[string]ErrorDetail
	once     sync.Once
)

// ResponseWithCode handles error responses using predefined error codes from config/errors.json
func (r *Request) ResponseWithCode(c *echo.Context, errorCode int, errs ...interface{}) error {
	loadErrorConfig()

	// If the first error is an AppError, use its code
	if len(errs) > 0 {
		if appErr, ok := errs[0].(*AppError); ok {
			errorCode = appErr.Code
		}
	}

	codeStr := fmt.Sprintf("%d", errorCode)
	detail, ok := errorMap[codeStr]
	if !ok {
		// Default to 500 if code not found
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    500,
			Message: "Internal Server Error",
		})
	}

	response := ErrorResponse{
		Code:    errorCode,
		Message: detail.FA,
	}

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
