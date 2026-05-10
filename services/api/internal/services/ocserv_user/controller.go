package ocserv_user

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	"github.com/mmtaee/ocserv-dashboard/common/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"net/http"
	"slices"
	"sync"
	"time"
)

type Controller struct {
	request         request.CustomRequestInterface
	userRepo        repository.UserRepositoryInterface
	ocservUserRepo  repository.OcservUserRepositoryInterface
	ocservOcctlRepo repository.OcctlRepositoryInterface
	reportRepo      repository.ReportRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request:         request.NewCustomRequest(),
		ocservUserRepo:  repository.NewtOcservUserRepository(),
		ocservOcctlRepo: repository.NewOcctlRepository(),
		reportRepo:      repository.NewtReportRepository(),
	}
}

// OcservUsers 	 List of Ocserv Users
//
// @Summary      List of Ocserv Users
// @Description  List of Ocserv Users
// @Tags         Ocserv(Users)
// @Accept       json
// @Produce      json
// @Param 		 page query int false "Page number, starting from 1" minimum(1)
// @Param 		 size query int false "Number of items per page" minimum(1) maximum(100) name(size)
// @Param 		 order query string false "Field to order by"
// @Param 		 sort query string false "Sort order, either ASC or DESC" Enums(ASC, DESC)
// @Param 		 q query string false "ocserv username q search" minLength(2)
// @Param 		 filter query string false "filter ocserv user by statues" Enums(online, active, deactivated, locked)
// @Param 		 group query string false "filter ocserv user by group name"
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object}  OcservUsersResponse
// @Router       /ocserv/users [get]
func (ctl *Controller) OcservUsers(c echo.Context) error {
	owner := ""

	val, ok := c.Get("isAdmin").(bool)
	if !ok || !val {
		usernameVal, ok := c.Get("username").(string)
		if !ok || usernameVal == "" {
			return ctl.request.BadRequest(c, errors.New("invalid user uid"))
		}
		owner = usernameVal
	}

	q := c.QueryParam("q")
	group := c.QueryParam("group")
	pagination := ctl.request.Pagination(c)

	filter := c.QueryParam("filter")
	switch filter {
	case "online", "active", "deactivated", "locked":
	default:
		filter = ""
	}

	ctx := c.Request().Context()

	// -------------------------
	// ONLINE FILTER MODE
	// -------------------------
	if filter == "online" {
		onlineUsers, err := ctl.ocservOcctlRepo.OnlineUsers()
		if err != nil {
			return ctl.request.BadRequest(c, err)
		}

		users, total, err := ctl.ocservUserRepo.UsersByUsername(
			ctx,
			pagination,
			owner,
			onlineUsers,
			q,
			group,
		)
		if err != nil {
			return ctl.request.BadRequest(c, err)
		}

		// Users are already filtered to occtl-reported sessions; IsOnline must be true
		// or the UI shows "disconnected" (normal-list branch sets this via onlineMap).
		for i := range users {
			users[i].IsOnline = true
		}

		return c.JSON(http.StatusOK, OcservUsersResponse{
			Meta: request.Meta{
				Page:         pagination.Page,
				TotalRecords: total,
				PageSize:     pagination.PageSize,
			},
			Result: users,
		})
	}

	// -------------------------
	// NORMAL MODE
	// -------------------------
	users, total, err := ctl.ocservUserRepo.Users(
		ctx,
		pagination,
		owner,
		q,
		filter,
		group,
	)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	// attach online status
	if len(users) > 0 {
		onlineUsers, err := ctl.ocservOcctlRepo.OnlineUsers()
		if err != nil {
			return ctl.request.BadRequest(c, err)
		}

		onlineMap := make(map[string]struct{}, len(onlineUsers))
		for _, u := range onlineUsers {
			onlineMap[u] = struct{}{}
		}

		for i := range users {
			if _, ok := onlineMap[users[i].Username]; ok {
				users[i].IsOnline = true
			}
		}
	}

	return c.JSON(http.StatusOK, OcservUsersResponse{
		Meta: request.Meta{
			Page:         pagination.Page,
			TotalRecords: total,
			PageSize:     pagination.PageSize,
		},
		Result: users,
	})
}

// OcservUser 	 Ocserv user detail
//
// @Summary      Ocserv user detail
// @Description  Ocserv user detail
// @Tags         Ocserv(Users)
// @Accept       json
// @Produce      json
// @Param 		 uid path string true "Ocserv User UID"
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object}  models.OcservUser
// @Router       /ocserv/users/{uid} [get]
func (ctl *Controller) OcservUser(c echo.Context) error {
	// TODO: add staff filter to get ocserv user for same owner
	userUID := c.Param("uid")
	if userUID == "" {
		return ctl.request.BadRequest(c, errors.New("invalid user uid"))
	}

	u, err := ctl.ocservUserRepo.GetByUID(c.Request().Context(), userUID)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, u)
}

// CreateOcservUser 	     Ocserv User creation
//
// @Summary      Ocserv User creation
// @Description  Ocserv User creation
// @Tags         Ocserv(Users)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param        request    body  CreateOcservUserData  true "ocserv user create data"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      201  {object} models.OcservUser
// @Router       /ocserv/users [post]
func (ctl *Controller) CreateOcservUser(c echo.Context) error {
	var data CreateOcservUserData

	owner := c.Get("username").(string)
	if owner == "" {
		return ctl.request.BadRequest(c, errors.New("admin or staff username not found"))
	}

	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	var expireAt *time.Time
	if data.Unlimited {
		expireAt = nil
	} else {
		expireAtTime, err := time.Parse("2006-01-02", data.ExpireAt)
		if err != nil {
			t := time.Now().AddDate(0, 0, 30)
			expireAt = &t
		} else {
			expireAt = &expireAtTime
		}
	}

	if data.TrafficType == models.Free {
		data.TrafficSize = 0
	}

	ocUser := &models.OcservUser{
		Owner:       owner,
		Username:    data.Username,
		Password:    data.Password,
		Group:       data.Group,
		ExpireAt:    expireAt,
		TrafficSize: data.TrafficSize,
		TrafficType: data.TrafficType,
		Config:      data.Config,
	}

	u, err := ctl.ocservUserRepo.Create(c.Request().Context(), ocUser)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusCreated, u)
}

// UpdateOcservUser 	     Ocserv User update
//
// @Summary      Ocserv User update
// @Description  Ocserv User update
// @Tags         Ocserv(Users)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 uid path string true "Ocserv User UID"
// @Param        request    body  UpdateOcservUserData  true "ocserv user update data"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      201  {object} models.OcservUser
// @Router       /ocserv/users/{uid} [patch]
func (ctl *Controller) UpdateOcservUser(c echo.Context) error {
	userID := c.Param("uid")
	if userID == "" {
		return ctl.request.BadRequest(c, errors.New("user id is required"))
	}

	var data UpdateOcservUserData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	ocservUser, err := ctl.ocservUserRepo.GetByUID(c.Request().Context(), userID)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	if data.Group != nil {
		ocservUser.Group = *data.Group
	}
	if data.Password != nil {
		ocservUser.Password = *data.Password
	}
	if data.Description != nil {
		ocservUser.Description = *data.Description
	}
	if data.TrafficSize != nil {
		ocservUser.TrafficSize = *data.TrafficSize
	}
	if data.TrafficType != nil && slices.Contains([]string{
		"Free",
		"MonthlyTransmit",
		"MonthlyReceive",
		"MonthlyRxTx",
		"TotallyTransmit",
		"TotallyReceive",
		"TotallyRxTx",
	}, *data.TrafficType) {
		ocservUser.TrafficType = *data.TrafficType
	}
	if data.Config != nil {
		ocservUser.Config = data.Config
	}

	if data.Unlimited {
		ocservUser.ExpireAt = nil
	} else if data.ExpireAt != nil {
		if expire, err := time.Parse("2006-01-02", *data.ExpireAt); err == nil {
			ocservUser.ExpireAt = &expire
		}
	}

	updatedOcservUser, err := ctl.ocservUserRepo.Update(c.Request().Context(), ocservUser)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, updatedOcservUser)
}

// DeleteOcservUser 	     Ocserv User delete
//
// @Summary      Ocserv User delete
// @Description  Ocserv User delete
// @Tags         Ocserv(Users)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 uid path string true "Ocserv User UID"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      204  {object} nil
// @Router       /ocserv/users/{uid} [delete]
func (ctl *Controller) DeleteOcservUser(c echo.Context) error {
	userID := c.Param("uid")
	if userID == "" {
		return ctl.request.BadRequest(c, errors.New("user id is required"))
	}

	username, err := ctl.ocservUserRepo.Delete(c.Request().Context(), userID)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	go func() {
		_, _ = ctl.ocservOcctlRepo.Disconnect(username)
	}()

	return c.JSON(http.StatusNoContent, nil)
}

// LockOcservUser 	     Ocserv User locking
//
// @Summary      Ocserv User locking
// @Description  Ocserv User locking
// @Tags         Ocserv(Users)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 uid path string true "Ocserv User UID"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} nil
// @Router       /ocserv/users/{uid}/lock [post]
func (ctl *Controller) LockOcservUser(c echo.Context) error {
	userID := c.Param("uid")
	if userID == "" {
		return ctl.request.BadRequest(c, errors.New("user id is required"))
	}

	err := ctl.ocservUserRepo.Lock(c.Request().Context(), userID)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		u, err := ctl.ocservUserRepo.GetByUID(ctx, userID)
		if err != nil {
			logger.Error("failed to fetch ocserv user error: ", err)
		}
		_, err = ctl.ocservOcctlRepo.Disconnect(u.Username)
		if err != nil {
			logger.Error("failed to disconnect ocserv user error: ", err)
		}
		return
	}()

	return c.JSON(http.StatusOK, nil)
}

// UnLockOcservUser 	     Ocserv User unlocking
//
// @Summary      Ocserv User unlocking
// @Description  Ocserv User unlocking
// @Tags         Ocserv(Users)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 uid path string true "Ocserv User UID"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} nil
// @Router       /ocserv/users/{uid}/unlock [post]
func (ctl *Controller) UnLockOcservUser(c echo.Context) error {
	userID := c.Param("uid")
	if userID == "" {
		return ctl.request.BadRequest(c, errors.New("user id is required"))
	}

	err := ctl.ocservUserRepo.UnLock(c.Request().Context(), userID)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// DisconnectOcservUser 	     Ocserv User disconnecting
//
// @Summary      Disconnect Ocserv User
// @Description  Disconnect Ocserv User
// @Tags         Ocserv(Users)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 username path string true "Ocserv User username"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} nil
// @Router       /ocserv/users/{username}/disconnect [post]
func (ctl *Controller) DisconnectOcservUser(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return ctl.request.BadRequest(c, errors.New("user id is required"))
	}
	_, err := ctl.ocservOcctlRepo.Disconnect(username)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, nil)
}

// OcservUserStatistics 	     Ocserv User Statistics
//
// @Summary      Ocserv User Statistics
// @Description  Ocserv User Statistics
// @Tags         Ocserv(Users)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 uid path string true "Ocserv User UID"
// @Param 		 date_start query string false "date_start"
// @Param 		 date_end query string false "date_end"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} StatisticsResponse
// @Router       /ocserv/users/{uid}/statistics [get]
func (ctl *Controller) OcservUserStatistics(c echo.Context) error {
	userID := c.Param("uid")
	if userID == "" {
		return ctl.request.BadRequest(c, errors.New("user id is required"))
	}

	var data StatisticsData
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
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		endDate = &t
	}

	ctx := c.Request().Context()
	var (
		stats []models.DailyTraffic
		total repository.TotalBandwidths
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		s, err := ctl.ocservUserRepo.UserStatistics(ctx, userID, startDate, endDate)
		if err != nil {
			return err
		}
		stats = s
		return nil
	})

	g.Go(func() error {
		t, err := ctl.reportRepo.TotalBandWidthUser(ctx, userID)
		if err != nil {
			return err
		}
		total = t
		return nil
	})

	if err := g.Wait(); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, StatisticsResponse{
		Statistics:      stats,
		TotalBandwidths: total,
	})
}

// OcpasswdUsers  Ocserv Users from ocpasswd file
//
// @Summary      Ocserv Users from ocpasswd file
// @Description  Ocserv Users from ocpasswd file
// @Tags         Ocserv(Ocpasswd)
// @Accept       json
// @Produce      json
// @Param 		 page query int false "Page number, starting from 1" minimum(1)
// @Param 		 size query int false "Number of items per page" minimum(1) maximum(100) name(size)
// @Param 		 order query string false "Field to order by"
// @Param 		 sort query string false "Sort order, either ASC or DESC" Enums(ASC, DESC)
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200 {object} OcservUsersSyncResponse
// @Router       /ocserv/users/ocpasswd [get]
func (ctl *Controller) OcpasswdUsers(c echo.Context) error {
	pagination := ctl.request.Pagination(c)

	users, total, err := ctl.ocservUserRepo.Ocpasswd(c.Request().Context(), pagination)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, OcservUsersSyncResponse{
		Meta: request.Meta{
			Page:         pagination.Page,
			TotalRecords: int64(total),
			PageSize:     pagination.PageSize,
		},
		Result: users,
	})
}

// SyncToDB      Ocserv Users from ocpasswd file to db
//
// @Summary      Ocserv Users from ocpasswd file to db
// @Description  Ocserv Users from ocpasswd file to db
// @Tags         Ocserv(Ocpasswd)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param        request    body  SyncOcpasswdRequest  true "list of users with config to sync in db"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200 {object} []string
// @Router       /ocserv/users/ocpasswd/sync [post]
func (ctl *Controller) SyncToDB(c echo.Context) error {
	owner := c.Get("username").(string)
	if owner == "" {
		return ctl.request.BadRequest(c, errors.New("admin or staff username not found"))
	}

	var data SyncOcpasswdRequest
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	expireAt, err := time.Parse("2006-01-02", *data.ExpireAt)
	if err != nil {
		expireAt, _ = time.Parse("2006-01-02", time.Now().AddDate(0, 0, 30).Format("2006-01-02"))
	}

	var users []models.OcservUser
	var wg sync.WaitGroup
	var mux sync.Mutex

	for _, u := range data.Users {
		wg.Add(1)

		go func(u user.Ocpasswd) {
			defer wg.Done()

			newUser := models.OcservUser{
				Username:    u.Username,
				Password:    "Secret-Ocpasswd",
				Group:       u.Group,
				Owner:       owner,
				ExpireAt:    &expireAt,
				TrafficSize: *data.TrafficSize,
				TrafficType: *data.TrafficType,
				Config:      data.Config,
			}

			mux.Lock()
			users = append(users, newUser)
			mux.Unlock()
		}(u)
	}
	wg.Wait()

	if len(users) == 0 {
		return ctl.request.BadRequest(c, errors.New("no users found"))
	}

	syncUsers, err := ctl.ocservUserRepo.OcpasswdSyncToDB(c.Request().Context(), users)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	var syncUsernames []string

	for _, u := range syncUsers {
		syncUsernames = append(syncUsernames, u.Username)
	}

	return c.JSON(http.StatusOK, syncUsernames)
}

// ActivateExpiredOcservUsers     Restore and activate expired Ocserv User accounts
//
// @Summary      Restore and activate expired Ocserv User accounts
// @Description  Restore and activate expired Ocserv User accounts
// @Tags         Ocserv(Users)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 uid path string true "Ocserv User UID"
// @Param        request    body  ActivateUserData  true "list of ocserv users and expire time"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200 {object} nil
// @Router       /ocserv/users/{uid}/activate [post]
func (ctl *Controller) ActivateExpiredOcservUsers(c echo.Context) error {
	userID := c.Param("uid")
	if userID == "" {
		return ctl.request.BadRequest(c, errors.New("user id is required"))
	}

	var data ActivateUserData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	var (
		expireAt *time.Time
		err      error
	)
	if data.ExpireAt != nil {
		expireAtTime, err := time.Parse("2006-01-02", *data.ExpireAt)
		if err == nil {
			expireAt = &expireAtTime
		}
	}

	err = ctl.ocservUserRepo.RestoreExpired(c.Request().Context(), userID, expireAt)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, nil)
}

// OcservUserSessionLogs 	     Ocserv User session logs
//
// @Summary      Ocserv User session logs
// @Description  Ocserv User session logs
// @Tags         Ocserv(Users)
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Param 		 page query int false "Page number, starting from 1" minimum(1)
// @Param 		 size query int false "Number of items per page" minimum(1) maximum(100) name(size)
// @Param 		 order query string false "Field to order by"
// @Param 		 sort query string false "Sort order, either ASC or DESC" Enums(ASC, DESC)
// @Param 		 uid path string true "Ocserv User UID"
// @Param 		 date_start query string false "date_start"
// @Param 		 date_end query string false "date_end"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} SessionLogsResponse
// @Router       /ocserv/users/{uid}/session_logs [get]
func (ctl *Controller) OcservUserSessionLogs(c echo.Context) error {
	userID := c.Param("uid")
	if userID == "" {
		return ctl.request.BadRequest(c, errors.New("user id is required"))
	}

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

	u, err := ctl.ocservUserRepo.GetByUID(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, nil)
		}
		return ctl.request.BadRequest(c, err)
	}

	logs, total, err := ctl.ocservUserRepo.UserSessionLogs(c.Request().Context(), pagination, u.Username, startDate, endDate)
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
