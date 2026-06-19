package bot

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/ocserv_telegram_bot/internal/bot/cbdata"
	"github.com/mmtaee/ocserv-dashboard/ocserv_telegram_bot/internal/bot/handlers"
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
	brand := api.Self.FirstName
	if brand == "" {
		brand = api.Self.UserName
	}
	hub := handlers.NewHub(handlers.Deps{
		API:        api,
		Repo:       mgr.Repo(),
		Sessions:   mgr.Sessions(),
		Verifier:   mgr.Verifier(),
		ReceiptDir: mgr.ReceiptsDir(),
		BrandName:  brand,
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
	if upd.Message.Chat == nil {
		return
	}

	chatID := upd.Message.Chat.ID
	if upd.Message.Chat.IsPrivate() {
		r.hub.SyncTelegramUsernameFromAPI(ctx, chatID)
	}
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
	if cq.Message.Chat == nil {
		return
	}
	chatID := cq.Message.Chat.ID
	srcMsgID := cq.Message.MessageID
	if cq.Message.Chat.IsPrivate() {
		r.hub.SyncTelegramUsernameFromAPI(ctx, chatID)
	}
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
	case data == cbdata.MainMenu:
		r.mgr.Sessions().Reset(chatID)
		// "User view" for admins: MainMenu always renders the user menu.
		r.hub.SendUserMenu(ctx, chatID, lang, srcMsgID)

	case data == cbdata.AdminMenu:
		r.mgr.Sessions().Reset(chatID)
		if !r.hub.IsAdmin(ctx, chatID) {
			r.hub.SendUserMenu(ctx, chatID, lang, srcMsgID)
		} else {
			r.hub.SendAdminMenu(ctx, chatID, lang, srcMsgID)
		}

	case data == cbdata.AdminPending:
		if r.hub.IsAdmin(ctx, chatID) {
			r.hub.ShowAdminPending(ctx, chatID, lang, srcMsgID)
		}

	case data == cbdata.AdminReceipts:
		if r.hub.IsAdmin(ctx, chatID) {
			r.hub.ShowAdminReceipts(ctx, chatID, lang, srcMsgID)
		}

	case data == cbdata.AdminStats:
		if r.hub.IsAdmin(ctx, chatID) {
			r.hub.ShowAdminStats(ctx, chatID, lang, srcMsgID)
		}

	case data == cbdata.AddAccount:
		r.hub.StartAddAccount(ctx, chatID, srcMsgID)

	case data == cbdata.MyAccounts:
		r.hub.SendMyAccounts(ctx, chatID, lang, srcMsgID)

	case data == cbdata.NewOrder:
		r.hub.StartNewOrder(ctx, chatID, srcMsgID)

	case data == cbdata.Help:
		r.hub.ShowHelp(ctx, chatID, lang, srcMsgID)

	case data == cbdata.Language:
		r.hub.ShowLanguageMenu(ctx, chatID, lang, srcMsgID)

	case strings.HasPrefix(data, models.TelegramLangCallbackPrefix):
		if code, ok := models.TelegramLangFromCallback(data); ok {
			r.hub.SetLanguage(ctx, chatID, code, srcMsgID)
			toast = "✓"
		}

	case strings.HasPrefix(data, cbdata.AccountDetail):
		r.hub.ShowAccountDetail(ctx, chatID, parseUintSuffix(data, cbdata.AccountDetail), srcMsgID)

	case strings.HasPrefix(data, cbdata.AccountUsage):
		r.hub.SendAccountUsage(ctx, chatID, parseUintSuffix(data, cbdata.AccountUsage), lang, srcMsgID)

	case strings.HasPrefix(data, cbdata.AccountRenew):
		r.hub.StartRenewForAccount(ctx, chatID, parseUintSuffix(data, cbdata.AccountRenew), srcMsgID)

	case strings.HasPrefix(data, cbdata.AccountRemove):
		r.hub.RemoveAccount(ctx, chatID, parseUintSuffix(data, cbdata.AccountRemove), srcMsgID)

	case strings.HasPrefix(data, cbdata.PickPackageNew):
		r.hub.PickedPackageNew(ctx, chatID, parseUintSuffix(data, cbdata.PickPackageNew), srcMsgID)

	case strings.HasPrefix(data, cbdata.PickPackageRenew):
		r.hub.PickedPackageRenew(ctx, chatID, parseUintSuffix(data, cbdata.PickPackageRenew), srcMsgID)

	default:
		logger.Warn("ocserv_telegram_bot: unknown callback data: %s", data)
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
