package report

import (
	"context"
	"fmt"
	"slices"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
	"golang.org/x/sync/errgroup"
)

type Usecase struct {
	repository Repository
	sessions   SessionProvider
}

func New(repository Repository, sessions SessionProvider) *Usecase {
	return &Usecase{repository: repository, sessions: sessions}
}

func (u *Usecase) SessionLogs(ctx context.Context, pagination *request.Pagination, input SessionLogsData) (*[]models.OcservUserSessionLog, int64, error) {
	start, end, err := parseDateRange(input.DateStart, input.DateEnd, false, false)
	if err != nil {
		return nil, 0, err
	}
	return u.repository.SessionLogs(ctx, pagination, start, end)
}

func (u *Usecase) Statistics(ctx context.Context, input StatisticsData) (*[]models.DailyTraffic, error) {
	start, end, err := parseDateRange(input.DateStart, input.DateEnd, true, true)
	if err != nil {
		return nil, err
	}
	return u.repository.Statistics(ctx, start, end)
}

func (u *Usecase) TotalBandwidthDateRange(ctx context.Context, input TotalBandwidthData) (repository.TotalBandwidths, error) {
	start, end, err := parseDateRange(input.DateStart, input.DateEnd, false, true)
	if err != nil {
		return repository.TotalBandwidths{}, err
	}
	return u.repository.TotalBandwidthDateRange(ctx, start, end)
}

func (u *Usecase) TotalBandWidthUser(ctx context.Context, uid string) (repository.TotalBandwidths, error) {
	return u.repository.TotalBandWidthUser(ctx, uid)
}

func (u *Usecase) TenDaysStats(ctx context.Context) ([]models.DailyTraffic, error) {
	return u.repository.TenDaysStats(ctx)
}

func (u *Usecase) TotalUsers(ctx context.Context) (int64, error) {
	return u.repository.TotalUsers(ctx)
}

func (u *Usecase) TopBandwidthUser(ctx context.Context) (repository.TopBandwidthUsers, error) {
	return u.repository.TopBandwidthUser(ctx)
}

func (u *Usecase) TotalBandwidth(ctx context.Context) (repository.TotalBandwidths, error) {
	return u.repository.TotalBandwidth(ctx)
}

func (u *Usecase) UserReport(ctx context.Context) (*UserReport, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	var onlineUsers []string
	var report UserReport

	group.Go(func() error {
		sessions, err := u.sessions.OnlineSessions()
		if err != nil {
			return fmt.Errorf("failed to get online users: %w", err)
		}
		for _, session := range sessions {
			if !slices.Contains(onlineUsers, session.Username) {
				onlineUsers = append(onlineUsers, session.Username)
			}
		}
		return nil
	})
	group.Go(func() error {
		stats, err := u.repository.UsersStat(groupCtx)
		if err != nil {
			return fmt.Errorf("failed to get users stats: %w", err)
		}
		report.Active = stats.Active
		report.Deactivated = stats.Deactivated
		report.Locked = stats.Locked
		return nil
	})
	if err := group.Wait(); err != nil {
		return nil, err
	}
	report.Online = len(onlineUsers)
	return &report, nil
}
