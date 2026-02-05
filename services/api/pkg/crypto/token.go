package crypto

import (
	"github.com/golang-jwt/jwt/v5"
	apiModels "github.com/mmtaee/ocserv-users-management/api/internal/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"github.com/oklog/ulid/v2"
	"time"
)

func GenerateAccessToken(user *apiModels.User, expire int64) (string, error) {
	cfg := config.Get()

	claims := jwt.MapClaims{
		"sub":      user.UID,
		"sub-id":   user.ID,
		"jti":      ulid.Make().String(),
		"exp":      expire,
		"iat":      time.Now().Unix(),
		"role":     user.Role,
		"username": user.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}
