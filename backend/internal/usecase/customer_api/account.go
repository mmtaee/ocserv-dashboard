package customer

import (
	"context"
	"errors"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

func (u *Usecase) Summary(ctx context.Context, username string) (*Summary, error) {
	user, err := u.user(ctx, username)
	if err != nil {
		return nil, err
	}
	dateEnd := u.now()
	firstOfMonth := time.Date(dateEnd.Year(), dateEnd.Month(), 1, 0, 0, 0, 0, dateEnd.Location())
	dateStart := firstOfMonth.AddDate(0, -1, 0)
	bandwidths, err := u.users.TotalBandwidthUserDateRange(ctx, user.ID, &dateStart, &dateEnd)
	if err != nil {
		return nil, err
	}
	return &Summary{OcservUser: customerFromModel(user), Usage: Usage{DateStart: dateStart, DateEnd: dateEnd, Bandwidths: bandwidths}}, nil
}

func (u *Usecase) Sessions(_ context.Context, username string) ([]models.OnlineUserSession, error) {
	sessions, err := u.occtl.OnlineSessions()
	if err != nil {
		return nil, err
	}
	result := make([]models.OnlineUserSession, 0)
	for _, session := range sessions {
		if session.Username == username {
			result = append(result, session)
		}
	}
	return result, nil
}

func (u *Usecase) Disconnect(_ context.Context, username string) error {
	_, err := u.occtl.Disconnect(username)
	return err
}

func (u *Usecase) Terminate(_ context.Context, username string) error {
	_, err := u.occtl.Terminate(username)
	return err
}

func (u *Usecase) ChangePassword(ctx context.Context, username, password string) error {
	user, err := u.user(ctx, username)
	if err != nil {
		return err
	}
	previous := user.Password
	user.Password = password
	if _, err = u.users.Update(ctx, user); err != nil {
		return err
	}
	if err = u.certificates.Create(user.Group, user.Username, password, user.Config); err != nil {
		user.Password = previous
		_, _ = u.users.Update(ctx, user)
	}
	return err
}

func (u *Usecase) Statistics(ctx context.Context, username string, input DateRange) ([]models.DailyTraffic, error) {
	user, err := u.user(ctx, username)
	if err != nil {
		return nil, err
	}
	start, end, err := parseDateRange(input)
	if err != nil {
		return nil, err
	}
	return u.users.UserStatistics(ctx, user.ID, start, end)
}

func (u *Usecase) Activities(ctx context.Context, username string, pagination *request.Pagination, input DateRange) (*Activities, error) {
	start, end, err := parseDateRange(input)
	if err != nil {
		return nil, err
	}
	logs, total, err := u.users.UserSessionLogs(ctx, pagination, username, start, end)
	return &Activities{Logs: logs, Total: total}, err
}

func (u *Usecase) Bandwidth(ctx context.Context, username string, input DateRange) (repository.TotalBandwidths, error) {
	user, err := u.user(ctx, username)
	if err != nil {
		return repository.TotalBandwidths{}, err
	}
	start, end, err := parseDateRange(input)
	if err != nil {
		return repository.TotalBandwidths{}, err
	}
	return u.users.TotalBandwidthUserDateRange(ctx, user.ID, start, end)
}

func parseDateRange(input DateRange) (*time.Time, *time.Time, error) {
	parse := func(value string) (*time.Time, error) {
		if value == "" {
			return nil, nil
		}
		result, err := time.Parse("2006-01-02", value)
		return &result, err
	}
	start, err := parse(input.DateStart)
	if err != nil {
		return nil, nil, errors.New("invalid date_start")
	}
	end, err := parse(input.DateEnd)
	if err != nil {
		return nil, nil, errors.New("invalid date_end")
	}
	if end != nil {
		value := end.Add(24*time.Hour - time.Nanosecond)
		end = &value
	}
	return start, end, nil
}
