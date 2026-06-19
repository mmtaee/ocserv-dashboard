package usecase

import (
	"errors"
	"time"

	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/pkg/auth"
)

type AuthUseCase interface {
	Login(username, password string) (string, time.Time, error)
}

type authUseCase struct {
	ocservUserRepo repository.OcservUserRepository
}

func NewAuthUseCase(ocservUserRepo repository.OcservUserRepository) AuthUseCase {
	return &authUseCase{
		ocservUserRepo: ocservUserRepo,
	}
}

func (uc *authUseCase) Login(username, password string) (string, time.Time, error) {
	user, err := uc.ocservUserRepo.FindByUsername(username)
	if err != nil {
		return "", time.Time{}, err
	}

	if user.IsLocked {
		return "", time.Time{}, errors.New("8001")
	}

	if user.Password != password {
		return "", time.Time{}, errors.New("invalid username or password")
	}

	token, err := auth.CreateCustomerToken(username)
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(time.Hour * 1)

	return token, expiresAt, nil
}
