package occtl_docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"net/http"
	"time"
)

type WebhookPayload struct {
	Username string `json:"username"`
}

type OcservOcctlDocker struct {
	apiURL string
}

type OcservOcctlUsersDocker interface {
	DisconnectUser(username string) (string, error)
	Lock(username string) (string, error)
	Unlock(username string) (string, error)
}

func NewOcservOcctlDocker() *OcservOcctlDocker {
	return &OcservOcctlDocker{apiURL: "http://ocserv:8888"}
}

// call webhook endpoint api
func (d *OcservOcctlDocker) call(name string, username string) error {
	endpoint := fmt.Sprintf("%s/webhook/%s", d.apiURL, name)

	logger.Info("Docker webhook call for %s %s", name, username)

	payload := WebhookPayload{Username: username}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to call webhook endpoint: %v", err)
		return fmt.Errorf("call webhook %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Error("Failed to call webhook endpoint with status: %d", resp.StatusCode)
		return fmt.Errorf("webhook %s failed: status %d", name, resp.StatusCode)
	}

	return nil
}

func (d *OcservOcctlDocker) DisconnectUser(username string) (string, error) {
	return "", d.call("disconnect", username)
}

func (d *OcservOcctlDocker) Lock(username string) (string, error) {
	return "", d.call("lock", username)
}

func (d *OcservOcctlDocker) Unlock(username string) (string, error) {
	return "", d.call("unlock", username)
}
