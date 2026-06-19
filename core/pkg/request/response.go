package request

import "fmt"

// MessageResponse is the common JSON shape for endpoints that only need to
// return a human-readable status message.
type MessageResponse struct {
	Message string `json:"message"`
}

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
