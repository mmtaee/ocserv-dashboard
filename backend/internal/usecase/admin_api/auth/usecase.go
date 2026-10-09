package auth

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Usecase struct {
	repository Repository
}

func New(repository Repository) *Usecase {
	return &Usecase{repository: repository}
}

func (u *Usecase) Logout(ctx context.Context, sessionID uint) error {
	return u.repository.DeleteByID(ctx, sessionID)
}

func (u *Usecase) Revoke(ctx context.Context, sessionID uint) error {
	return u.repository.DeleteByID(ctx, sessionID)
}

func (u *Usecase) Sessions(ctx context.Context, pagination *request.Pagination) ([]Session, int64, error) {
	stored, total, err := u.repository.List(ctx, pagination)
	if err != nil {
		return nil, 0, err
	}
	result := make([]Session, 0, len(stored))
	for _, session := range stored {
		result = append(result, Session{
			ID: session.ID, UserID: session.UserID, Username: session.User.Username,
			UserAgent: session.UserAgent, CreatedAt: session.CreatedAt, ExpireAt: session.ExpireAt,
		})
	}
	return result, total, nil
}
