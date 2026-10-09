package auth

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

type Repository interface {
	OcservUserByUsername(ctx context.Context, username string) (*models.OcservUser, error)
}
