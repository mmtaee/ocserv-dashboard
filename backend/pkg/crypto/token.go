package crypto

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	"github.com/oklog/ulid/v2"
)

func GenerateAccessToken(userID uint, username string, expire int64, superadmin bool) (string, error) {
	cfg := config.Get()

	claims := jwt.MapClaims{
		"sub":        strconv.FormatUint(uint64(userID), 10),
		"jti":        ulid.Make().String(),
		"exp":        expire,
		"iat":        time.Now().Unix(),
		"user_id":    userID,
		"superadmin": superadmin,
		"username":   username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}
