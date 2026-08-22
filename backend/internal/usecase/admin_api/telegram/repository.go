package telegram

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
)

// Repository is the persistence boundary required by this usecase.
type Repository interface {
	repository.TelegramRepositoryInterface
}

type UserManager interface {
	Create(ctx context.Context, user *models.OcservUser) (*models.OcservUser, error)
	Update(ctx context.Context, user *models.OcservUser) (*models.OcservUser, error)
	GetByUID(ctx context.Context, uid string) (*models.OcservUser, error)
}

type Client interface {
	Username(ctx context.Context, token string) (string, error)
	Send(ctx context.Context, token string, chatID int64, text string, html bool) (int64, error)
	Delete(ctx context.Context, token string, chatID, messageID int64) error
}
