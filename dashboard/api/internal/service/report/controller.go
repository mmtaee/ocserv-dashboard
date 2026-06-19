package report

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/occtl"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
)

type ReportController struct {
	reportUC  usecase.ReportUseCase
	req       *request.Request
	validator *request.Validator
	occtl     *occtl.OcservOcctl
}

func NewReportController(reportUC usecase.ReportUseCase) *ReportController {
	return &ReportController{
		reportUC:  reportUC,
		req:       &request.Request{},
		validator: request.NewValidator(),
		occtl:     occtl.NewOcservOcctl(),
	}
}

// SessionLogs godoc
// @Summary Ocserv session logs
// @Description Get all Ocserv session logs
// @Tags Report
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param page query int false "Page number, starting from 1" minimum(1)
// @Param limit query int false "Number of items per page" minimum(1) maximum(100)
// @Param order_by query string false "Field to order by" Enums(id, created_at)
// @Param sort query string false "Sort order" Enums(asc, desc)
// @Param date_start query string false "date_start"
// @Param date_end query string false "date_end"
// @Success 200 {object} SessionLogsResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /reports/session_logs [get]
func (ctrl *ReportController) SessionLogs(c *echo.Context) error {
	var data SessionLogsData
	if err := ctrl.validator.Validate(c, &data); err != nil {
		return err
	}

	page, limit := request.GetPaginationParams(c)
	orderBy := c.QueryParam("order_by")
	sort := c.QueryParam("sort")

	var startDate, endDate *time.Time
	if data.DateStart != "" {
		t, err := time.Parse("2006-01-02", data.DateStart)
		if err == nil {
			startDate = &t
		}
	}

	if data.DateEnd != "" {
		t, err := time.Parse("2006-01-02", data.DateEnd)
		if err == nil {
			e := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &e
		}
	}

	logs, total, err := ctrl.reportUC.SessionLogs(page, limit, orderBy, sort, startDate, endDate)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}

	return c.JSON(http.StatusOK, SessionLogsResponse{
		Meta:   request.NewPagination(total, page, limit),
		Result: logs,
	})
}

// Statistics godoc
// @Summary Ocserv Users Statistics
// @Description Get Ocserv users daily statistics
// @Tags Report
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param date_start query string true "date_start"
// @Param date_end query string true "date_end"
// @Success 200 {object} []models.DailyTraffic
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /reports/statistics [get]
func (ctrl *ReportController) Statistics(c *echo.Context) error {
	var data StatisticsData
	if err := ctrl.validator.Validate(c, &data); err != nil {
		return err
	}

	if data.DateStart == "" || data.DateEnd == "" {
		return ctrl.req.BadRequest(c, errors.New("statistics date start and end are required"))
	}

	var startDate, endDate *time.Time
	tStart, err := time.Parse("2006-01-02", data.DateStart)
	if err != nil {
		return ctrl.req.BadRequest(c, fmt.Errorf("invalid date_start: %w", err))
	}
	startDate = &tStart

	tEnd, err := time.Parse("2006-01-02", data.DateEnd)
	if err != nil {
		return ctrl.req.BadRequest(c, fmt.Errorf("invalid date_end: %w", err))
	}
	tEnd = tEnd.Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999999999*time.Nanosecond)
	endDate = &tEnd

	if tStart.After(*endDate) {
		return ctrl.req.BadRequest(c, errors.New("date start is after end"))
	}

	stats, err := ctrl.reportUC.Statistics(startDate, endDate)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, stats)
}

// TotalBandwidth godoc
// @Summary Ocserv Users Total Bandwidth
// @Description Calculate total bandwidth usage over date range
// @Tags Report
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param date_start query string false "date_start"
// @Param date_end query string false "date_end"
// @Success 200 {object} repository.TotalBandwidths
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /reports/total-bandwidth [get]
func (ctrl *ReportController) TotalBandwidth(c *echo.Context) error {
	var data TotalBandwidthData
	if err := ctrl.validator.Validate(c, &data); err != nil {
		return err
	}

	var startDate, endDate *time.Time
	if data.DateStart != "" {
		t, err := time.Parse("2006-01-02", data.DateStart)
		if err == nil {
			startDate = &t
		}
	}

	if data.DateEnd != "" {
		t, err := time.Parse("2006-01-02", data.DateEnd)
		if err == nil {
			e := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999999999*time.Nanosecond)
			endDate = &e
		}
	}

	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return ctrl.req.BadRequest(c, errors.New("date start is after end"))
	}

	bandwidth, err := ctrl.reportUC.TotalBandwidthDateRange(startDate, endDate)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, bandwidth)
}

// OcservUserReport godoc
// @Summary Result of all user reports
// @Description Get summary statistics for Ocserv users
// @Tags Report
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Success 200 {object} OcservUserReportResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Router /reports/users [get]
func (ctrl *ReportController) OcservUserReport(c *echo.Context) error {
	var wg sync.WaitGroup
	var onlineUsers []string
	var result repository.UserStatsResult

	errChan := make(chan error, 2)
	wg.Add(2)

	go func() {
		defer wg.Done()
		usersPtr, err := ctrl.occtl.OnlineSessions()
		if err != nil {
			errChan <- fmt.Errorf("failed to get online users: %w", err)
			return
		}
		onlineUsernames := make([]string, 0)
		if usersPtr != nil {
			for _, u := range *usersPtr {
				if !slices.Contains(onlineUsernames, u.Username) {
					onlineUsernames = append(onlineUsernames, u.Username)
				}
			}
		}
		onlineUsers = onlineUsernames
	}()

	adminID := c.Get("id").(uint)
	role := c.Get("role").(string)
	go func() {
		defer wg.Done()
		res, err := ctrl.reportUC.UsersStat(adminID, role)
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
		return ctrl.req.BadRequest(c, errors.New(errs[0]))
	}

	return c.JSON(http.StatusOK, OcservUserReportResponse{
		Online:      len(onlineUsers),
		Active:      result.Active,
		Deactivated: result.Deactivated,
		Locked:      result.Locked,
	})
}
