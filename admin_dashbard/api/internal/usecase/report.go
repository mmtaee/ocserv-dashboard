package usecase

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"slices"
	"strings"
)

type ReportUsecaseInterface interface {
	SessionLogs(ctx context.Context, pagination *request.Pagination, dateStart, dateEnd *time.Time) ([]models.OcservUserSessionLog, int64, error)
	Statistics(ctx context.Context, dateStart, dateEnd *time.Time) ([]models.DailyTraffic, error)
	TotalBandwidth(ctx context.Context, dateStart, dateEnd *time.Time) (repository.TotalBandwidths, error)
	UsersReport(ctx context.Context) (int, repository.UserStatsResult, error)
}

type ReportUsecase struct {
	reportRepo      repository.ReportRepositoryInterface
	ocservOcctlRepo repository.OcctlRepositoryInterface
}

func NewReportUsecase(
	reportRepo repository.ReportRepositoryInterface,
	ocservOcctlRepo repository.OcctlRepositoryInterface,
) *ReportUsecase {
	return &ReportUsecase{
		reportRepo:      reportRepo,
		ocservOcctlRepo: ocservOcctlRepo,
	}
}

func (uc *ReportUsecase) SessionLogs(
	ctx context.Context,
	pagination *request.Pagination,
	dateStart, dateEnd *time.Time,
) ([]models.OcservUserSessionLog, int64, error) {
	return uc.reportRepo.SessionLogs(ctx, pagination, dateStart, dateEnd)
}

func (uc *ReportUsecase) Statistics(
	ctx context.Context,
	dateStart, dateEnd *time.Time,
) ([]models.DailyTraffic, error) {
	return uc.reportRepo.Statistics(ctx, dateStart, dateEnd)
}

func (uc *ReportUsecase) TotalBandwidth(
	ctx context.Context,
	dateStart, dateEnd *time.Time,
) (repository.TotalBandwidths, error) {
	return uc.reportRepo.TotalBandwidthDateRange(ctx, dateStart, dateEnd)
}

func (uc *ReportUsecase) UsersReport(ctx context.Context) (int, repository.UserStatsResult, error) {
	var wg sync.WaitGroup
	var onlineUsers []string
	var result repository.UserStatsResult

	errChan := make(chan error, 2)
	wg.Add(2)

	go func() {
		defer wg.Done()

		users, err := uc.ocservOcctlRepo.OnlineSessions()
		if err != nil {
			logger.Error("failed to get online users: %v", err)
			errChan <- errors.New("failed to get online users")
			return
		}
		onlineUsernames := make([]string, 0)

		for _, u := range users {
			if !slices.Contains(onlineUsernames, u.Username) {
				onlineUsernames = append(onlineUsernames, u.Username)
			}
		}
		onlineUsers = onlineUsernames
	}()

	go func() {
		defer wg.Done()

		res, err := uc.reportRepo.UsersStat(ctx)
		if err != nil {
			logger.Error("failed to get users stats: %v", err)
			errChan <- errors.New("failed to get users stats")
			return
		}
		result = res
	}()

	wg.Wait()
	close(errChan)

	var errs []string
	for e := range errChan {
		errs = append(errs, e.Error())
	}

	if len(errs) > 0 {
		return 0, repository.UserStatsResult{}, errors.New(strings.Join(errs, "; "))
	}

	return len(onlineUsers), result, nil
}
