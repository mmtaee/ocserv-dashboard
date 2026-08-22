package customer

import (
	"context"
	"strconv"
	"time"
)

func (u *Usecase) Summary(ctx context.Context, credentials Credentials) (*Summary, error) {
	user, err := u.authenticate(ctx, credentials)
	if err != nil {
		return nil, err
	}
	dateEnd := u.now()
	firstOfMonth := time.Date(dateEnd.Year(), dateEnd.Month(), 1, 0, 0, 0, 0, dateEnd.Location())
	dateStart := firstOfMonth.AddDate(0, -1, 0)
	bandwidths, err := u.users.TotalBandwidthUserDateRange(ctx, strconv.FormatUint(uint64(user.ID), 10), &dateStart, &dateEnd)
	if err != nil {
		return nil, err
	}
	return &Summary{OcservUser: customerFromModel(user), Usage: Usage{DateStart: dateStart, DateEnd: dateEnd, Bandwidths: bandwidths}}, nil
}

func (u *Usecase) Disconnect(ctx context.Context, credentials Credentials) error {
	user, err := u.authenticate(ctx, credentials)
	if err != nil {
		return err
	}
	_, err = u.occtl.Disconnect(user.Username)
	return err
}
