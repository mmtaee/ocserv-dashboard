package webhook

import "github.com/labstack/echo/v5"

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.Any("/webhook/*", h.HandleWebhook)
}
