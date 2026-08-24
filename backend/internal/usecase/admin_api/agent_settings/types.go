package agent_settings

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

type Repository interface {
	GetOrCreate(ctx context.Context, token string) (*models.AgentToken, error)
	Replace(ctx context.Context, token string) (*models.AgentToken, error)
	Delete(ctx context.Context) error
}

type TokenGenerator func() (string, error)
