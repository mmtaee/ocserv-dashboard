package ocservuser

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("ocserv user not found")

func (u *Usecase) List(ctx context.Context, options ListOptions) (*ListResult, error) {
	filter := options.Filter
	if !slices.Contains([]string{"online", "active", "deactivated", "locked"}, filter) {
		filter = ""
	}
	sessions, sessionErr := u.occtl.OnlineSessions()
	byUsername := make(map[string][]models.OnlineUserSession)
	usernames := make([]string, 0)
	seen := make(map[string]struct{})
	for _, session := range sessions {
		if _, ok := seen[session.Username]; !ok {
			seen[session.Username] = struct{}{}
			usernames = append(usernames, session.Username)
		}
		byUsername[session.Username] = append(byUsername[session.Username], session)
	}
	var users []models.OcservUser
	var total int64
	var err error
	if filter == "online" {
		if sessionErr != nil {
			return nil, sessionErr
		}
		users, total, err = u.Repository.UsersByUsername(ctx, options.Pagination, options.Owner, usernames, options.Query, options.Group)
	} else {
		users, total, err = u.Repository.Users(ctx, options.Pagination, options.Owner, options.Query, filter, options.Group)
	}
	if err != nil {
		return nil, err
	}
	for i := range users {
		u.applyCertificateStatus(&users[i])
		if online, ok := byUsername[users[i].Username]; ok {
			users[i].IsOnline = true
			users[i].OnlineUserSessions = online
		}
	}
	return &ListResult{Users: users, Total: total}, nil
}

func (u *Usecase) User(ctx context.Context, uid string) (*models.OcservUser, error) {
	if uid == "" {
		return nil, errors.New("invalid user uid")
	}
	return u.GetByUID(ctx, uid)
}

func (u *Usecase) GetByUID(ctx context.Context, uid string) (*models.OcservUser, error) {
	user, err := u.Repository.GetByUID(ctx, uid)
	if err == nil {
		u.applyCertificateStatus(user)
	}
	return user, err
}

func (u *Usecase) GetByUsername(ctx context.Context, username string) (*models.OcservUser, error) {
	user, err := u.Repository.GetByUsername(ctx, username)
	if err == nil {
		u.applyCertificateStatus(user)
	}
	return user, err
}

func (u *Usecase) CreateUser(ctx context.Context, owner string, input CreateOcservUserData) (*models.OcservUser, error) {
	if owner == "" {
		return nil, errors.New("admin or staff username not found")
	}
	var expiresAt *time.Time
	if !input.Unlimited {
		parsed, err := time.Parse("2006-01-02", input.ExpireAt)
		if err != nil {
			parsed = time.Now().AddDate(0, 0, 30)
		}
		expiresAt = &parsed
	}
	if input.TrafficType == models.Free {
		input.TrafficSize = 0
	}
	return u.Create(ctx, &models.OcservUser{
		Owner: owner, Username: input.Username, Password: input.Password, Group: input.Group, ExpireAt: expiresAt,
		TrafficSize: input.TrafficSize, TrafficType: input.TrafficType, Description: input.Description, Config: input.Config,
	})
}

func (u *Usecase) Create(ctx context.Context, account *models.OcservUser) (*models.OcservUser, error) {
	created, err := u.Repository.Create(ctx, account)
	if err != nil {
		return nil, err
	}
	if err := u.accounts.Create(created.Group, created.Username, created.Password, created.Config); err != nil {
		_, _ = u.Repository.Delete(ctx, created.UID)
		return nil, err
	}
	if err := u.accounts.CreateCertificate(created.Username, created.Password); err != nil {
		_, _ = u.accounts.Delete(created.Username)
		_, _ = u.Repository.Delete(ctx, created.UID)
		return nil, err
	}
	u.reload()
	u.applyCertificateStatus(created)
	return created, nil
}

func (u *Usecase) UpdateUser(ctx context.Context, uid string, input UpdateOcservUserData) (*models.OcservUser, error) {
	if uid == "" {
		return nil, errors.New("user id is required")
	}
	user, err := u.GetByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	if input.Group != nil {
		user.Group = *input.Group
	}
	if input.Password != nil {
		user.Password = *input.Password
	}
	if input.Description != nil {
		user.Description = *input.Description
	}
	if input.TrafficSize != nil {
		user.TrafficSize = *input.TrafficSize
	}
	if input.TrafficType != nil && slices.Contains([]string{models.Free, models.MonthlyTransmit, models.MonthlyReceive, models.MonthlyRxTx, models.TotallyTransmit, models.TotallyReceive, models.TotallyRxTx}, *input.TrafficType) {
		user.TrafficType = *input.TrafficType
	}
	if input.Config != nil {
		user.Config = input.Config
	}
	if input.Unlimited {
		user.ExpireAt = nil
	} else if input.ExpireAt != nil {
		if parsed, err := time.Parse("2006-01-02", *input.ExpireAt); err == nil {
			user.ExpireAt = &parsed
		}
	}
	return u.Update(ctx, user)
}

func (u *Usecase) Update(ctx context.Context, account *models.OcservUser) (*models.OcservUser, error) {
	previous, err := u.Repository.GetByUID(ctx, account.UID)
	if err != nil {
		return nil, err
	}
	updated, err := u.Repository.Update(ctx, account)
	if err != nil {
		return nil, err
	}
	if err := u.accounts.Create(updated.Group, updated.Username, updated.Password, updated.Config); err != nil {
		_, _ = u.Repository.Update(ctx, previous)
		return nil, err
	}
	u.reload()
	u.applyCertificateStatus(updated)
	return updated, nil
}

func (u *Usecase) DeleteUser(ctx context.Context, uid string) error {
	if uid == "" {
		return errors.New("user id is required")
	}
	account, err := u.Repository.Delete(ctx, uid)
	if err != nil {
		return err
	}
	if _, err := u.accounts.Delete(account.Username); err != nil {
		_, _ = u.Repository.Create(ctx, account)
		return err
	}
	u.reload()
	_, _ = u.occtl.Terminate(account.Username)
	return nil
}

func (u *Usecase) LockUser(ctx context.Context, uid string) error {
	if uid == "" {
		return errors.New("user id is required")
	}
	account, err := u.Repository.GetByUID(ctx, uid)
	if err != nil {
		return err
	}
	if err := u.Repository.Lock(ctx, uid); err != nil {
		return err
	}
	if _, err := u.accounts.Lock(account.Username); err != nil {
		_ = u.Repository.UnLock(ctx, uid)
		return err
	}
	if _, err := u.occtl.Disconnect(account.Username); err != nil {
		logger.Error("failed to disconnect ocserv user error: %v", err)
	}
	return nil
}

func (u *Usecase) UnlockUser(ctx context.Context, uid string) error {
	if uid == "" {
		return errors.New("user id is required")
	}
	account, err := u.Repository.GetByUID(ctx, uid)
	if err != nil {
		return err
	}
	if err := u.Repository.UnLock(ctx, uid); err != nil {
		return err
	}
	if _, err := u.accounts.UnLock(account.Username); err != nil {
		_ = u.Repository.Lock(ctx, uid)
		return err
	}
	return nil
}

func (u *Usecase) Statistics(ctx context.Context, uid string, input StatisticsData) (*StatisticsResponse, error) {
	if uid == "" {
		return nil, errors.New("user id is required")
	}
	start, end, err := parseDateRange(input.DateStart, input.DateEnd)
	if err != nil {
		return nil, err
	}
	group, groupCtx := errgroup.WithContext(ctx)
	var result StatisticsResponse
	group.Go(func() error {
		value, err := u.Repository.UserStatistics(groupCtx, uid, start, end)
		result.Statistics = value
		return err
	})
	group.Go(func() error {
		value, err := u.reports.TotalBandWidthUser(groupCtx, uid)
		result.TotalBandwidths = value
		return err
	})
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return &result, nil
}

func (u *Usecase) ListOcpasswd(ctx context.Context, pagination *request.Pagination) (*OcpasswdResult, error) {
	accounts, _, err := u.accounts.Ocpasswd(ctx)
	if err != nil {
		return nil, err
	}
	if len(*accounts) == 0 {
		return &OcpasswdResult{Users: []user.Ocpasswd{}}, nil
	}
	usernames := make([]string, 0, len(*accounts))
	for _, account := range *accounts {
		usernames = append(usernames, account.Username)
	}
	existing, err := u.Repository.ExistingUsernames(ctx, usernames)
	if err != nil {
		return nil, err
	}
	existingSet := make(map[string]struct{}, len(existing))
	for _, username := range existing {
		existingSet[username] = struct{}{}
	}
	unsynced := make([]user.Ocpasswd, 0, len(*accounts))
	for _, account := range *accounts {
		if _, found := existingSet[account.Username]; !found {
			unsynced = append(unsynced, account)
		}
	}
	total := len(unsynced)
	start := (pagination.Page - 1) * pagination.PageSize
	if start >= total {
		return &OcpasswdResult{Users: []user.Ocpasswd{}, Total: total}, nil
	}
	end := min(start+pagination.PageSize, total)
	return &OcpasswdResult{Users: unsynced[start:end], Total: total}, nil
}

func (u *Usecase) SyncOcpasswd(ctx context.Context, owner string, input SyncOcpasswdRequest) ([]string, error) {
	if owner == "" {
		return nil, errors.New("admin or staff username not found")
	}
	expiresAt := time.Now().AddDate(0, 0, 30)
	if input.ExpireAt != nil && *input.ExpireAt != "" {
		if parsed, err := time.Parse("2006-01-02", *input.ExpireAt); err == nil {
			expiresAt = parsed
		}
	}
	if input.TrafficType == nil {
		return nil, errors.New("traffic_type is required")
	}
	if input.TrafficSize == nil {
		return nil, errors.New("traffic_size is required")
	}
	size := *input.TrafficSize
	if *input.TrafficType == models.Free {
		size = 0
	}
	users := make([]models.OcservUser, 0, len(input.Users))
	for _, item := range input.Users {
		users = append(users, models.OcservUser{Username: item.Username, Password: "Secret-Ocpasswd", Group: item.Group, Owner: owner, ExpireAt: &expiresAt, TrafficSize: size, TrafficType: *input.TrafficType, Config: input.Config})
	}
	if len(users) == 0 {
		return nil, errors.New("no users found")
	}
	synced, err := u.Repository.OcpasswdSyncToDB(ctx, users)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(synced))
	reload := false
	for _, item := range synced {
		names = append(names, item.Username)
		if item.Config != nil {
			reload = true
		}
	}
	if reload {
		u.reload()
	}
	return names, nil
}

func (u *Usecase) Activate(ctx context.Context, uid string, input ActivateUserData) error {
	if uid == "" {
		return errors.New("user id is required")
	}
	var expiresAt *time.Time
	if input.ExpireAt != nil {
		if parsed, err := time.Parse("2006-01-02", *input.ExpireAt); err == nil {
			expiresAt = &parsed
		}
	}
	account, err := u.Repository.GetByUID(ctx, uid)
	if err != nil {
		return err
	}
	output, err := u.occtl.Terminate(account.Username)
	if err != nil && !isNoActiveOcctlUserError(output, err) {
		return fmt.Errorf("failed to terminate ocserv user %q: %s: %w", account.Username, strings.TrimSpace(output), err)
	}
	output, err = u.accounts.UnLock(account.Username)
	if err != nil && !isAlreadyUnlockedOcpasswdError(output, err) {
		return fmt.Errorf("failed to unlock ocserv user %q: %s: %w", account.Username, strings.TrimSpace(output), err)
	}
	return u.Repository.RestoreExpired(ctx, uid, expiresAt)
}

func (u *Usecase) CreateUserCertificate(ctx context.Context, uid string) error {
	if uid == "" {
		return errors.New("user id is required")
	}
	account, err := u.Repository.GetByUID(ctx, uid)
	if err != nil {
		return err
	}
	return u.accounts.CreateCertificate(account.Username, account.Password)
}

func (u *Usecase) UserCertificate(ctx context.Context, uid string) (string, string, error) {
	if uid == "" {
		return "", "", errors.New("user id is required")
	}
	account, err := u.Repository.GetByUID(ctx, uid)
	if err != nil {
		return "", "", err
	}
	path, err := u.accounts.CertificatePath(account.Username)
	return account.Username, path, err
}

func (u *Usecase) SessionLogs(ctx context.Context, uid string, pagination *request.Pagination, input SessionLogsData) (*SessionLogsResult, error) {
	if uid == "" {
		return nil, errors.New("user id is required")
	}
	start, end, err := parseDateRange(input.DateStart, input.DateEnd)
	if err != nil {
		return nil, err
	}
	user, err := u.GetByUID(ctx, uid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	logs, total, err := u.Repository.UserSessionLogs(ctx, pagination, user.Username, start, end)
	if err != nil {
		return nil, err
	}
	return &SessionLogsResult{Logs: logs, Total: total}, nil
}

func (u *Usecase) DisconnectUser(username string) error {
	return u.sessionAction(username, "disconnect", u.occtl.Disconnect)
}
func (u *Usecase) DisconnectSession(id string) error {
	return u.sessionAction(id, "disconnect", u.occtl.DisconnectSession)
}
func (u *Usecase) TerminateUser(username string) error {
	return u.sessionAction(username, "terminate", u.occtl.Terminate)
}
func (u *Usecase) TerminateSession(id string) error {
	return u.sessionAction(id, "terminate", u.occtl.TerminateSession)
}

func (u *Usecase) sessionAction(value, action string, invoke func(string) (string, error)) error {
	if value == "" {
		return errors.New("user id is required")
	}
	_, err := invoke(value)
	if err != nil && !strings.Contains(err.Error(), "could not "+action+" user") {
		return err
	}
	return nil
}

func parseDateRange(startValue, endValue string) (*time.Time, *time.Time, error) {
	var start, end *time.Time
	if startValue != "" {
		parsed, err := time.Parse("2006-01-02", startValue)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid date_start: %w", err)
		}
		start = &parsed
	}
	if endValue != "" {
		parsed, err := time.Parse("2006-01-02", endValue)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid date_end: %w", err)
		}
		parsed = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		end = &parsed
	}
	return start, end, nil
}

func (u *Usecase) applyCertificateStatus(account *models.OcservUser) {
	status := u.accounts.CertificateStatus(account.Username)
	account.CertificateEnabled = status.Enabled
	account.CertificateAvailable = status.Available
}

func (u *Usecase) reload() {
	if _, err := u.occtl.Reload(); err != nil {
		logger.Warn("Failed to reload ocserv configuration: %v", err)
	}
}

func isAlreadyUnlockedOcpasswdError(output string, err error) bool {
	text := strings.ToLower(strings.TrimSpace(output + " " + err.Error()))
	return strings.Contains(text, "not locked") || strings.Contains(text, "not disabled") ||
		strings.Contains(text, "already unlocked") || strings.Contains(text, "already enabled")
}

func isNoActiveOcctlUserError(output string, err error) bool {
	text := strings.ToLower(strings.TrimSpace(output + " " + err.Error()))
	return strings.Contains(text, "could not terminate user") || strings.Contains(text, "could not disconnect user") ||
		strings.Contains(text, "not found") || strings.Contains(text, "no such user")
}
