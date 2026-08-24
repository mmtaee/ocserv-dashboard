package ocservuser

import (
	"context"
	"errors"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/authz"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

// ResetUsage clears the live counters and starts a new usage accounting window.
func (u *Usecase) ResetUsage(ctx context.Context, principal authz.Principal, id uint) (*models.OcservUser, error) {
	if id == 0 {
		return nil, errors.New("user id is required")
	}

	account, err := u.authorizedUser(ctx, principal, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	account.RunningRx = 0
	account.RunningTx = 0
	account.UsageResetAt = &now

	return u.Repository.ResetUsage(ctx, account)
}
