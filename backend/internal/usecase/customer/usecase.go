package customer

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	ocservuser "github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/user"
)

const ciscoSetupCertificateTokenTTL = 10 * time.Minute

var ErrInvalidCredentials = errors.New("invalid username or password")

type Usecase struct {
	systems   SystemRepository
	users     OcservUserRepository
	occtl     OcctlRepository
	secretKey string
	now       func() time.Time
}

func New(systems SystemRepository, users OcservUserRepository, occtl OcctlRepository, secretKey string) *Usecase {
	return &Usecase{systems: systems, users: users, occtl: occtl, secretKey: secretKey, now: time.Now}
}

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
	return &Summary{
		OcservUser: customerFromModel(user),
		Usage:      Usage{DateStart: dateStart, DateEnd: dateEnd, Bandwidths: bandwidths},
	}, nil
}

func (u *Usecase) CertificatePath(ctx context.Context, credentials Credentials) (string, string, error) {
	user, err := u.authenticate(ctx, credentials)
	if err != nil {
		return "", "", err
	}
	path, err := u.users.CertificatePathByUsername(ctx, user.Username)
	return path, user.Username, err
}

func (u *Usecase) CiscoCertificatePath(ctx context.Context, token string) (string, string, error) {
	username, err := u.parseToken(token)
	if err != nil {
		return "", "", err
	}
	user, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}
	path, err := u.users.CertificatePathByUsername(ctx, user.Username)
	if err != nil {
		if err = u.users.CreateCertificate(ctx, user.UID); err != nil {
			return "", "", err
		}
		path, err = u.users.CertificatePathByUsername(ctx, user.Username)
	}
	return path, user.Username, err
}

func (u *Usecase) CiscoSetup(ctx context.Context, credentials Credentials, publicBaseURL string) (*CiscoSetup, error) {
	user, err := u.authenticate(ctx, credentials)
	if err != nil {
		return nil, err
	}
	system, err := u.systems.System(ctx)
	if err != nil {
		return nil, err
	}
	connectionName, err := ocservuser.NormalizeProfileConnectionName(system.ClientProfileConnectionName)
	if err != nil {
		return nil, err
	}
	serverAddress, err := ocservuser.NormalizeProfileServerAddress(system.ClientProfileServerAddress)
	if err != nil {
		return nil, err
	}
	serverPort, err := ocservuser.NormalizeProfileServerPort(system.ClientProfileServerPort)
	if err != nil {
		return nil, err
	}
	expiresAt := u.now().Add(ciscoSetupCertificateTokenTTL)
	token, err := u.createToken(user.Username, expiresAt)
	if err != nil {
		return nil, err
	}
	certificateURI, err := ocservuser.BuildAnyConnectImportURI(publicBaseURL + "/api/customers/setup/cisco/certificate/" + url.PathEscape(token))
	if err != nil {
		return nil, err
	}
	connectionURI, err := ocservuser.BuildAnyConnectCreateURI(connectionName, serverAddress, serverPort, user.Username)
	if err != nil {
		return nil, err
	}
	return &CiscoSetup{CertificateImportURI: certificateURI, ConnectionCreateURI: connectionURI, CertificatePassword: user.Password, ConnectionName: connectionName, ServerAddress: serverAddress, ServerPort: serverPort, ExpiresAt: expiresAt}, nil
}

func (u *Usecase) Disconnect(ctx context.Context, credentials Credentials) error {
	user, err := u.authenticate(ctx, credentials)
	if err != nil {
		return err
	}
	_, err = u.occtl.Disconnect(user.Username)
	return err
}

func (u *Usecase) authenticate(ctx context.Context, credentials Credentials) (*models.OcservUser, error) {
	if credentials.Password == "Secret-Ocpasswd" {
		return nil, ErrInvalidCredentials
	}
	user, err := u.users.GetByUsername(ctx, credentials.Username)
	if err != nil {
		return nil, err
	}
	if user.Password != credentials.Password {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (u *Usecase) createToken(username string, expiresAt time.Time) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" || strings.Contains(username, "|") {
		return "", errors.New("invalid username")
	}
	payload := username + "|" + strconv.FormatInt(expiresAt.Unix(), 10)
	signature, err := u.sign(payload)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString([]byte(payload + "|" + signature)), nil
}

func (u *Usecase) parseToken(token string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil {
		return "", errors.New("invalid token")
	}
	parts := strings.Split(string(raw), "|")
	if len(parts) != 3 {
		return "", errors.New("invalid token")
	}
	expires, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || u.now().After(time.Unix(expires, 0)) {
		return "", errors.New("token has expired")
	}
	expected, err := u.sign(parts[0] + "|" + parts[1])
	if err != nil {
		return "", err
	}
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return "", errors.New("invalid token signature")
	}
	return parts[0], nil
}

func (u *Usecase) sign(payload string) (string, error) {
	if strings.TrimSpace(u.secretKey) == "" {
		return "", errors.New("secret key is not configured")
	}
	mac := hmac.New(sha256.New, []byte(u.secretKey))
	if _, err := mac.Write([]byte(payload)); err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func customerFromModel(user *models.OcservUser) Customer {
	return Customer{Owner: user.Owner, Username: user.Username, IsLocked: user.IsLocked, CertificateEnabled: user.CertificateEnabled, CertificateAvailable: user.CertificateAvailable, ExpireAt: user.ExpireAt, DeactivatedAt: user.DeactivatedAt, TrafficType: user.TrafficType, TrafficSize: user.TrafficSize, Rx: user.Rx, Tx: user.Tx}
}
