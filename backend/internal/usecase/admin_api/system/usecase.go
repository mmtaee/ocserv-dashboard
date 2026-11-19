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
	installDir      string
	sourceDir       string
	dockerMode      bool
	agentNode       bool
	updateStarter   func(string) error
	captchaMu       sync.Mutex
}

func New(systems Repository, users UserRepository, sessions SessionRepository, captcha CaptchaVerifier, passwords PasswordManager, options Options) *Usecase {
	client := options.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: options.ReleaseTimeout}
	}
	if options.InstallDir == "" {
		options.InstallDir = "/opt/ocserv-dashboard"
	}
	if options.SourceDir == "" {
		options.SourceDir = options.InstallDir + "/source"
	}
	return &Usecase{
		systems: systems, users: users, sessions: sessions, captcha: captcha, passwords: passwords, httpClient: client,
		secretKey: options.SecretKey, currentRelease: options.CurrentRelease, telegramEnabled: options.TelegramEnabled,
		installDir: options.InstallDir, sourceDir: options.SourceDir, dockerMode: options.DockerMode, agentNode: options.AgentNode, updateStarter: options.UpdateStarter,
	}
}
