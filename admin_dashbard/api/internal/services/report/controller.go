package report

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/models"
)

type SessionLogsData struct {
	DateStart string `json:"date_start" query:"date_start" validate:"omitempty" example:"2025-1-31"`
	DateEnd   string `json:"date_end" query:"date_end" validate:"omitempty" example:"2025-12-31"`
}

type SessionLogsResponse struct {
	Meta   request.Meta                   `json:"meta" validate:"required"`
	Result []models.OcservUserSessionLog `json:"result" validate:"omitempty"`
}

type StatisticsData struct {
	DateStart string `json:"date_start" query:"date_start" validate:"omitempty" example:"2025-1-31"`
	DateEnd   string `json:"date_end" query:"date_end" validate:"omitempty" example:"2025-12-31"`
}

type StatisticsResponse struct {
	Statistics      []models.DailyTraffic      `json:"statistics" validate:"required"`
	TotalBandwidths repository.TotalBandwidths `json:"total_bandwidths" validate:"required"`
}

type TotalBandwidthData struct {
	DateStart string `json:"date_start" query:"date_start" validate:"omitempty" example:"2025-1-31"`
	DateEnd   string `json:"date_end" query:"date_end" validate:"omitempty" example:"2025-12-31"`
}

type OcservUserReportResponse struct {
	Online      int   `json:"online"`
	Active      int64 `json:"active"`
	Deactivated int64 `json:"deactivated"`
	Locked      int64 `json:"locked"`
}

type Controller struct {
	request request.CustomRequestInterface
	usecase usecase.ReportUsecaseInterface
}

func New(reportUsecase usecase.ReportUsecaseInterface) *Controller {
	return &Controller{
		request: request.NewCustomRequest(),
		usecase: reportUsecase,
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
		return ctl.request.BadRequest(c, err, "1003")
	}

	pagination := ctl.request.Pagination(c)

	var startDate, endDate *time.Time

	if data.DateStart != "" {
		t, err := time.Parse("2006-01-02", data.DateStart)
		if err != nil {
			return ctl.request.BadRequest(c, fmt.Errorf("invalid date_start: %w", err), "1010")
		}
		startDate = &t
	}

	if data.DateEnd != "" {
		t, err := time.Parse("2006-01-02", data.DateEnd)
		if err != nil {
			return ctl.request.BadRequest(c, fmt.Errorf("invalid date_end: %w", err), "1011")
		}
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		endDate = &t
	}

	logs, total, err := ctl.usecase.SessionLogs(c.Request().Context(), pagination, startDate, endDate)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
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
		return ctl.request.BadRequest(c, err, "1003")
	}

	if data.DateStart == "" || data.DateEnd == "" {
		return ctl.request.BadRequest(c, errors.New("statistics date start and end are required"), "1031")
	}

	var startDate, endDate *time.Time

	tStart, err := time.Parse("2006-01-02", data.DateStart)
	if err != nil {
		return ctl.request.BadRequest(c, fmt.Errorf("invalid date_start: %w", err), "1010")
	}
	startDate = &tStart

	tEnd, err := time.Parse("2006-01-02", data.DateEnd)
	if err != nil {
		return ctl.request.BadRequest(c, fmt.Errorf("invalid date_end: %w", err), "1011")
	}
	tEnd = tEnd.Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999999999*time.Nanosecond)
	endDate = &tEnd

	if tStart.After(*endDate) {
		return ctl.request.BadRequest(c, errors.New("date start is after end"), "1030")
	}

	stats, err := ctl.usecase.Statistics(c.Request().Context(), startDate, endDate)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
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
		return ctl.request.BadRequest(c, err, "1003")
	}

	var startDate, endDate *time.Time

	if data.DateStart != "" {
		t, err := time.Parse("2006-01-02", data.DateStart)
		if err != nil {
			return ctl.request.BadRequest(c, fmt.Errorf("invalid date_start: %w", err), "1010")
		}
		startDate = &t
	}

	if data.DateEnd != "" {
		t, err := time.Parse("2006-01-02", data.DateEnd)
		if err != nil {
			return ctl.request.BadRequest(c, fmt.Errorf("invalid date_end: %w", err), "1011")
		}
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999999999*time.Nanosecond)
		endDate = &t
	}

	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return ctl.request.BadRequest(c, errors.New("date start is after end"), "1030")
	}

	bandwidth, err := ctl.usecase.TotalBandwidth(c.Request().Context(), startDate, endDate)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
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
	online, userStats, err := ctl.usecase.UsersReport(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(http.StatusOK, OcservUserReportResponse{
		Online:      online,
		Active:      userStats.Active,
		Deactivated: userStats.Deactivated,
		Locked:      userStats.Locked,
	})
}
