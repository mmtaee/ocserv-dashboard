package customer

import (
	"context"
	"errors"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type Usecase struct {
	systems      SystemRepository
	users        OcservUserRepository
	certificates CertificateStore
	occtl        OcctlRepository
	secretKey    string
	now          func() time.Time
}

func New(systems SystemRepository, users OcservUserRepository, occtl OcctlRepository, secretKey string, certificates ...CertificateStore) *Usecase {
	usecase := &Usecase{systems: systems, users: users, occtl: occtl, secretKey: secretKey, now: time.Now}
	if len(certificates) > 0 {
		usecase.certificates = certificates[0]
	}
	return usecase
}

func (u *Usecase) authenticate(ctx context.Context, credentials Credentials) (*models.OcservUser, error) {
	if credentials.Password == "Secret-Ocpasswd" {
		return nil, ErrInvalidCredentials
	}
	user, err := u.users.GetByUsername(ctx, credentials.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.Password != credentials.Password {
		return nil, ErrInvalidCredentials
	}
	if u.certificates != nil {
		status := u.certificates.CertificateStatus(user.Username)
		user.CertificateEnabled = status.Enabled
		user.CertificateAvailable = status.Available
	}
	return user, nil
}

func (u *Usecase) Login(ctx context.Context, credentials Credentials) (*LoginResponse, error) {
	user, err := u.authenticate(ctx, credentials)
	if err != nil {
		return nil, err
	}
	expiresAt := u.now().Add(accessTokenTTL)
	token, err := u.createToken(user.ID, user.Username, expiresAt)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{User: customerFromModel(user), Token: token, ExpiresAt: expiresAt}, nil
}

func (u *Usecase) user(ctx context.Context, username string) (*models.OcservUser, error) {
	user, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if u.certificates != nil {
		status := u.certificates.CertificateStatus(user.Username)
		user.CertificateEnabled = status.Enabled
		user.CertificateAvailable = status.Available
	}
	return user, nil
}

func customerFromModel(user *models.OcservUser) Customer {
	return Customer{Owner: user.Owner.Username, Username: user.Username, IsLocked: user.IsLocked, CertificateEnabled: user.CertificateEnabled, CertificateAvailable: user.CertificateAvailable, ExpiryMode: user.ExpiryMode, ExpireAt: user.ExpireAt, ExpireDaysAfterFirstConnection: user.ExpireDaysAfterFirstConnection, FirstConnectedAt: user.FirstConnectedAt, DeactivatedAt: user.DeactivatedAt, TrafficType: user.TrafficType, TrafficSize: user.TrafficSize, RunningRx: user.RunningRx, RunningTx: user.RunningTx}
}
