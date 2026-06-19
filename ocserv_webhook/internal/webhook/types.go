package webhook

// WebhookPayload represents the incoming webhook payload
type WebhookPayload struct {
	Username string `json:"username" validate:"required"`
}
