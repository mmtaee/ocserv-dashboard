package superadmin

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
)

type Repository interface {
	EnsureSuperadmin(ctx context.Context, user *models.User) (*models.User, error)
}

type PasswordManager interface {
	CreatePassword(passwd string, saltLength ...int) crypto.CustomPassword
}
