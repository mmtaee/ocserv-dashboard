package system

import (
	"net/http"
	"sync"
)

type Usecase struct {
	systems         Repository
	users           UserRepository
	sessions        SessionRepository
	captcha         CaptchaVerifier
	passwords       PasswordManager
	httpClient      *http.Client
	secretKey       string
	currentRelease  string
	telegramEnabled bool
	captchaMu       sync.Mutex
}

func New(systems Repository, users UserRepository, sessions SessionRepository, captcha CaptchaVerifier, passwords PasswordManager, options Options) *Usecase {
	client := options.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: options.ReleaseTimeout}
	}
	return &Usecase{
		systems: systems, users: users, sessions: sessions, captcha: captcha, passwords: passwords, httpClient: client,
		secretKey: options.SecretKey, currentRelease: options.CurrentRelease, telegramEnabled: options.TelegramEnabled,
	}
}
