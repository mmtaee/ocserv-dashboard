package units

import (
	"context"
	"errors"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

type staticAuthenticator struct{}

func (staticAuthenticator) FindActiveByToken(_ context.Context, token string) (*models.UserToken, error) {
	superadmin := token == "superadmin"
	if !superadmin && token != "normal" {
		return nil, errors.New("invalid token")
	}
	return &models.UserToken{
		ID: 10, User: models.User{ID: 7, Username: "staff", Superadmin: superadmin},
	}, nil
}
