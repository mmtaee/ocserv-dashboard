package usecase

import (
	"context"
	"time"

	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
)

type OcservUserUsecaseInterface interface {
	Users(ctx context.Context, pagination *request.Pagination, owner string, q string, filter string, group string, onlineUsers []models.OnlineUserSession) ([]models.OcservUser, int64, error)
	UsersByUsername(ctx context.Context, pagination *request.Pagination, owner string, usernames []string, q string, group string) ([]models.OcservUser, int64, error)
	GetByUID(ctx context.Context, userUID string) (*models.OcservUser, error)
	Create(ctx context.Context, ocservUser *models.OcservUser) (*models.OcservUser, error)
	Update(ctx context.Context, ocservUser *models.OcservUser) (*models.OcservUser, error)
	Delete(ctx context.Context, userID string) (string, error)
	Lock(ctx context.Context, userID string) error
	Unlock(ctx context.Context, userID string) error
	RestoreExpired(ctx context.Context, userID string, expireAt *time.Time) error
	UserStatistics(ctx context.Context, userID string, startDate *time.Time, endDate *time.Time) ([]models.DailyTraffic, error)
	TotalBandwidthUser(ctx context.Context, userID string) (repository.TotalBandwidths, error)
	Ocpasswd(ctx context.Context, pagination *request.Pagination) ([]user.Ocpasswd, int, error)
	OcpasswdSyncToDB(ctx context.Context, users []models.OcservUser) ([]models.OcservUser, error)
	CreateCertificate(ctx context.Context, userID string) error
	CertificatePath(ctx context.Context, userID string) (string, string, error)
	UserSessionLogs(ctx context.Context, pagination *request.Pagination, username string, startDate *time.Time, endDate *time.Time) ([]models.OcservUserSessionLog, int64, error)
	Disconnect(username string) (interface{}, error)
	DisconnectSession(sessionID string) (interface{}, error)
	Terminate(username string) (interface{}, error)
	TerminateSession(sessionID string) (interface{}, error)
	OnlineSessions() ([]models.OnlineUserSession, error)
}

type OcservUserUsecase struct {
	ocservUserRepo  repository.OcservUserRepositoryInterface
	ocservOcctlRepo repository.OcctlRepositoryInterface
	reportRepo      repository.ReportRepositoryInterface
}

func NewOcservUserUsecase(
	ocservUserRepo repository.OcservUserRepositoryInterface,
	ocservOcctlRepo repository.OcctlRepositoryInterface,
	reportRepo repository.ReportRepositoryInterface,
) *OcservUserUsecase {
	return &OcservUserUsecase{
		ocservUserRepo:  ocservUserRepo,
		ocservOcctlRepo: ocservOcctlRepo,
		reportRepo:      reportRepo,
	}
}

func (uc *OcservUserUsecase) Users(ctx context.Context, pagination *request.Pagination, owner string, q string, filter string, group string, onlineUsers []models.OnlineUserSession) ([]models.OcservUser, int64, error) {
	users, total, err := uc.ocservUserRepo.Users(ctx, pagination, owner, q, filter, group)
	if err != nil {
		return nil, 0, err
	}

	// Attach online status
	onlineUsersMap := make(map[string][]models.OnlineUserSession)
	for _, u := range onlineUsers {
		onlineUsersMap[u.Username] = append(onlineUsersMap[u.Username], u)
	}

	for i := range users {
		if sessions, ok := onlineUsersMap[users[i].Username]; ok {
			users[i].IsOnline = true
			users[i].OnlineUserSessions = sessions
		}
	}

	return users, total, nil
}

func (uc *OcservUserUsecase) UsersByUsername(ctx context.Context, pagination *request.Pagination, owner string, usernames []string, q string, group string) ([]models.OcservUser, int64, error) {
	users, total, err := uc.ocservUserRepo.UsersByUsername(ctx, pagination, owner, usernames, q, group)
	if err != nil {
		return nil, 0, err
	}

	// Attach online status
	onlineUsers, err := uc.ocservOcctlRepo.OnlineSessions()
	if err != nil {
		return nil, 0, err
	}
	onlineUsersMap := make(map[string][]models.OnlineUserSession)
	for _, u := range onlineUsers {
		onlineUsersMap[u.Username] = append(onlineUsersMap[u.Username], u)
	}

	for i := range users {
		users[i].IsOnline = true
		users[i].OnlineUserSessions = onlineUsersMap[users[i].Username]
	}

	return users, total, nil
}

func (uc *OcservUserUsecase) GetByUID(ctx context.Context, userUID string) (*models.OcservUser, error) {
	return uc.ocservUserRepo.GetByUID(ctx, userUID)
}

func (uc *OcservUserUsecase) Create(ctx context.Context, ocservUser *models.OcservUser) (*models.OcservUser, error) {
	return uc.ocservUserRepo.Create(ctx, ocservUser)
}

func (uc *OcservUserUsecase) Update(ctx context.Context, ocservUser *models.OcservUser) (*models.OcservUser, error) {
	return uc.ocservUserRepo.Update(ctx, ocservUser)
}

func (uc *OcservUserUsecase) Delete(ctx context.Context, userID string) (string, error) {
	username, err := uc.ocservUserRepo.Delete(ctx, userID)
	if err != nil {
		return "", err
	}

	go func() {
		_, _ = uc.ocservOcctlRepo.Terminate(username)
	}()

	return username, nil
}

func (uc *OcservUserUsecase) Lock(ctx context.Context, userID string) error {
	err := uc.ocservUserRepo.Lock(ctx, userID)
	if err != nil {
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		u, err := uc.ocservUserRepo.GetByUID(ctx, userID)
		if err != nil {
			logger.Error("failed to fetch ocserv user error: %v", err)
		}
		_, err = uc.ocservOcctlRepo.Disconnect(u.Username)
		if err != nil {
			logger.Error("failed to disconnect ocserv user error: %v", err)
		}
	}()

	return nil
}

func (uc *OcservUserUsecase) Unlock(ctx context.Context, userID string) error {
	return uc.ocservUserRepo.UnLock(ctx, userID)
}

func (uc *OcservUserUsecase) RestoreExpired(ctx context.Context, userID string, expireAt *time.Time) error {
	return uc.ocservUserRepo.RestoreExpired(ctx, userID, expireAt)
}

func (uc *OcservUserUsecase) UserStatistics(ctx context.Context, userID string, startDate *time.Time, endDate *time.Time) ([]models.DailyTraffic, error) {
	return uc.ocservUserRepo.UserStatistics(ctx, userID, startDate, endDate)
}

func (uc *OcservUserUsecase) TotalBandwidthUser(ctx context.Context, userID string) (repository.TotalBandwidths, error) {
	return uc.reportRepo.TotalBandWidthUser(ctx, userID)
}

func (uc *OcservUserUsecase) Ocpasswd(ctx context.Context, pagination *request.Pagination) ([]user.Ocpasswd, int, error) {
	return uc.ocservUserRepo.Ocpasswd(ctx, pagination)
}

func (uc *OcservUserUsecase) OcpasswdSyncToDB(ctx context.Context, users []models.OcservUser) ([]models.OcservUser, error) {
	return uc.ocservUserRepo.OcpasswdSyncToDB(ctx, users)
}

func (uc *OcservUserUsecase) CreateCertificate(ctx context.Context, userID string) error {
	return uc.ocservUserRepo.CreateCertificate(ctx, userID)
}

func (uc *OcservUserUsecase) CertificatePath(ctx context.Context, userID string) (string, string, error) {
	return uc.ocservUserRepo.CertificatePath(ctx, userID)
}

func (uc *OcservUserUsecase) UserSessionLogs(ctx context.Context, pagination request.Pagination, username string, startDate *time.Time, endDate *time.Time) ([]models.OcservUserSessionLog, int64, error) {
	return uc.ocservUserRepo.UserSessionLogs(ctx, &pagination, username, startDate, endDate)
}

func (uc *OcservUserUsecase) Disconnect(username string) (interface{}, error) {
	return uc.ocservOcctlRepo.Disconnect(username)
}

func (uc *OcservUserUsecase) DisconnectSession(sessionID string) (interface{}, error) {
	return uc.ocservOcctlRepo.DisconnectSession(sessionID)
}

func (uc *OcservUserUsecase) Terminate(username string) (interface{}, error) {
	return uc.ocservOcctlRepo.Terminate(username)
}

func (uc *OcservUserUsecase) TerminateSession(sessionID string) (interface{}, error) {
	return uc.ocservOcctlRepo.TerminateSession(sessionID)
}

func (uc *OcservUserUsecase) OnlineSessions() ([]models.OnlineUserSession, error) {
	return uc.ocservOcctlRepo.OnlineSessions()
}
