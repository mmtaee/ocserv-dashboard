package customer

import (
	"context"
	"errors"
)

var ErrCertificateStoreUnavailable = errors.New("certificate store unavailable")

func (u *Usecase) CertificatePath(ctx context.Context, username string) (string, string, error) {
	user, err := u.user(ctx, username)
	if err != nil {
		return "", "", err
	}
	if u.certificates == nil {
		return "", "", ErrCertificateStoreUnavailable
	}
	path, err := u.certificates.CertificatePath(user.Username)
	return path, user.Username, err
}

func (u *Usecase) CiscoCertificatePath(ctx context.Context, username string) (string, string, error) {
	user, err := u.user(ctx, username)
	if err != nil {
		return "", "", err
	}
	if u.certificates == nil {
		return "", "", ErrCertificateStoreUnavailable
	}
	path, err := u.certificates.CertificatePath(user.Username)
	if err != nil {
		if err = u.certificates.CreateCertificate(user.Username, user.Password); err != nil {
			return "", "", err
		}
		path, err = u.certificates.CertificatePath(user.Username)
	}
	return path, user.Username, err
}
