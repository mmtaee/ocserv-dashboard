package customer

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/authz"
)

const accessTokenTTL = 7 * time.Hour

var jwtHeader = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

var ErrInvalidToken = errors.New("invalid token")

type tokenClaims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"sub"`
	IssuedAt int64  `json:"iat"`
	Expires  int64  `json:"exp"`
}

func (u *Usecase) createToken(userID uint, username string, expiresAt time.Time) (string, error) {
	claims, err := json.Marshal(tokenClaims{UserID: userID, Username: username, IssuedAt: u.now().Unix(), Expires: expiresAt.Unix()})
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(claims)
	signature, err := u.sign(jwtHeader + "." + payload)
	if err != nil {
		return "", err
	}
	return jwtHeader + "." + payload + "." + signature, nil
}

func (u *Usecase) AuthenticateToken(ctx context.Context, token string) (authz.Principal, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 || parts[0] != jwtHeader {
		return authz.Principal{}, ErrInvalidToken
	}
	expected, err := u.sign(parts[0] + "." + parts[1])
	if err != nil || !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return authz.Principal{}, ErrInvalidToken
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return authz.Principal{}, ErrInvalidToken
	}
	var claims tokenClaims
	if json.Unmarshal(raw, &claims) != nil || claims.UserID == 0 || strings.TrimSpace(claims.Username) == "" || !u.now().Before(time.Unix(claims.Expires, 0)) {
		return authz.Principal{}, ErrInvalidToken
	}
	user, err := u.users.GetByUsername(ctx, claims.Username)
	if err != nil || user.ID != claims.UserID {
		return authz.Principal{}, ErrInvalidToken
	}
	return authz.Principal{UserID: claims.UserID, Username: claims.Username}, nil
}

func (u *Usecase) sign(payload string) (string, error) {
	if strings.TrimSpace(u.secretKey) == "" {
		return "", errors.New("secret key is not configured")
	}
	mac := hmac.New(sha256.New, []byte(u.secretKey))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
