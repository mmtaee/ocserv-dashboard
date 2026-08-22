package customer

import (
	"context"
	"errors"
)

var ErrCertificateStoreUnavailable = errors.New("certificate store unavailable")

func (u *Usecase) CertificatePath(ctx context.Context, credentials Credentials) (string, string, error) {
	user, err := u.authenticate(ctx, credentials)
	if err != nil {
		return "", "", err
	}
	if u.certificates == nil {
		return "", "", ErrCertificateStoreUnavailable
	}
	path, err := u.certificates.CertificatePath(user.Username)
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
