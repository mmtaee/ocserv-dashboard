package crypto

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

const secureTokenBytes = 32

func GenerateSecureToken() (string, error) {
	random := make([]byte, secureTokenBytes)
	if _, err := cryptorand.Read(random); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(random), nil
}

func HashToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}
