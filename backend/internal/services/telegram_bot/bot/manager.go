package bot

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/telegram_bot/auth"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/telegram_bot/conversation"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/telegram_bot/conversation/session"
)

const settingsPollInterval = 30 * time.Second

// Manager owns the lifecycle of the underlying telegram-bot-api client and
// reacts to changes in the persisted TelegramSettings (token, enabled flag).
// It exposes Send so that other goroutines (notifier, API service) can post
// messages through the active bot.
type Manager struct {
	mu           sync.RWMutex
	api          *tgbotapi.BotAPI
	currentTok   string
	enabled      bool
	repo         SettingsRepository
	conversation conversation.Repository
	sessions     *session.Store
	verifier     *auth.Verifier
	receiptsDir  string

	stopUpdates context.CancelFunc
	router      *Router
	consumeWG   sync.WaitGroup
}

type SettingsRepository interface {
	Settings(ctx context.Context) (*models.TelegramSettings, error)
	SetBotUsername(ctx context.Context, username string) error
}

func NewManager(receiptsDir string, repo SettingsRepository, conversations conversation.Repository, sessions *session.Store, verifier *auth.Verifier) *Manager {
	return &Manager{
		repo:         repo,
		conversation: conversations,
		sessions:     sessions,
		verifier:     verifier,
		receiptsDir:  receiptsDir,
	}
}

// Run watches the database for token/enabled changes and (re)starts the
// long-poll loop accordingly. The loop exits when ctx is cancelled.
func (m *Manager) Run(ctx context.Context) error {
	ticker := time.NewTicker(settingsPollInterval)
	defer ticker.Stop()
	defer m.Stop()

	fatalErrors := make(chan error, 1)
	if err := m.refresh(ctx, fatalErrors); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-fatalErrors:
			return err
		case <-ticker.C:
			if err := m.refresh(ctx, fatalErrors); err != nil {
				return err
			}
		}
	}
}

// Stop tears down the active updates loop. Safe to call multiple times.
func (m *Manager) Stop() {
	m.mu.Lock()
	if m.stopUpdates != nil {
		m.stopUpdates()
		m.stopUpdates = nil
	}
	m.mu.Unlock()
	m.consumeWG.Wait()
}

// API returns the active bot API or nil if the bot is currently disabled.
func (m *Manager) API() *tgbotapi.BotAPI {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.api
}

// Send posts an HTML message via the active bot. Returns nil silently if the
// bot is disabled — callers should treat sends as best-effort notifications.
// HTML mode matches the rest of the codebase (i18n catalog uses <b>, <code>).
func (m *Manager) Send(chatID int64, text string) error {
	api := m.API()
	if api == nil {
		return nil
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	_, err := api.Send(msg)
	return err
}

// SendPlain bypasses parse_mode for messages that may contain unescaped user
// content (e.g. raw admin notifications with arbitrary text).
func (m *Manager) SendPlain(chatID int64, text string) error {
	api := m.API()
	if api == nil {
		return nil
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true
	_, err := api.Send(msg)
	return err
}

// Repo exposes the underlying repository so collaborators (notifier) can read
// settings and accounts without re-instantiating their own connection.
func (m *Manager) Repo() conversation.Repository {
	return m.conversation
}

func (m *Manager) refresh(ctx context.Context, fatalErrors chan<- error) error {
	settings, err := m.repo.Settings(ctx)
	if err != nil {
		return fmt.Errorf("telegram bot settings: %w", err)
	}

	desiredToken := settings.BotToken
	desiredEnabled := settings.Enabled

	m.mu.Lock()
	currentTok := m.currentTok
	currentEnabled := m.enabled
	m.mu.Unlock()

	tokenChanged := currentTok != desiredToken
	enabledChanged := currentEnabled != desiredEnabled

	if !tokenChanged && !enabledChanged {
		return nil
	}

	if tokenChanged || !desiredEnabled {
		m.Stop()
		m.mu.Lock()
		m.api = nil
		m.currentTok = ""
		m.enabled = false
		m.router = nil
		m.mu.Unlock()
	}

	if !desiredEnabled || desiredToken == "" {
		logger.Info("telegram_bot: idle (enabled=%v token_set=%v)", desiredEnabled, desiredToken != "")
		return nil
	}

	api, err := tgbotapi.NewBotAPI(desiredToken)
	if err != nil {
		return fmt.Errorf("initialize telegram bot API: %w", err)
	}

	if api.Self.UserName != "" && api.Self.UserName != settings.BotUsername {
		_ = m.repo.SetBotUsername(ctx, api.Self.UserName)
	}

	// Push the localized command list, descriptions and menu button to
	// BotFather. This is what makes /start, /help, /settings, /language
	// appear in the in-app command picker on every Telegram client.
	applyBotMetadata(api)

	router := NewRouter(m, api)

	updatesCtx, cancel := context.WithCancel(ctx)
	m.mu.Lock()
	m.api = api
	m.currentTok = desiredToken
	m.enabled = true
	m.stopUpdates = cancel
	m.router = router
	m.mu.Unlock()

	m.consumeWG.Add(1)
	go func() {
		defer m.consumeWG.Done()
		if err := m.consume(updatesCtx, api, router); err != nil {
			select {
			case fatalErrors <- err:
			default:
			}
		}
	}()

	logger.Info("telegram_bot: connected as @%s", api.Self.UserName)
	return nil
}

func (m *Manager) consume(ctx context.Context, api *tgbotapi.BotAPI, router *Router) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := api.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			api.StopReceivingUpdates()
			return nil
		case upd, ok := <-updates:
			if !ok {
				if ctx.Err() != nil {
					return nil
				}
				return errors.New("telegram updates channel closed unexpectedly")
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("telegram_bot panic: %v", r)
					}
				}()
				router.Dispatch(ctx, upd)
			}()
		}
	}
}

// Sessions exposes the in-memory session store for handlers to use.
func (m *Manager) Sessions() *session.Store {
	return m.sessions
}

// Verifier exposes the credentials verifier for handlers.
func (m *Manager) Verifier() *auth.Verifier {
	return m.verifier
}

// ReceiptsDir returns the on-disk directory where receipts should be stored.
func (m *Manager) ReceiptsDir() string {
	return m.receiptsDir
}
