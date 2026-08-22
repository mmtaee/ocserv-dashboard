package customer

import (
	"context"
	"errors"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

const ciscoSetupCertificateTokenTTL = 10 * time.Minute

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
		return nil, err
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

func customerFromModel(user *models.OcservUser) Customer {
	return Customer{Owner: user.Owner, Username: user.Username, IsLocked: user.IsLocked, CertificateEnabled: user.CertificateEnabled, CertificateAvailable: user.CertificateAvailable, ExpireAt: user.ExpireAt, DeactivatedAt: user.DeactivatedAt, TrafficType: user.TrafficType, TrafficSize: user.TrafficSize, Rx: user.Rx, Tx: user.Tx}
}
