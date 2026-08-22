package conversation

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/telegram_bot/auth"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/telegram_bot/conversation/session"
)

type Repository interface {
	Settings(ctx context.Context) (*models.TelegramSettings, error)
	AccountsByChatID(ctx context.Context, chatID int64) ([]models.TelegramAccount, error)
	AccountByID(ctx context.Context, id uint) (*models.TelegramAccount, error)
	UpsertAccount(ctx context.Context, chatID int64, telegramUsername, language string, ocservUserID uint) (*models.TelegramAccount, error)
	SetTelegramUsernameForChat(ctx context.Context, chatID int64, telegramUsername string) error
	DeleteAccount(ctx context.Context, id uint) error
	UpdateLanguageForChat(ctx context.Context, chatID int64, language string) error
	OcservUserByID(ctx context.Context, id uint) (*models.OcservUser, error)
	OcservUsersByIDs(ctx context.Context, ids []uint) (map[uint]*models.OcservUser, error)
	ActivePackages(ctx context.Context) ([]models.TelegramPackage, error)
	PackageByID(ctx context.Context, id uint) (*models.TelegramPackage, error)
	PendingByChat(ctx context.Context, chatID int64) (*models.TelegramRequest, error)
	CreateRequest(ctx context.Context, request *models.TelegramRequest) (*models.TelegramRequest, error)
	AttachReceipt(ctx context.Context, id uint, path string) error
	RequestsByStatuses(ctx context.Context, statuses []string, limit int) ([]models.TelegramRequest, error)
	AdminStats(ctx context.Context) (repository.AdminStats, error)
}

type Deps struct {
	API        *tgbotapi.BotAPI
	Repo       Repository
	Sessions   *session.Store
	Verifier   *auth.Verifier
	ReceiptDir string
	BrandName  string
}

type Hub struct {
	deps Deps
}
