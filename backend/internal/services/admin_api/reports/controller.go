package reports

import (
	"net/http"

	"github.com/labstack/echo/v5"
	reportusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/reports"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request    request.CustomRequestInterface
	reportRepo *reportusecase.Usecase
}

func New(reports *reportusecase.Usecase) *Controller {
	return &Controller{
		request:    request.NewCustomRequest(),
		reportRepo: reports,
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
func (ctl *Controller) SessionLogs(c *echo.Context) error {
	var data SessionLogsData
	if err := c.Bind(&data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	pagination := ctl.request.Pagination(c)

	logs, total, err := ctl.reportRepo.SessionLogs(c.Request().Context(), pagination, data)
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
func (ctl *Controller) Statistics(c *echo.Context) error {
	var data StatisticsData
	if err := c.Bind(&data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	stats, err := ctl.reportRepo.Statistics(c.Request().Context(), data)
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
func (ctl *Controller) TotalBandwidth(c *echo.Context) error {
	var data TotalBandwidthData
	if err := c.Bind(&data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	bandwidth, err := ctl.reportRepo.TotalBandwidthDateRange(c.Request().Context(), data)
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
func (ctl *Controller) OcservUserReport(c *echo.Context) error {
	result, err := ctl.reportRepo.UserReport(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}
