package webhook

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/occtl"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/user"
)

type Handler struct {
	occtlHandler      occtl.OcservOcctlInterface
	ocservUserHandler user.OcservUserInterface
	dockerMode        bool
}

func NewHandler(dockerMode bool) *Handler {
	h := &Handler{
		dockerMode: dockerMode,
	}

	// For now, we only support non-docker mode. Docker mode would require occtl docker implementation.
	h.occtlHandler = occtl.NewOcservOcctl()
	h.ocservUserHandler = user.NewOcservUser()

	return h
}

func (h *Handler) HandleWebhook(c echo.Context) error {
	// Only accept POST method
	if c.Request().Method != http.MethodPost {
		return c.String(http.StatusMethodNotAllowed, "Method not allowed")
	}

	// Parse payload
	var payload WebhookPayload
	if err := c.Bind(&payload); err != nil {
		return c.String(http.StatusBadRequest, "Invalid payload: "+err.Error())
	}

	// Validate username
	if payload.Username == "" {
		return c.String(http.StatusBadRequest, "Username is required")
	}

	// Extract action from path (e.g., /webhook/disconnect -> "disconnect")
	path := c.Request().URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return c.String(http.StatusBadRequest, "Action not specified in URL path")
	}
	action := strings.ToLower(parts[1])

	logger.Info("Received webhook action: %s for username %s", action, payload.Username)

	// Handle action
	switch action {
	case "disconnect":
		msg, err := h.occtlHandler.DisconnectUser(payload.Username)
		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to disconnect user: "+err.Error())
		}
		return c.String(http.StatusOK, fmt.Sprintf("User %s disconnected successfully. Message: %s", payload.Username, msg))

	case "lock":
		msg, err := h.ocservUserHandler.Lock(payload.Username)
		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to lock user: "+err.Error())
		}
		return c.String(http.StatusOK, fmt.Sprintf("User %s locked successfully. Message: %s", payload.Username, msg))

	case "unlock":
		msg, err := h.ocservUserHandler.UnLock(payload.Username)
		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to unlock user: "+err.Error())
		}
		return c.String(http.StatusOK, fmt.Sprintf("User %s unlocked successfully. Message: %s", payload.Username, msg))

	default:
		return c.String(http.StatusBadRequest, "Unknown action: "+action)
	}
}
