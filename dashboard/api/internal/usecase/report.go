package usecase

import (
	"time"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
)

type ReportUseCase interface {
	SessionLogs(page, limit int, orderBy, sort string, startDate, endDate *time.Time) ([]models.OcservUserSessionLog, int64, error)
	Statistics(startDate, endDate *time.Time) ([]models.DailyTraffic, error)
	TotalBandwidthDateRange(startDate, endDate *time.Time) (repository.TotalBandwidths, error)
	UsersStat(adminID uint, role string) (repository.UserStatsResult, error)
}

type reportUseCase struct {
	reportRepo repository.ReportRepository
}

func NewReportUseCase(reportRepo repository.ReportRepository) ReportUseCase {
	return &reportUseCase{reportRepo: reportRepo}
}

func (uc *reportUseCase) SessionLogs(page, limit int, orderBy, sort string, startDate, endDate *time.Time) ([]models.OcservUserSessionLog, int64, error) {
	return uc.reportRepo.SessionLogs(page, limit, orderBy, sort, startDate, endDate)
}

func (uc *reportUseCase) Statistics(startDate, endDate *time.Time) ([]models.DailyTraffic, error) {
	return uc.reportRepo.Statistics(startDate, endDate)
}

func (uc *reportUseCase) TotalBandwidthDateRange(startDate, endDate *time.Time) (repository.TotalBandwidths, error) {
	return uc.reportRepo.TotalBandwidthDateRange(startDate, endDate)
}

func (uc *reportUseCase) UsersStat(adminID uint, role string) (repository.UserStatsResult, error) {
	return uc.reportRepo.UsersStat(adminID, role)
}
