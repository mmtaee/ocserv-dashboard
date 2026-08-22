package system

import (
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/captcha"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
)

// Repository is the persistence boundary required by this usecase.
type Repository interface {
	repository.SystemRepositoryInterface
}

type UserRepository interface {
	repository.UserRepositoryInterface
}

type CaptchaVerifier interface {
	captcha.GoogleCaptchaInterface
}

type PasswordManager interface {
	crypto.CustomPasswordInterface
}
