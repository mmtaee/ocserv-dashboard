package report

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
)

type Controller struct {
	request         request.CustomRequestInterface
	reportRepo      repository.ReportRepositoryInterface
	ocservOcctlRepo repository.OcctlRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request:         request.NewCustomRequest(),
		reportRepo:      repository.NewtReportRepository(),
		ocservOcctlRepo: repository.NewOcctlRepository(),
	}
}

// SessionLogs 	 Ocserv session logs
//
// @Summary      Ocserv session logs
// @Description  Ocserv session logs
// @Tags         Report
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 page query int false "Page number, starting from 1" minimum(1)
// @Param 		 size query int false "Number of items per page" minimum(1) maximum(100) name(size)
// @Param 		 order query string false "Field to order by"
// @Param 		 sort query string false "Sort order, either ASC or DESC" Enums(ASC, DESC)
// @Param 		 date_start query string false "date_start"
// @Param 		 date_end query string false "date_end"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} SessionLogsResponse
// @Router       /reports/session_logs [get]
func (ctl *Controller) SessionLogs(c echo.Context) error {
	var data SessionLogsData
	if err := c.Bind(&data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	pagination := ctl.request.Pagination(c)

	var startDate, endDate *time.Time

	if data.DateStart != "" {
		t, err := time.Parse("2006-01-02", data.DateStart)
		if err != nil {
			return ctl.request.BadRequest(c, fmt.Errorf("invalid date_start: %w", err))
		}
		startDate = &t
	}

	if data.DateEnd != "" {
		t, err := time.Parse("2006-01-02", data.DateEnd)
		if err != nil {
			return ctl.request.BadRequest(c, fmt.Errorf("invalid date_end: %w", err))
		}
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		endDate = &t
	}

	logs, total, err := ctl.reportRepo.SessionLogs(c.Request().Context(), pagination, startDate, endDate)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, SessionLogsResponse{
		Meta: request.Meta{
			Page:         pagination.Page,
			TotalRecords: total,
			PageSize:     pagination.PageSize,
		},
		Result: logs,
	})
}

// Statistics 	 Ocserv Users Statistics
//
// @Summary      Ocserv Users Statistics
// @Description  Ocserv Users Statistics
// @Tags         Report
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 date_start query string true "date_start"
// @Param 		 date_end query string true "date_end"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200 {object} []models.DailyTraffic
// @Router       /reports/statistics [get]
func (ctl *Controller) Statistics(c echo.Context) error {
	var data StatisticsData
	if err := c.Bind(&data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	if data.DateStart == "" || data.DateEnd == "" {
		return ctl.request.BadRequest(c, errors.New("statistics date start and end are required"))
	}

	var startDate, endDate *time.Time

	tStart, err := time.Parse("2006-01-02", data.DateStart)
	if err != nil {
		return ctl.request.BadRequest(c, fmt.Errorf("invalid date_start: %w", err))
	}
	startDate = &tStart

	tEnd, err := time.Parse("2006-01-02", data.DateEnd)
	if err != nil {
		return ctl.request.BadRequest(c, fmt.Errorf("invalid date_end: %w", err))
	}
	tEnd = tEnd.Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999999999*time.Nanosecond)
	endDate = &tEnd

	if tStart.After(*endDate) {
		return ctl.request.BadRequest(c, errors.New("date start is after end"))
	}

	stats, err := ctl.reportRepo.Statistics(c.Request().Context(), startDate, endDate)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, stats)
}

// TotalBandwidth 	 Ocserv Users TotalBandwidth calculating
//
// @Summary      Ocserv Users TotalBandwidth calculating
// @Description  Ocserv Users TotalBandwidth calculating
// @Tags         Report
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 date_start query string true "date_start"
// @Param 		 date_end query string true "date_end"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200 {object} repository.TotalBandwidths
// @Router       /reports/total-bandwidth [get]
func (ctl *Controller) TotalBandwidth(c echo.Context) error {
	var data TotalBandwidthData
	if err := c.Bind(&data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	var startDate, endDate *time.Time

	if data.DateStart != "" {
		t, err := time.Parse("2006-01-02", data.DateStart)
		if err != nil {
			return ctl.request.BadRequest(c, fmt.Errorf("invalid date_start: %w", err))
		}
		startDate = &t
	}

	if data.DateEnd != "" {
		t, err := time.Parse("2006-01-02", data.DateEnd)
		if err != nil {
			return ctl.request.BadRequest(c, fmt.Errorf("invalid date_end: %w", err))
		}
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999999999*time.Nanosecond)
		endDate = &t
	}

	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return ctl.request.BadRequest(c, errors.New("date start is after end"))
	}

	bandwidth, err := ctl.reportRepo.TotalBandwidthDateRange(c.Request().Context(), startDate, endDate)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, bandwidth)
}

// OcservUserReport     Result of all user reports
//
// @Summary      Result of all user reports
// @Description  Result of all user reports
// @Tags         Report
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200 {object} OcservUserReportResponse
// @Router       /reports/users [get]
func (ctl *Controller) OcservUserReport(c echo.Context) error {
	var wg sync.WaitGroup
	var onlineUsers []string
	var result repository.UserStatsResult

	errChan := make(chan error, 2)
	wg.Add(2)

	go func() {
		defer wg.Done()

		users, err := ctl.ocservOcctlRepo.OnlineSessions()
		if err != nil {
			errChan <- fmt.Errorf("failed to get online users: %w", err)
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

		res, err := ctl.reportRepo.UsersStat(c.Request().Context())
		if err != nil {
			errChan <- fmt.Errorf("failed to get users stats: %w", err)
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
		return ctl.request.BadRequest(c, errors.New(strings.Join(errs, "; ")))
	}

	return c.JSON(http.StatusOK, OcservUserReportResponse{
		Online:      len(onlineUsers),
		Active:      result.Active,
		Deactivated: result.Deactivated,
		Locked:      result.Locked,
	})
}
