package home

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-users-management/api/internal/repository"
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"net/http"
	"sync"
)

type Controller struct {
	request        request.CustomRequestInterface
	occtlRepo      repository.OcctlRepositoryInterface
	ocservUserRepo repository.OcservUserRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request:        request.NewCustomRequest(),
		occtlRepo:      repository.NewOcctlRepository(),
		ocservUserRepo: repository.NewtOcservUserRepository(),
	}
}

// Home 	     Content of home
//
// @Summary      Content of home
// @Description  Content of home
// @Tags         Home
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} GetHomeResponse
// @Router       /home [get]
func (ctl *Controller) Home(c echo.Context) error {
	ctx := c.Request().Context()

	var (
		status           ServerStatusResponse
		statistics       *[]models.DailyTraffic
		onlineUsers      *[]models.OnlineUserSession
		TotalUser        int64
		ipBans           *[]models.IPBanPoints
		topBandwidthUser repository.TopBandwidthUsers
		totalBandwidth   repository.TotalBandwidths
		errs             = make(chan error, 4)
		wg               sync.WaitGroup
	)

	wg.Add(7)

	go func() {
		defer wg.Done()
		serverStatus, err := ctl.occtlRepo.Status()
		if err != nil {
			errs <- err
			return
		}
		if serverStatusMap, ok := serverStatus.(map[string]interface{}); ok {
			status = ParseServerStatus(serverStatusMap)
		}
	}()

	go func() {
		defer wg.Done()
		data, err := ctl.ocservUserRepo.TenDaysStats(ctx)
		if err != nil {
			errs <- err
			return
		}
		statistics = data
	}()

	go func() {
		defer wg.Done()
		users, err := ctl.occtlRepo.OnlineUsersInfo()
		if err != nil {
			errs <- err
			return
		}
		onlineUsers = users
	}()

	go func() {
		defer wg.Done()
		ips, err := ctl.occtlRepo.IPBans()
		if err != nil {
			errs <- err
			return
		}
		ipBans = ips
	}()

	go func() {
		defer wg.Done()
		users, err := ctl.ocservUserRepo.TotalUsers(ctx)
		if err != nil {
			errs <- err
			return
		}
		TotalUser = users
	}()

	go func() {
		defer wg.Done()
		topUser, err := ctl.ocservUserRepo.TopBandwidthUser(ctx)
		if err != nil {
			errs <- err
			return
		}
		topBandwidthUser = topUser
	}()

	go func() {
		defer wg.Done()
		bandwidth, err := ctl.ocservUserRepo.TotalBandwidth(ctx)
		if err != nil {
			errs <- err
			return
		}
		totalBandwidth = bandwidth
	}()

	wg.Wait()
	close(errs)

	if err := <-errs; err != nil {
		logger.Warn("error in Home handler: %v", err)
		return ctl.request.BadRequest(c, err)
	}

	resp := GetHomeResponse{
		ServerStatus: status,
		Statistics:   statistics,
		IPBans:       ipBans,
		Users: GetHomeUser{
			Total:  TotalUser,
			Online: onlineUsers,
		},
		TopBandwidthUser: topBandwidthUser,
		TotalBandwidth:   totalBandwidth,
	}

	return c.JSON(http.StatusOK, resp)
}
