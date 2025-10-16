package crypto

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"github.com/oklog/ulid/v2"
	"log"
	"time"
)

func GenerateAccessToken(userID, username string, expire int64, isAdmin bool) (string, error) {
	cfg := config.Get()

	log.Println("\n\n\n ", username)
	claims := jwt.MapClaims{
		"sub":      userID,
		"jti":      ulid.Make().String(),
		"exp":      expire,
		"iat":      time.Now().Unix(),
		"isAdmin":  isAdmin,
		"username": username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}
