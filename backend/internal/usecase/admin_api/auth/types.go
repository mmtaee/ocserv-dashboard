package auth

import (
	"context"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Repository interface {
	DeleteByID(ctx context.Context, id uint) error
	List(ctx context.Context, pagination *request.Pagination) ([]models.UserToken, int64, error)
}

type Session struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
}

type SessionsResponse struct {
	Meta   request.Meta `json:"meta"`
	Result []Session    `json:"result"`
}
