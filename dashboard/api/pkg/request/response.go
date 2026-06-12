package request

// MessageResponse is the common JSON shape for endpoints that only need to
// return a human-readable status message.
type MessageResponse struct {
	Message string `json:"message"`
}
