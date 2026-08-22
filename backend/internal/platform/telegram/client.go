package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const apiBase = "https://api.telegram.org"

type Client struct {
	http *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{http: httpClient}
}

func (c *Client) Username(ctx context.Context, token string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/bot%s/getMe", apiBase, token), nil)
	if err != nil {
		return "", err
	}
	response, err := c.http.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("telegram getMe returned status %d", response.StatusCode)
	}
	var envelope struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		return "", err
	}
	if !envelope.OK {
		return "", fmt.Errorf("telegram getMe unsuccessful")
	}
	return envelope.Result.Username, nil
}

func (c *Client) Send(ctx context.Context, token string, chatID int64, text string, html bool) (int64, error) {
	form := url.Values{"chat_id": {strconv.FormatInt(chatID, 10)}, "text": {text}, "disable_web_page_preview": {"true"}}
	if html {
		form.Set("parse_mode", "HTML")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/bot%s/sendMessage", apiBase, token), strings.NewReader(form.Encode()))
	if err != nil {
		return 0, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := c.http.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("telegram sendMessage status=%d body=%s", response.StatusCode, string(body))
	}
	var envelope struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil || !envelope.OK {
		return 0, fmt.Errorf("telegram sendMessage: invalid or unsuccessful response: %s", string(body))
	}
	return envelope.Result.MessageID, nil
}

func (c *Client) Delete(ctx context.Context, token string, chatID, messageID int64) error {
	form := url.Values{"chat_id": {strconv.FormatInt(chatID, 10)}, "message_id": {strconv.FormatInt(messageID, 10)}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/bot%s/deleteMessage", apiBase, token), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram deleteMessage returned status %d", response.StatusCode)
	}
	return nil
}
