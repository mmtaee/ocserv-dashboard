package handlers

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/auth"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/i18n"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/telegram_bot/internal/session"
)

// All bot messages are sent with parse_mode=HTML so the i18n catalog can use
// <b>, <code>, <i> etc. User-supplied values (usernames, notes) MUST be
// passed through htmlEscape before being interpolated.
const parseModeHTML = "HTML"

func htmlEscape(s string) string {
	return html.EscapeString(s)
}

const (
	cbMainMenu         = "menu:main"
	cbAddAccount       = "menu:add"
	cbMyAccounts       = "menu:list"
	cbNewOrder         = "menu:order"
	cbHelp             = "menu:help"
	cbLanguage         = "menu:lang"
	cbLangEN           = "lang:en"
	cbLangFA           = "lang:fa"
	cbAccountDetail    = "acc:detail:"
	cbAccountUsage     = "acc:usage:"
	cbAccountRenew     = "acc:renew:"
	cbAccountRemove    = "acc:remove:"
	cbPickPackageNew   = "pkgn:"
	cbPickPackageRenew = "pkgr:"
)

type Deps struct {
	API        *tgbotapi.BotAPI
	Repo       *repository.Repository
	Sessions   *session.Store
	Verifier   *auth.Verifier
	ReceiptDir string
}

type Hub struct {
	deps Deps
}

func NewHub(d Deps) *Hub {
	return &Hub{deps: d}
}

// =============================================================================
// Helpers
// =============================================================================

// LanguageFor returns the preferred language for the given chat. Falls back
// to the default language from settings when no account is linked yet.
func (h *Hub) LanguageFor(ctx context.Context, chatID int64) string {
	accounts, err := h.deps.Repo.AccountsByChatID(ctx, chatID)
	if err == nil {
		for _, a := range accounts {
			if a.Language != "" {
				return a.Language
			}
		}
	}
	settings, err := h.deps.Repo.Settings(ctx)
	if err != nil || settings.DefaultLanguage == "" {
		return models.TelegramLanguageEN
	}
	return settings.DefaultLanguage
}

func (h *Hub) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = parseModeHTML
	msg.DisableWebPagePreview = true
	if _, err := h.deps.API.Send(msg); err != nil {
		logger.Warn("telegram_bot: send failed: %v", err)
	}
}

func (h *Hub) sendKB(chatID int64, text string, markup tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = parseModeHTML
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = markup
	if _, err := h.deps.API.Send(msg); err != nil {
		logger.Warn("telegram_bot: send failed: %v", err)
	}
}

func (h *Hub) deleteMessage(chatID int64, messageID int) {
	cfg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, _ = h.deps.API.Request(cfg)
}

// sendTyping pings Telegram so the user sees a "typing..." indicator while
// we run a database lookup. Best-effort, errors are ignored.
func (h *Hub) sendTyping(chatID int64) {
	_, _ = h.deps.API.Request(tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping))
}

// respond either edits the source message in place (when srcMsgID > 0) or
// sends a new message. This keeps menu navigation feeling like a single,
// updating screen instead of an ever-growing chat log. When the edit fails
// (e.g. message too old or identical content), it falls back to sending a
// new message so the user is never left without a response.
func (h *Hub) respond(chatID int64, srcMsgID int, text string, markup *tgbotapi.InlineKeyboardMarkup) {
	if srcMsgID > 0 {
		if markup != nil {
			edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, srcMsgID, text, *markup)
			edit.ParseMode = parseModeHTML
			edit.DisableWebPagePreview = true
			if _, err := h.deps.API.Send(edit); err == nil {
				return
			} else {
				logger.Warn("telegram_bot: edit failed for chat=%d msg=%d: %v", chatID, srcMsgID, err)
			}
		} else {
			edit := tgbotapi.NewEditMessageText(chatID, srcMsgID, text)
			edit.ParseMode = parseModeHTML
			edit.DisableWebPagePreview = true
			if _, err := h.deps.API.Send(edit); err == nil {
				return
			} else {
				logger.Warn("telegram_bot: edit failed for chat=%d msg=%d: %v", chatID, srcMsgID, err)
			}
		}
	}
	if markup != nil {
		h.sendKB(chatID, text, *markup)
		return
	}
	h.send(chatID, text)
}

func mainMenuKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnAddAccount), cbAddAccount),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnMyAccounts), cbMyAccounts),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnNewOrder), cbNewOrder),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnLanguage), cbLanguage),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnHelp), cbHelp),
		),
	)
}

func languageKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("English", cbLangEN),
			tgbotapi.NewInlineKeyboardButtonData("فارسی", cbLangFA),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbMainMenu),
		),
	)
}

func backToMenuKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbMainMenu),
		),
	)
}

func accountDetailKeyboard(accountID uint, lang string) tgbotapi.InlineKeyboardMarkup {
	idStr := strconv.FormatUint(uint64(accountID), 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnUsage), cbAccountUsage+idStr),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnRenew), cbAccountRenew+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnRemove), cbAccountRemove+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbMyAccounts),
		),
	)
}

func packageKeyboard(packages []models.TelegramPackage, prefix, lang string) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(packages)+1)
	for _, p := range packages {
		title := p.Title
		if p.PriceText != "" {
			title = title + " (" + p.PriceText + ")"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(title, prefix+strconv.FormatUint(uint64(p.ID), 10)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbMainMenu),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// =============================================================================
// Top-level menu actions
// =============================================================================

func (h *Hub) HandleStart(ctx context.Context, m *tgbotapi.Message) {
	chatID := m.Chat.ID
	settings, err := h.deps.Repo.Settings(ctx)
	if err != nil || !settings.Enabled {
		h.send(chatID, i18n.T(h.LanguageFor(ctx, chatID), i18n.BotDisabled))
		return
	}
	lang := h.LanguageFor(ctx, chatID)
	text := i18n.T(lang, i18n.Welcome) + "\n\n" + i18n.T(lang, i18n.MainMenu)
	kb := mainMenuKeyboard(lang)
	h.sendKB(chatID, text, kb)
}

// SendMainMenu renders the main menu either by editing the source message
// (when srcMsgID > 0) or by sending a new one.
func (h *Hub) SendMainMenu(ctx context.Context, chatID int64, lang string, srcMsgID int) {
	kb := mainMenuKeyboard(lang)
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.MainMenu), &kb)
}

func (h *Hub) ShowLanguageMenu(ctx context.Context, chatID int64, lang string, srcMsgID int) {
	kb := languageKeyboard(lang)
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.BtnLanguage), &kb)
}

func (h *Hub) SetLanguage(ctx context.Context, chatID int64, newLang string, srcMsgID int) {
	if err := h.deps.Repo.UpdateLanguageForChat(ctx, chatID, newLang); err != nil {
		logger.Warn("telegram_bot: failed to update language: %v", err)
	}
	text := i18n.T(newLang, i18n.LanguagePicked) + "\n\n" + i18n.T(newLang, i18n.MainMenu)
	kb := mainMenuKeyboard(newLang)
	h.respond(chatID, srcMsgID, text, &kb)
}

func (h *Hub) ShowHelp(ctx context.Context, chatID int64, lang string, srcMsgID int) {
	kb := backToMenuKeyboard(lang)
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.HelpText), &kb)
}

// =============================================================================
// Account linking flow
// =============================================================================

func (h *Hub) StartAddAccount(ctx context.Context, chatID int64, srcMsgID int) {
	lang := h.LanguageFor(ctx, chatID)
	if !h.deps.Sessions.RegisterAttempt(chatID) {
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.RateLimited), nil)
		return
	}
	h.deps.Sessions.Set(chatID, &session.Session{State: session.WaitingUsernameForLink})
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.AskUsername), nil)
}

func (h *Hub) HandleStateful(ctx context.Context, m *tgbotapi.Message) bool {
	chatID := m.Chat.ID
	sess := h.deps.Sessions.Get(chatID)
	if sess.State == session.Idle {
		return false
	}

	lang := h.LanguageFor(ctx, chatID)
	text := strings.TrimSpace(m.Text)

	switch sess.State {
	case session.WaitingUsernameForLink:
		sess.BufferUsername = text
		sess.State = session.WaitingPasswordForLink
		h.deps.Sessions.Set(chatID, sess)
		h.send(chatID, i18n.T(lang, i18n.AskPassword))
		return true

	case session.WaitingPasswordForLink:
		username := sess.BufferUsername
		password := text
		// Delete the password message immediately so it does not linger in
		// the chat history.
		h.deleteMessage(chatID, m.MessageID)
		h.completeLink(ctx, chatID, username, password, lang)
		return true

	case session.WaitingUsernameForNew:
		if !validNewUsername(text) {
			h.send(chatID, i18n.T(lang, i18n.AskUsernameNew))
			return true
		}
		sess.BufferDesired = text
		sess.State = session.WaitingPackageForNew
		h.deps.Sessions.Set(chatID, sess)
		h.sendPackages(ctx, chatID, lang, cbPickPackageNew, 0)
		return true

	case session.WaitingNoteForNew:
		note := text
		if note == "/skip" {
			note = ""
		}
		h.finalizeNewRequest(ctx, chatID, sess, note, lang)
		return true

	case session.WaitingNoteForRenew:
		note := text
		if note == "/skip" {
			note = ""
		}
		h.finalizeRenewRequest(ctx, chatID, sess, note, lang)
		return true
	}
	return false
}

func (h *Hub) HandleSkip(ctx context.Context, m *tgbotapi.Message) {
	sess := h.deps.Sessions.Get(m.Chat.ID)
	switch sess.State {
	case session.WaitingNoteForNew, session.WaitingNoteForRenew:
		// Reuse the same code path the normal text handler uses.
		m.Text = "/skip"
		h.HandleStateful(ctx, m)
	}
}

func (h *Hub) completeLink(ctx context.Context, chatID int64, username, password, lang string) {
	user, err := h.deps.Verifier.Verify(ctx, username, password)
	if err != nil {
		h.deps.Sessions.Reset(chatID)
		switch {
		case errors.Is(err, auth.ErrUserLocked):
			h.send(chatID, i18n.T(lang, i18n.AuthLocked))
		case errors.Is(err, auth.ErrUserInactive):
			h.send(chatID, i18n.T(lang, i18n.OcservDeactivated))
		default:
			h.send(chatID, i18n.T(lang, i18n.AuthFail))
		}
		h.SendMainMenu(ctx, chatID, lang, 0)
		return
	}

	existing, err := h.deps.Repo.AccountsByChatID(ctx, chatID)
	if err == nil {
		for _, a := range existing {
			if a.OcservUserID == user.ID {
				h.deps.Sessions.Reset(chatID)
				h.send(chatID, i18n.T(lang, i18n.AlreadyLinked))
				h.SendMainMenu(ctx, chatID, lang, 0)
				return
			}
		}
	}

	if _, err := h.deps.Repo.UpsertAccount(ctx, chatID, "", lang, user.ID); err != nil {
		logger.Warn("telegram_bot: failed to link account: %v", err)
	}
	h.deps.Sessions.Reset(chatID)
	h.send(chatID, i18n.T(lang, i18n.AuthSuccess))
	h.SendMainMenu(ctx, chatID, lang, 0)
}

// =============================================================================
// My accounts
// =============================================================================

// SendMyAccounts renders the linked accounts in a single message (one button
// per account). Picking an account opens its detail submenu.
func (h *Hub) SendMyAccounts(ctx context.Context, chatID int64, lang string, srcMsgID int) {
	accounts, err := h.deps.Repo.AccountsByChatID(ctx, chatID)
	if err != nil || len(accounts) == 0 {
		text := i18n.T(lang, i18n.NoAccounts) + "\n\n" + i18n.T(lang, i18n.MainMenu)
		kb := mainMenuKeyboard(lang)
		h.respond(chatID, srcMsgID, text, &kb)
		return
	}

	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(accounts)+1)
	for _, a := range accounts {
		user, err := h.deps.Repo.OcservUserByID(ctx, a.OcservUserID)
		if err != nil {
			continue
		}
		// Inline keyboard button labels are plain text — no HTML escaping
		// required, but we strip control characters defensively.
		label := "• " + strings.ReplaceAll(user.Username, "\n", " ")
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, cbAccountDetail+strconv.FormatUint(uint64(a.ID), 10)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbMainMenu),
	))
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.BtnMyAccounts), &kb)
}

// ShowAccountDetail renders the per-account submenu (Usage / Renew / Remove
// / Back) by editing the source message.
func (h *Hub) ShowAccountDetail(ctx context.Context, chatID int64, accountID uint, srcMsgID int) {
	lang := h.LanguageFor(ctx, chatID)
	account, err := h.deps.Repo.AccountByID(ctx, accountID)
	if err != nil || account.ChatID != chatID {
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.NotLinked), nil)
		return
	}
	user, err := h.deps.Repo.OcservUserByID(ctx, account.OcservUserID)
	if err != nil {
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.NotLinked), nil)
		return
	}
	kb := accountDetailKeyboard(accountID, lang)
	text := "👤 <b>" + htmlEscape(user.Username) + "</b>"
	h.respond(chatID, srcMsgID, text, &kb)
}

func (h *Hub) SendAccountUsage(ctx context.Context, chatID int64, accountID uint, lang string, srcMsgID int) {
	h.sendTyping(chatID)
	account, err := h.deps.Repo.AccountByID(ctx, accountID)
	if err != nil || account.ChatID != chatID {
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.NotLinked), nil)
		return
	}
	user, err := h.deps.Repo.OcservUserByID(ctx, account.OcservUserID)
	if err != nil {
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.NotLinked), nil)
		return
	}
	status := "active"
	if user.IsLocked {
		status = "locked"
	}
	if user.DeactivatedAt != nil {
		status = "deactivated"
	}
	expires := "—"
	if user.ExpireAt != nil {
		expires = user.ExpireAt.Format("2006-01-02")
	}
	rxGB := float64(user.Rx) / (1 << 30)
	txGB := float64(user.Tx) / (1 << 30)
	msg := i18n.T(lang, i18n.UsageText, htmlEscape(user.Username), status, user.TrafficSize, rxGB, txGB, expires)

	idStr := strconv.FormatUint(uint64(accountID), 10)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbAccountDetail+idStr),
		),
	)
	h.respond(chatID, srcMsgID, msg, &kb)
}

func (h *Hub) RemoveAccount(ctx context.Context, chatID int64, accountID uint, srcMsgID int) {
	lang := h.LanguageFor(ctx, chatID)
	account, err := h.deps.Repo.AccountByID(ctx, accountID)
	if err != nil || account.ChatID != chatID {
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.NotLinked), nil)
		return
	}
	if err := h.deps.Repo.DeleteAccount(ctx, accountID); err != nil {
		logger.Warn("telegram_bot: failed to delete account: %v", err)
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.NotLinked), nil)
		return
	}
	text := i18n.T(lang, i18n.AccountRemoved) + "\n\n" + i18n.T(lang, i18n.MainMenu)
	kb := mainMenuKeyboard(lang)
	h.respond(chatID, srcMsgID, text, &kb)
}

// =============================================================================
// New / Renew flows
// =============================================================================

func (h *Hub) StartNewOrder(ctx context.Context, chatID int64, srcMsgID int) {
	lang := h.LanguageFor(ctx, chatID)

	pending, err := h.deps.Repo.PendingByChat(ctx, chatID)
	if err == nil && pending != nil {
		kb := backToMenuKeyboard(lang)
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.RequestExists), &kb)
		return
	}

	h.deps.Sessions.Set(chatID, &session.Session{State: session.WaitingUsernameForNew})
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.AskUsernameNew), nil)
}

func (h *Hub) StartRenewForAccount(ctx context.Context, chatID int64, accountID uint, srcMsgID int) {
	lang := h.LanguageFor(ctx, chatID)
	account, err := h.deps.Repo.AccountByID(ctx, accountID)
	if err != nil || account.ChatID != chatID {
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.NotLinked), nil)
		return
	}
	pending, err := h.deps.Repo.PendingByChat(ctx, chatID)
	if err == nil && pending != nil {
		kb := backToMenuKeyboard(lang)
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.RequestExists), &kb)
		return
	}
	h.deps.Sessions.Set(chatID, &session.Session{
		State:          session.WaitingPackageForRenew,
		BufferTargetID: account.OcservUserID,
	})
	h.sendPackages(ctx, chatID, lang, cbPickPackageRenew, srcMsgID)
}

func (h *Hub) sendPackages(ctx context.Context, chatID int64, lang, prefix string, srcMsgID int) {
	packages, err := h.deps.Repo.ActivePackages(ctx)
	if err != nil || len(packages) == 0 {
		text := i18n.T(lang, i18n.NoPackages) + "\n\n" + i18n.T(lang, i18n.MainMenu)
		kb := mainMenuKeyboard(lang)
		h.respond(chatID, srcMsgID, text, &kb)
		return
	}
	kb := packageKeyboard(packages, prefix, lang)
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.PickPackage), &kb)
}

func (h *Hub) PickedPackageNew(ctx context.Context, chatID int64, packageID uint, srcMsgID int) {
	sess := h.deps.Sessions.Get(chatID)
	lang := h.LanguageFor(ctx, chatID)
	if sess.State != session.WaitingPackageForNew {
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.SessionTimedOut), nil)
		return
	}
	sess.BufferPackage = packageID
	sess.State = session.WaitingNoteForNew
	h.deps.Sessions.Set(chatID, sess)
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.AskMessage), nil)
}

func (h *Hub) PickedPackageRenew(ctx context.Context, chatID int64, packageID uint, srcMsgID int) {
	sess := h.deps.Sessions.Get(chatID)
	lang := h.LanguageFor(ctx, chatID)
	if sess.State != session.WaitingPackageForRenew {
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.SessionTimedOut), nil)
		return
	}
	sess.BufferPackage = packageID
	sess.State = session.WaitingNoteForRenew
	h.deps.Sessions.Set(chatID, sess)
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.AskMessage), nil)
}

func (h *Hub) finalizeNewRequest(ctx context.Context, chatID int64, sess *session.Session, note, lang string) {
	pkgID := sess.BufferPackage
	desired := sess.BufferDesired

	req := &models.TelegramRequest{
		ChatID:          chatID,
		Type:            models.TelegramRequestTypeNew,
		PackageID:       ptrUint(pkgID),
		DesiredUsername: desired,
		Status:          models.TelegramRequestStatusPending,
		UserMessage:     note,
	}
	created, err := h.deps.Repo.CreateRequest(ctx, req)
	if err != nil {
		logger.Warn("telegram_bot: failed to create request: %v", err)
		h.send(chatID, i18n.T(lang, i18n.UnknownCommand))
		return
	}
	h.deps.Sessions.Reset(chatID)
	h.send(chatID, i18n.T(lang, i18n.RequestCreated))
	h.SendMainMenu(ctx, chatID, lang, 0)

	go h.notifyAdmin(ctx, "New account request",
		fmt.Sprintf("Request #%d (new) — chat=%d desired=%s package=%d note=%s",
			created.ID, chatID, desired, pkgID, note))
}

func (h *Hub) finalizeRenewRequest(ctx context.Context, chatID int64, sess *session.Session, note, lang string) {
	pkgID := sess.BufferPackage
	target := sess.BufferTargetID

	req := &models.TelegramRequest{
		ChatID:         chatID,
		Type:           models.TelegramRequestTypeRenew,
		PackageID:      ptrUint(pkgID),
		TargetOcservID: ptrUint(target),
		Status:         models.TelegramRequestStatusPending,
		UserMessage:    note,
	}
	created, err := h.deps.Repo.CreateRequest(ctx, req)
	if err != nil {
		logger.Warn("telegram_bot: failed to create request: %v", err)
		h.send(chatID, i18n.T(lang, i18n.UnknownCommand))
		return
	}
	h.deps.Sessions.Reset(chatID)
	h.send(chatID, i18n.T(lang, i18n.RequestCreated))
	h.SendMainMenu(ctx, chatID, lang, 0)

	go h.notifyAdmin(ctx, "Renewal request",
		fmt.Sprintf("Request #%d (renew) — chat=%d target_user=%d package=%d note=%s",
			created.ID, chatID, target, pkgID, note))
}

// =============================================================================
// Photo handler — receipt upload
// =============================================================================

func (h *Hub) HandlePhoto(ctx context.Context, m *tgbotapi.Message) {
	chatID := m.Chat.ID
	lang := h.LanguageFor(ctx, chatID)

	pending, err := h.deps.Repo.PendingByChat(ctx, chatID)
	if err != nil || pending == nil {
		h.send(chatID, i18n.T(lang, i18n.NotApprovedYet))
		return
	}
	if pending.Status != models.TelegramRequestStatusAwaitingPayment {
		h.send(chatID, i18n.T(lang, i18n.NotApprovedYet))
		return
	}

	photo := m.Photo[len(m.Photo)-1]
	fileURL, err := h.deps.API.GetFileDirectURL(photo.FileID)
	if err != nil {
		logger.Warn("telegram_bot: get file url failed: %v", err)
		return
	}

	if err := os.MkdirAll(h.deps.ReceiptDir, 0o750); err != nil {
		logger.Warn("telegram_bot: mkdir receipts: %v", err)
		return
	}
	path := filepath.Join(h.deps.ReceiptDir, fmt.Sprintf("req_%d_%d.jpg", pending.ID, time.Now().Unix()))

	if err := downloadFile(fileURL, path); err != nil {
		logger.Warn("telegram_bot: download receipt: %v", err)
		return
	}

	if err := h.deps.Repo.AttachReceipt(ctx, pending.ID, path); err != nil {
		logger.Warn("telegram_bot: attach receipt: %v", err)
		return
	}

	h.send(chatID, i18n.T(lang, i18n.ReceiptSaved))

	go h.notifyAdmin(ctx, "Receipt uploaded",
		fmt.Sprintf("Receipt for request #%d uploaded by chat=%d", pending.ID, chatID))
}

// =============================================================================
// Misc
// =============================================================================

func (h *Hub) notifyAdmin(ctx context.Context, title, body string) {
	settings, err := h.deps.Repo.Settings(ctx)
	if err != nil || settings.AdminChatID == 0 {
		return
	}
	text := fmt.Sprintf("[%s]\n%s", title, body)
	msg := tgbotapi.NewMessage(settings.AdminChatID, text)
	if _, err := h.deps.API.Send(msg); err != nil {
		logger.Warn("telegram_bot: notifyAdmin failed: %v", err)
	}
}

func ptrUint(v uint) *uint {
	if v == 0 {
		return nil
	}
	out := v
	return &out
}

func validNewUsername(s string) bool {
	if len(s) < 3 || len(s) > 32 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == '.':
		default:
			return false
		}
	}
	return true
}

func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download status %d", resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}
