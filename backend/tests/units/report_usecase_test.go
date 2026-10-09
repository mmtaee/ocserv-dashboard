package units

import (
	"context"
	"testing"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	reportusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/reports"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
	"github.com/stretchr/testify/require"
)

type reportRepository struct {
	start *time.Time
	end   *time.Time
}

func (r *reportRepository) SessionLogs(_ context.Context, _ *request.Pagination, start, end *time.Time) (*[]models.OcservUserSessionLog, int64, error) {
	r.start, r.end = start, end
	logs := []models.OcservUserSessionLog{}
	return &logs, 0, nil
}

func (r *reportRepository) Statistics(context.Context, *time.Time, *time.Time) (*[]models.DailyTraffic, error) {
	result := []models.DailyTraffic{}
	return &result, nil
}

func (r *reportRepository) TotalBandwidthDateRange(context.Context, *time.Time, *time.Time) (repository.TotalBandwidths, error) {
	return repository.TotalBandwidths{}, nil
}

func (r *reportRepository) UsersStat(context.Context) (repository.UserStatsResult, error) {
	return repository.UserStatsResult{}, nil
}

func (r *reportRepository) TotalBandWidthUser(context.Context, uint) (repository.TotalBandwidths, error) {
	return repository.TotalBandwidths{}, nil
}

func (r *reportRepository) TenDaysStats(context.Context) ([]models.DailyTraffic, error) {
	return nil, nil
}

func (r *reportRepository) TotalUsers(context.Context) (int64, error) { return 0, nil }

func (r *reportRepository) TopBandwidthUser(context.Context) (repository.TopBandwidthUsers, error) {
	return repository.TopBandwidthUsers{}, nil
}

func (r *reportRepository) TotalBandwidth(context.Context) (repository.TotalBandwidths, error) {
	return repository.TotalBandwidths{}, nil
}

type reportSessions struct{}

func (reportSessions) OnlineSessions() ([]models.OnlineUserSession, error) { return nil, nil }

func TestReportSessionLogsNormalizesDateRangeInUsecase(t *testing.T) {
	repo := &reportRepository{}
	usecase := reportusecase.New(repo, reportSessions{})

	_, _, err := usecase.SessionLogs(
		context.Background(),
		&request.Pagination{Page: 1, PageSize: 10},
		reportusecase.SessionLogsData{DateStart: "2026-08-01", DateEnd: "2026-08-02"},
	)

	require.NoError(t, err)
	require.Equal(t, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), *repo.start)
	require.Equal(t, time.Date(2026, 8, 2, 23, 59, 59, 0, time.UTC), *repo.end)
}

func TestReportStatisticsRejectsInvalidRangeInUsecase(t *testing.T) {
	usecase := reportusecase.New(&reportRepository{}, reportSessions{})

	_, err := usecase.Statistics(context.Background(), reportusecase.StatisticsData{
		DateStart: "2026-08-03",
		DateEnd:   "2026-08-02",
	})

	require.EqualError(t, err, "date start is after end")
}
