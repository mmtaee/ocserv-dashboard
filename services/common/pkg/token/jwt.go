package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"time"
)

// Check parses and validates a JWT string using HMAC signing and the
// configured secret (cfg.JWTSecret). It also checks whether the token is expired.
// Returns the claims and a boolean indicating whether the token is valid and not expired.
func Check(tokenStr string) (jwt.MapClaims, bool) {
	conf := config.Get()
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(conf.JWTSecret), nil
	})
	if err != nil || token == nil {
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, false
	}

	// Check expiration (`exp` claim is in Unix time)
	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			// Token is expired
			return nil, false
		}
	}

	return claims, true
}
