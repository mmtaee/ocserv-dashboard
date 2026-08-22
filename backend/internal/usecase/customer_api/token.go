package customer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (u *Usecase) createToken(username string, expiresAt time.Time) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" || strings.Contains(username, "|") {
		return "", errors.New("invalid username")
	}
	payload := username + "|" + strconv.FormatInt(expiresAt.Unix(), 10)
	signature, err := u.sign(payload)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString([]byte(payload + "|" + signature)), nil
}

func (u *Usecase) parseToken(token string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil {
		return "", errors.New("invalid token")
	}
	parts := strings.Split(string(raw), "|")
	if len(parts) != 3 {
		return "", errors.New("invalid token")
	}
	expires, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || u.now().After(time.Unix(expires, 0)) {
		return "", errors.New("token has expired")
	}
	expected, err := u.sign(parts[0] + "|" + parts[1])
	if err != nil {
		return "", err
	}
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return "", errors.New("invalid token signature")
	}
	return parts[0], nil
}

func (u *Usecase) sign(payload string) (string, error) {
	if strings.TrimSpace(u.secretKey) == "" {
		return "", errors.New("secret key is not configured")
	}
	mac := hmac.New(sha256.New, []byte(u.secretKey))
	if _, err := mac.Write([]byte(payload)); err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
