package auth

import (
	"context"
	"errors"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/repository"
)

var (
	ErrUserNotFound = errors.New("ocserv user not found")
	ErrInvalidCreds = errors.New("invalid credentials")
)

type Verifier struct {
	repo *repository.Repository
}

func NewVerifier(repo *repository.Repository) *Verifier {
	return &Verifier{repo: repo}
}

// Verify validates the given username/password against the ocserv_users table.
// On success it returns the matching user; the caller is responsible for
// linking it to a Telegram chat.
//
// Locked (quota/expiry exhausted) and deactivated accounts are intentionally
// allowed to authenticate: in both cases the customer is exactly who needs to
// be able to link the account, inspect its status and request a renewal
// through the bot. The caller is expected to surface a hint about the
// degraded state via i18n (see handlers.completeLink).
func (v *Verifier) Verify(ctx context.Context, username, password string) (*models.OcservUser, error) {
	user, err := v.repo.OcservUserByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if user.Password != password {
		return nil, ErrInvalidCreds
	}
	return user, nil
}
