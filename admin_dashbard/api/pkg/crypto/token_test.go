package crypto

import (
	"testing"
	"time"

	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAccessToken(t *testing.T) {
	userID := "12345"
	adminUsername := "admin"
	secret := "my-secret-key"
	t.Setenv("JWT_SECRET", secret)
	config.Init(false, "", 0)

	expire := time.Now().Add(time.Hour).Unix()

	tokenString, err := GenerateAccessToken(userID, adminUsername, expire, true)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims["sub"])
	assert.Equal(t, true, claims["isAdmin"])
}
