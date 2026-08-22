package logprocessor

import (
	"context"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

type Repository interface {
	FindUser(ctx context.Context, username string) (*models.OcservUser, error)
	CreateTraffic(ctx context.Context, traffic *models.OcservUserTrafficStatistics) error
	CurrentMonthTotals(ctx context.Context, userID uint, usageResetAt *time.Time) (totalRX, totalTX int, err error)
	SaveUser(ctx context.Context, user *models.OcservUser) error
	SaveSessionLog(ctx context.Context, log *models.OcservUserSessionLog) error
}

type AccessController interface {
	Disconnect(username string) error
	Lock(username string) error
}

type UserStats struct {
	Username  string
	IP        string
	SessionID string
	RX        int
	TX        int
}
