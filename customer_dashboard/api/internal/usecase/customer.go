package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/config"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/internal/repository"
)

type CustomerUseCase interface {
	GetUser(username string) (*models.OcservUser, error)
	IsOnline(username string) (bool, error)
	OnlineSessions(username string) ([]models.OnlineUserSession, error)
	TerminateAllSessions(username string) error
	TerminateSession(username, sessionID string) error
	UpdatePassword(username, oldPassword, newPassword string) error
	CreateCertificate(username string) error
	DownloadCertificate(username string) (string, error)
	UserStatistics(username string, startDate, endDate *time.Time) ([]models.DailyTraffic, error)
	SessionLogs(username string, page, limit int, orderBy, sort string, startDate, endDate *time.Time) ([]models.OcservUserSessionLog, int64, error)
	Summary(username string) (*models.OcservUser, *user.CertificateStatus, []models.DailyTraffic, error)
	DisconnectSessions(username string) error
	CiscoSetup(username string, publicAPIBase string) (*CiscoSetupResult, error)
	DownloadCiscoSetupCertificate(token string) (string, string, error)
}

type customerUseCase struct {
	systemRepo     repository.SystemRepository
	ocservUserRepo repository.OcservUserRepository
	occtlRepo      repository.OcctlRepository
	ocservUser     user.OcservUserInterface
}

type CiscoSetupResult struct {
	CertificateImportURI string
	ConnectionCreateURI  string
	CertificatePassword  string
	ConnectionName       string
	ServerAddress        string
	ServerPort           int
	ExpiresAt            time.Time
}

func NewCustomerUseCase(
	systemRepo repository.SystemRepository,
	ocservUserRepo repository.OcservUserRepository,
	occtlRepo repository.OcctlRepository,
	ocservUser user.OcservUserInterface,
) CustomerUseCase {
	return &customerUseCase{
		systemRepo:     systemRepo,
		ocservUserRepo: ocservUserRepo,
		occtlRepo:      occtlRepo,
		ocservUser:     ocservUser,
	}
}

func (uc *customerUseCase) GetUser(username string) (*models.OcservUser, error) {
	user, err := uc.ocservUserRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if user.IsLocked {
		return nil, errors.New("8001")
	}
	return user, nil
}

func (uc *customerUseCase) IsOnline(username string) (bool, error) {
	sessions, err := uc.occtlRepo.OnlineSessions()
	if err != nil {
		return false, err
	}

	for _, session := range sessions {
		if session.Username == username {
			return true, nil
		}
	}

	return false, nil
}

func (uc *customerUseCase) OnlineSessions(username string) ([]models.OnlineUserSession, error) {
	allSessions, err := uc.occtlRepo.OnlineSessions()
	if err != nil {
		return nil, err
	}

	var userSessions []models.OnlineUserSession
	for _, session := range allSessions {
		if session.Username == username {
			userSessions = append(userSessions, session)
		}
	}

	return userSessions, nil
}

func (uc *customerUseCase) TerminateAllSessions(username string) error {
	_, err := uc.occtlRepo.Terminate(username)
	return err
}

func (uc *customerUseCase) TerminateSession(username, sessionID string) error {
	allSessions, err := uc.occtlRepo.OnlineSessions()
	if err != nil {
		return err
	}

	found := false
	for _, session := range allSessions {
		if session.ID == sessionID && session.Username == username {
			found = true
			break
		}
	}
	if !found {
		return errors.New("8003")
	}

	_, err = uc.occtlRepo.TerminateSession(sessionID)
	return err
}

func (uc *customerUseCase) UpdatePassword(username, oldPassword, newPassword string) error {
	user, err := uc.ocservUserRepo.FindByUsername(username)
	if err != nil {
		return err
	}
	if user.Password != oldPassword {
		return errors.New("8004")
	}

	return uc.ocservUserRepo.UpdatePassword(username, newPassword)
}

func (uc *customerUseCase) CreateCertificate(username string) error {
	path, err := uc.ocservUserRepo.CertificatePath(username)
	if err == nil && path != "" {
		return nil
	}

	if err := uc.ocservUserRepo.CreateCertificate(username); err != nil {
		return errors.New("8002")
	}
	return nil
}

func (uc *customerUseCase) DownloadCertificate(username string) (string, error) {
	return uc.ocservUserRepo.CertificatePath(username)
}

func (uc *customerUseCase) UserStatistics(username string, startDate, endDate *time.Time) ([]models.DailyTraffic, error) {
	user, err := uc.ocservUserRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if user.IsLocked {
		return nil, errors.New("8001")
	}

	return uc.ocservUserRepo.UserStatistics(user.ID, startDate, endDate)
}

func (uc *customerUseCase) SessionLogs(username string, page, limit int, orderBy, sort string, startDate, endDate *time.Time) ([]models.OcservUserSessionLog, int64, error) {
	return uc.ocservUserRepo.UserSessionLogs(username, page, limit, orderBy, sort, startDate, endDate)
}

func (uc *customerUseCase) Summary(username string) (*models.OcservUser, *user.CertificateStatus, []models.DailyTraffic, error) {
	user, err := uc.ocservUserRepo.FindByUsername(username)
	if err != nil {
		return nil, nil, nil, err
	}

	dateEnd := time.Now()
	firstOfThisMonth := time.Date(dateEnd.Year(), dateEnd.Month(), 1, 0, 0, 0, 0, dateEnd.Location())
	dateStart := firstOfThisMonth.AddDate(0, -1, 0)

	stats, err := uc.ocservUserRepo.UserStatistics(user.ID, &dateStart, &dateEnd)
	if err != nil {
		return nil, nil, nil, err
	}

	certStatus := uc.ocservUser.CertificateStatus(username)

	return user, &certStatus, stats, nil
}

func (uc *customerUseCase) DisconnectSessions(username string) error {
	_, _ = uc.occtlRepo.Disconnect(username)
	return nil
}

const ciscoSetupCertificateTokenTTL = 10 * time.Minute

func (uc *customerUseCase) CiscoSetup(username string, publicAPIBase string) (*CiscoSetupResult, error) {
	user, err := uc.ocservUserRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}

	systemConfig, err := uc.systemRepo.Get()
	if err != nil {
		return nil, err
	}

	connectionName := normalizeProfileConnectionName(systemConfig.ClientProfileConnectionName)
	serverAddress := normalizeProfileServerAddress(systemConfig.ClientProfileServerAddress)
	serverPort := normalizeProfileServerPort(systemConfig.ClientProfileServerPort)

	expiresAt := time.Now().Add(ciscoSetupCertificateTokenTTL)
	token, err := createCiscoSetupCertificateToken(username, expiresAt)
	if err != nil {
		return nil, err
	}

	certificateURL := publicAPIBase + "/api/v1/customer/setup/cisco/certificate/" + urlPathEscape(token)
	certificateImportURI := buildAnyConnectImportURI(certificateURL)
	connectionCreateURI := buildAnyConnectCreateURI(connectionName, serverAddress, serverPort, username)

	return &CiscoSetupResult{
		CertificateImportURI: certificateImportURI,
		ConnectionCreateURI:  connectionCreateURI,
		CertificatePassword:  user.Password,
		ConnectionName:       connectionName,
		ServerAddress:        serverAddress,
		ServerPort:           serverPort,
		ExpiresAt:            expiresAt,
	}, nil
}

func (uc *customerUseCase) DownloadCiscoSetupCertificate(token string) (string, string, error) {
	username, err := parseCiscoSetupCertificateToken(token)
	if err != nil {
		return "", "", err
	}

	path, err := uc.ocservUserRepo.CertificatePath(username)
	if err != nil {
		if err := uc.ocservUserRepo.CreateCertificate(username); err != nil {
			return "", "", err
		}
		path, err = uc.ocservUserRepo.CertificatePath(username)
		if err != nil {
			return "", "", err
		}
	}

	return path, username, nil
}

// Helper functions for Cisco setup
func normalizeProfileConnectionName(name string) string {
	if name == "" {
		return "Ocserv VPN"
	}
	return name
}

func normalizeProfileServerAddress(address string) string {
	if address == "" {
		return "localhost"
	}
	return address
}

func normalizeProfileServerPort(port int) int {
	if port == 0 {
		return 443
	}
	return port
}

func urlPathEscape(s string) string {
	return url.PathEscape(s)
}

func buildAnyConnectImportURI(certificateURL string) string {
	return fmt.Sprintf("anyconnect://import/certificate?url=%s", url.PathEscape(certificateURL))
}

func buildAnyConnectCreateURI(name, server string, port int, username string) string {
	return fmt.Sprintf("anyconnect://create/profile?name=%s&host=%s&port=%d&username=%s",
		url.PathEscape(name), url.PathEscape(server), port, url.PathEscape(username))
}

func createCiscoSetupCertificateToken(username string, expiresAt time.Time) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", errors.New("username is required")
	}
	if strings.Contains(username, "|") {
		return "", errors.New("username contains invalid characters")
	}

	payload := username + "|" + strconv.FormatInt(expiresAt.Unix(), 10)
	signature, err := signCiscoSetupCertificatePayload(payload)
	if err != nil {
		return "", err
	}
	rawToken := payload + "|" + signature
	return base64.RawURLEncoding.EncodeToString([]byte(rawToken)), nil
}

func parseCiscoSetupCertificateToken(token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("token is required")
	}
	rawToken, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return "", errors.New("invalid token")
	}
	parts := strings.Split(string(rawToken), "|")
	if len(parts) != 3 {
		return "", errors.New("invalid token")
	}
	username := parts[0]
	expiresAtUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", errors.New("invalid token expiry")
	}
	if time.Now().After(time.Unix(expiresAtUnix, 0)) {
		return "", errors.New("token has expired")
	}
	payload := username + "|" + parts[1]
	expectedSignature, err := signCiscoSetupCertificatePayload(payload)
	if err != nil {
		return "", err
	}
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return "", errors.New("invalid token signature")
	}
	return username, nil
}

func signCiscoSetupCertificatePayload(payload string) (string, error) {
	secretKey := strings.TrimSpace(config.AppConfig.SecretKey)
	if secretKey == "" {
		return "", errors.New("secret key is not configured")
	}
	mac := hmac.New(sha256.New, []byte(secretKey))
	if _, err := mac.Write([]byte(payload)); err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
