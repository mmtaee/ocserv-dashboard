package captcha

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type GoogleCaptcha struct {
	SecretKey  string
	Client     *http.Client
	IsVerified bool
}

type GoogleCaptchaInterface interface {
	SetSecretKey(secretKey string)
	Verify(token string)
	IsValid() bool
}

func NewGoogleVerifier() *GoogleCaptcha {
	return &GoogleCaptcha{}
}

func (g *GoogleCaptcha) SetSecretKey(secretKey string) {
	g.SecretKey = secretKey
}

func (g *GoogleCaptcha) Verify(token string) {
	type captchaResponse struct {
		Success     bool      `json:"success"`
		ChallengeTS time.Time `json:"challenge_ts"`
		Hostname    string    `json:"hostname"`
		ErrorCodes  []string  `json:"error-codes"`
	}

	formData := url.Values{}
	formData.Set("secret", g.SecretKey)
	formData.Set("response", token)

	if g.Client == nil {
		g.Client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := g.Client.PostForm("https://www.google.com/recaptcha/api/siteverify", formData)
	if err != nil {
		return
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var response captchaResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}
	g.IsVerified = response.Success
	return
}

func (g *GoogleCaptcha) IsValid() bool {
	return g.IsVerified
}
