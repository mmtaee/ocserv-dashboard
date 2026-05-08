package bot

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/bot/handlers"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/i18n"
)

// Router is responsible for taking a single Telegram update and dispatching it
// to the appropriate handler based on the user's session state and the type of
// update (text command, callback query, photo).
type Router struct {
	mgr *Manager
	api *tgbotapi.BotAPI
	hub *handlers.Hub
}

func NewRouter(mgr *Manager, api *tgbotapi.BotAPI) *Router {
	hub := handlers.NewHub(handlers.Deps{
		API:        api,
		Repo:       mgr.Repo(),
		Sessions:   mgr.Sessions(),
		Verifier:   mgr.Verifier(),
		ReceiptDir: mgr.ReceiptsDir(),
	})
	return &Router{mgr: mgr, api: api, hub: hub}
}

// Dispatch is the single entry point for incoming Telegram updates.
func (r *Router) Dispatch(ctx context.Context, upd tgbotapi.Update) {
	if upd.CallbackQuery != nil {
		r.handleCallback(ctx, upd.CallbackQuery)
		return
	}
	if upd.Message == nil {
		return
	}

	chatID := upd.Message.Chat.ID
	if upd.Message.Photo != nil && len(upd.Message.Photo) > 0 {
		r.hub.HandlePhoto(ctx, upd.Message)
		return
	}

	text := strings.TrimSpace(upd.Message.Text)
	lang := r.hub.LanguageFor(ctx, chatID)

	if text == "/cancel" || text == "/stop" {
		r.mgr.Sessions().Reset(chatID)
		r.hub.SendMainMenu(ctx, chatID, lang)
		return
	}

	if isCommand(text, "/start") {
		r.hub.HandleStart(ctx, upd.Message)
		return
	}
	if isCommand(text, "/help") {
		_ = r.send(chatID, i18n.T(lang, i18n.HelpText))
		return
	}
	if isCommand(text, "/skip") {
		r.hub.HandleSkip(ctx, upd.Message)
		return
	}
	if isCommand(text, "/language") {
		r.hub.HandleLanguageMenu(ctx, chatID, lang)
		return
	}

	// Stateful flow steps
	if r.hub.HandleStateful(ctx, upd.Message) {
		return
	}

	_ = r.send(chatID, i18n.T(lang, i18n.UnknownCommand))
	r.hub.SendMainMenu(ctx, chatID, lang)
}

func (r *Router) handleCallback(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	if cq == nil || cq.Message == nil {
		return
	}
	chatID := cq.Message.Chat.ID
	data := cq.Data

	defer func() {
		ack := tgbotapi.NewCallback(cq.ID, "")
		_, _ = r.api.Request(ack)
	}()

	lang := r.hub.LanguageFor(ctx, chatID)

	switch {
	case data == cbMainMenu:
		r.mgr.Sessions().Reset(chatID)
		r.hub.SendMainMenu(ctx, chatID, lang)

	case data == cbAddAccount:
		r.hub.StartAddAccount(ctx, chatID)

	case data == cbMyAccounts:
		r.hub.SendMyAccounts(ctx, chatID, lang)

	case data == cbNewOrder:
		r.hub.StartNewOrder(ctx, chatID)

	case data == cbHelp:
		_ = r.send(chatID, i18n.T(lang, i18n.HelpText))

	case data == cbLanguage:
		r.hub.HandleLanguageMenu(ctx, chatID, lang)

	case data == cbLangEN:
		r.hub.SetLanguage(ctx, chatID, models.TelegramLanguageEN)
		r.hub.SendMainMenu(ctx, chatID, models.TelegramLanguageEN)

	case data == cbLangFA:
		r.hub.SetLanguage(ctx, chatID, models.TelegramLanguageFA)
		r.hub.SendMainMenu(ctx, chatID, models.TelegramLanguageFA)

	case strings.HasPrefix(data, cbAccountUsage):
		r.hub.SendAccountUsage(ctx, chatID, parseUintSuffix(data, cbAccountUsage), lang)

	case strings.HasPrefix(data, cbAccountRenew):
		r.hub.StartRenewForAccount(ctx, chatID, parseUintSuffix(data, cbAccountRenew))

	case strings.HasPrefix(data, cbAccountRemove):
		r.hub.RemoveAccount(ctx, chatID, parseUintSuffix(data, cbAccountRemove))

	case strings.HasPrefix(data, cbPickPackageNew):
		r.hub.PickedPackageNew(ctx, chatID, parseUintSuffix(data, cbPickPackageNew))

	case strings.HasPrefix(data, cbPickPackageRenew):
		r.hub.PickedPackageRenew(ctx, chatID, parseUintSuffix(data, cbPickPackageRenew))

	default:
		logger.Warn("telegram_bot: unknown callback data: %s", data)
	}
}

func (r *Router) send(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := r.api.Send(msg)
	return err
}

func isCommand(text, cmd string) bool {
	if text == cmd {
		return true
	}
	return strings.HasPrefix(text, cmd+" ") || strings.HasPrefix(text, cmd+"@")
}

func parseUintSuffix(data, prefix string) uint {
	suffix := strings.TrimPrefix(data, prefix)
	var v uint
	for _, ch := range suffix {
		if ch < '0' || ch > '9' {
			return 0
		}
		v = v*10 + uint(ch-'0')
	}
	return v
}
