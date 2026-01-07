package crypto

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"github.com/oklog/ulid/v2"
	"time"
)

func GenerateAccessToken(userID, username string, expire int64, role string) (string, error) {
	cfg := config.Get()

	claims := jwt.MapClaims{
		"sub":      userID,
		"jti":      ulid.Make().String(),
		"exp":      expire,
		"iat":      time.Now().Unix(),
		"role":     role,
		"username": username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}
