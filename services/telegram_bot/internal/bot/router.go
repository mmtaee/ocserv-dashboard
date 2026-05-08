package bot

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/bot/handlers"
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
		r.hub.SendMainMenu(ctx, chatID, lang, 0)
		return
	}

	if isCommand(text, "/start") {
		r.hub.HandleStart(ctx, upd.Message)
		return
	}
	if isCommand(text, "/help") {
		r.hub.ShowHelp(ctx, chatID, lang, 0)
		return
	}
	if isCommand(text, "/skip") {
		r.hub.HandleSkip(ctx, upd.Message)
		return
	}
	if isCommand(text, "/language") || isCommand(text, "/settings") {
		r.hub.ShowLanguageMenu(ctx, chatID, lang, 0)
		return
	}

	// Stateful flow steps
	if r.hub.HandleStateful(ctx, upd.Message) {
		return
	}

	r.hub.SendMainMenu(ctx, chatID, lang, 0)
}

func (r *Router) handleCallback(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	if cq == nil || cq.Message == nil {
		return
	}
	chatID := cq.Message.Chat.ID
	srcMsgID := cq.Message.MessageID
	data := cq.Data

	// toast is shown as a short, non-modal notification on the user's
	// device after the callback is processed. Telegram requires us to
	// answer every callback query — see Status Alerts in the Bot docs.
	toast := ""
	defer func() {
		ack := tgbotapi.NewCallback(cq.ID, toast)
		_, _ = r.api.Request(ack)
	}()

	lang := r.hub.LanguageFor(ctx, chatID)

	switch {
	case data == cbMainMenu:
		r.mgr.Sessions().Reset(chatID)
		r.hub.SendMainMenu(ctx, chatID, lang, srcMsgID)

	case data == cbAddAccount:
		r.hub.StartAddAccount(ctx, chatID, srcMsgID)

	case data == cbMyAccounts:
		r.hub.SendMyAccounts(ctx, chatID, lang, srcMsgID)

	case data == cbNewOrder:
		r.hub.StartNewOrder(ctx, chatID, srcMsgID)

	case data == cbHelp:
		r.hub.ShowHelp(ctx, chatID, lang, srcMsgID)

	case data == cbLanguage:
		r.hub.ShowLanguageMenu(ctx, chatID, lang, srcMsgID)

	case data == cbLangEN:
		r.hub.SetLanguage(ctx, chatID, models.TelegramLanguageEN, srcMsgID)
		toast = "✓ Language updated"

	case data == cbLangFA:
		r.hub.SetLanguage(ctx, chatID, models.TelegramLanguageFA, srcMsgID)
		toast = "✓ زبان تغییر کرد"

	case strings.HasPrefix(data, cbAccountDetail):
		r.hub.ShowAccountDetail(ctx, chatID, parseUintSuffix(data, cbAccountDetail), srcMsgID)

	case strings.HasPrefix(data, cbAccountUsage):
		r.hub.SendAccountUsage(ctx, chatID, parseUintSuffix(data, cbAccountUsage), lang, srcMsgID)

	case strings.HasPrefix(data, cbAccountRenew):
		r.hub.StartRenewForAccount(ctx, chatID, parseUintSuffix(data, cbAccountRenew), srcMsgID)

	case strings.HasPrefix(data, cbAccountRemove):
		r.hub.RemoveAccount(ctx, chatID, parseUintSuffix(data, cbAccountRemove), srcMsgID)

	case strings.HasPrefix(data, cbPickPackageNew):
		r.hub.PickedPackageNew(ctx, chatID, parseUintSuffix(data, cbPickPackageNew), srcMsgID)

	case strings.HasPrefix(data, cbPickPackageRenew):
		r.hub.PickedPackageRenew(ctx, chatID, parseUintSuffix(data, cbPickPackageRenew), srcMsgID)

	default:
		logger.Warn("telegram_bot: unknown callback data: %s", data)
	}
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
