package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// GenerateRandomToken creates a secure random token
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateAdministratorToken creates a new random auth token for an Administrator
func CreateAdministratorToken() (string, error) {
	token, err := GenerateRandomToken(32) // 64 hex characters
	if err != nil {
		return "", err
	}

	return token, nil
}
