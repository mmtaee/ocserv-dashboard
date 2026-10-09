package superadmin

import (
	"context"
	"errors"
	"strings"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

type Usecase struct {
	repository Repository
	passwords  PasswordManager
}

func New(repository Repository, passwords PasswordManager) *Usecase {
	return &Usecase{repository: repository, passwords: passwords}
}

func (u *Usecase) Ensure(ctx context.Context, username, password string) (*models.User, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if len(username) < 2 || len(username) > 16 {
		return nil, errors.New("SUPERADMIN_USERNAME must contain 2 to 16 characters")
	}
	if len(password) < 4 || len(password) > 64 {
		return nil, errors.New("SUPERADMIN_PASSWORD must contain 4 to 64 characters")
	}
	hashed := u.passwords.CreatePassword(password)
	return u.repository.EnsureSuperadmin(ctx, &models.User{
		Username: username, Password: hashed.Hash, Salt: hashed.Salt, Superadmin: true,
	})
}
