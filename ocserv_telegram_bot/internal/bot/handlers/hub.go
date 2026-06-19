package handlers

import (
	"context"
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
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/ocserv_telegram_bot/internal/auth"
	"github.com/mmtaee/ocserv-dashboard/ocserv_telegram_bot/internal/bot/cbdata"
	"github.com/mmtaee/ocserv-dashboard/ocserv_telegram_bot/internal/i18n"
	"github.com/mmtaee/ocserv-dashboard/ocserv_telegram_bot/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/ocserv_telegram_bot/internal/session"
)

// All bot messages are sent with parse_mode=HTML so the i18n catalog can use
// <b>, <code>, <i> etc. User-supplied values (usernames, notes) MUST be
// passed through htmlEscape before being interpolated.
const parseModeHTML = "HTML"

func htmlEscape(s string) string {
	return html.EscapeString(s)
}

type Deps struct {
	API        *tgbotapi.BotAPI
	Repo       *repository.Repository
	Sessions   *session.Store
	Verifier   *auth.Verifier
	ReceiptDir string
	// BrandName is what is shown in the welcome banner — typically the
	// bot's display name from BotFather (api.Self.FirstName) but callers
	// may pass a different label here.
	BrandName string
}

type Hub struct {
	deps Deps
}

func NewHub(d Deps) *Hub {
	if d.BrandName == "" {
		// Neutral fallback used only when Telegram cannot tell us the bot's
		// own first name / username (essentially never in production).
		d.BrandName = "VPN"
	}
	return &Hub{deps: d}
}

// =============================================================================
// Helpers
// =============================================================================

// SyncTelegramUsernameFromAPI calls Telegram getChat for this chat_id and, when
// the user has a public username on their profile, persists it on every
// telegram_accounts row for that chat. A numeric chat_id cannot be turned into
// t.me/username without Telegram returning a username (getChat is authoritative).
func (h *Hub) SyncTelegramUsernameFromAPI(ctx context.Context, chatID int64) {
	cfg := tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
	}
	chat, err := h.deps.API.GetChat(cfg)
	if err != nil {
		logger.Warn("ocserv_telegram_bot: getChat(%d) failed: %v", chatID, err)
		return
	}
	un := strings.TrimPrefix(strings.TrimSpace(chat.UserName), "@")
	if un == "" {
		return
	}
	if err := h.deps.Repo.SetTelegramUsernameForChat(ctx, chatID, un); err != nil {
		logger.Warn("ocserv_telegram_bot: set telegram_username from getChat failed: %v", err)
	}
}

// IsAdmin reports whether the chat ID is the configured admin chat.
func (h *Hub) IsAdmin(ctx context.Context, chatID int64) bool {
	settings, err := h.deps.Repo.Settings(ctx)
	if err != nil || settings.AdminChatID == 0 {
		return false
	}
	return settings.AdminChatID == chatID
}

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
		logger.Warn("ocserv_telegram_bot: send failed: %v", err)
	}
}

func (h *Hub) sendKB(chatID int64, text string, markup tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = parseModeHTML
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = markup
	if _, err := h.deps.API.Send(msg); err != nil {
		logger.Warn("ocserv_telegram_bot: send failed: %v", err)
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
				logger.Warn("ocserv_telegram_bot: edit failed for chat=%d msg=%d: %v", chatID, srcMsgID, err)
			}
		} else {
			edit := tgbotapi.NewEditMessageText(chatID, srcMsgID, text)
			edit.ParseMode = parseModeHTML
			edit.DisableWebPagePreview = true
			if _, err := h.deps.API.Send(edit); err == nil {
				return
			} else {
				logger.Warn("ocserv_telegram_bot: edit failed for chat=%d msg=%d: %v", chatID, srcMsgID, err)
			}
		}
	}
	if markup != nil {
		h.sendKB(chatID, text, *markup)
		return
	}
	h.send(chatID, text)
}

func adminMenuKeyboard(lang, panelURL string) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnAdminPending), cbdata.AdminPending),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnAdminReceipts), cbdata.AdminReceipts),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnAdminStats), cbdata.AdminStats),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnAdminUserView), cbdata.MainMenu),
		),
	}
	if panelURL != "" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(i18n.T(lang, i18n.BtnOpenPanel), panelURL),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnLanguage), cbdata.Language),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func mainMenuKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnAddAccount), cbdata.AddAccount),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnMyAccounts), cbdata.MyAccounts),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnNewOrder), cbdata.NewOrder),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnLanguage), cbdata.Language),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnHelp), cbdata.Help),
		),
	)
}

// adminUserViewKeyboard is the regular user menu with an extra "Back to Admin" row at the bottom,
// shown when the admin is previewing the user view.
func adminUserViewKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnAddAccount), cbdata.AddAccount),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnMyAccounts), cbdata.MyAccounts),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnNewOrder), cbdata.NewOrder),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnLanguage), cbdata.Language),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnHelp), cbdata.Help),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnAdminBack), cbdata.AdminMenu),
		),
	)
}

func languageKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	for _, opt := range models.TelegramLanguages {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(
			opt.Label,
			models.TelegramLangCallback(opt.Code),
		))
		if len(row) == 2 {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbdata.MainMenu),
	})
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func backToMenuKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbdata.MainMenu),
		),
	)
}

func accountDetailKeyboard(accountID uint, lang string) tgbotapi.InlineKeyboardMarkup {
	idStr := strconv.FormatUint(uint64(accountID), 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnUsage), cbdata.AccountUsage+idStr),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnRenew), cbdata.AccountRenew+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnRemove), cbdata.AccountRemove+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbdata.MyAccounts),
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
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbdata.MainMenu),
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

	h.SyncTelegramUsernameFromAPI(ctx, chatID)

	if settings.AdminChatID != 0 && settings.AdminChatID == chatID {
		text := i18n.T(lang, i18n.AdminWelcome, htmlEscape(h.deps.BrandName))
		kb := adminMenuKeyboard(lang, panelURL(settings))
		h.sendKB(chatID, text, kb)
		return
	}

	text := i18n.T(lang, i18n.Welcome, htmlEscape(h.deps.BrandName)) + "\n\n" + i18n.T(lang, i18n.MainMenu)
	kb := mainMenuKeyboard(lang)
	h.sendKB(chatID, text, kb)
}

// SendMainMenu renders the user main menu. Admins get the admin menu
// instead so /start always lands on the most useful screen for them.
func (h *Hub) SendMainMenu(ctx context.Context, chatID int64, lang string, srcMsgID int) {
	if h.IsAdmin(ctx, chatID) {
		h.SendAdminMenu(ctx, chatID, lang, srcMsgID)
		return
	}
	kb := mainMenuKeyboard(lang)
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.MainMenu), &kb)
}

// SendUserMenu always renders the user menu, even for the admin chat.
// When called for the admin, it adds a "Back to Admin" button so they
// can return to the admin panel at any time.
func (h *Hub) SendUserMenu(ctx context.Context, chatID int64, lang string, srcMsgID int) {
	if h.IsAdmin(ctx, chatID) {
		kb := adminUserViewKeyboard(lang)
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.MainMenu), &kb)
		return
	}
	kb := mainMenuKeyboard(lang)
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.MainMenu), &kb)
}

// SendAdminMenu shows the admin actions panel.
func (h *Hub) SendAdminMenu(ctx context.Context, chatID int64, lang string, srcMsgID int) {
	h.sendTyping(chatID)
	settings, _ := h.deps.Repo.Settings(ctx)
	kb := adminMenuKeyboard(lang, panelURL(settings))
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.AdminMenu), &kb)
}

// =============================================================================
// Admin views
// =============================================================================

func (h *Hub) ShowAdminPending(ctx context.Context, chatID int64, lang string, srcMsgID int) {
	h.sendTyping(chatID)
	requests, err := h.deps.Repo.RequestsByStatuses(ctx, []string{models.TelegramRequestStatusPending}, 10)
	settings, _ := h.deps.Repo.Settings(ctx)
	if err != nil || len(requests) == 0 {
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.AdminNoPending), backToAdminKeyboard(lang, panelURL(settings)))
		return
	}
	body := renderRequestList(lang, requests)
	h.respond(chatID, srcMsgID, body, backToAdminKeyboard(lang, panelURL(settings)))
}

func (h *Hub) ShowAdminReceipts(ctx context.Context, chatID int64, lang string, srcMsgID int) {
	h.sendTyping(chatID)
	requests, err := h.deps.Repo.RequestsByStatuses(ctx, []string{
		models.TelegramRequestStatusAwaitingPayment,
		models.TelegramRequestStatusPaymentUploaded,
	}, 10)
	settings, _ := h.deps.Repo.Settings(ctx)
	if err != nil || len(requests) == 0 {
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.AdminNoReceipts), backToAdminKeyboard(lang, panelURL(settings)))
		return
	}
	body := renderRequestList(lang, requests)
	h.respond(chatID, srcMsgID, body, backToAdminKeyboard(lang, panelURL(settings)))
}

func (h *Hub) ShowAdminStats(ctx context.Context, chatID int64, lang string, srcMsgID int) {
	h.sendTyping(chatID)
	stats, err := h.deps.Repo.AdminStats(ctx)
	settings, _ := h.deps.Repo.Settings(ctx)
	if err != nil {
		logger.Warn("ocserv_telegram_bot: AdminStats failed: %v", err)
		h.respond(chatID, srcMsgID, i18n.T(lang, i18n.UnknownCommand), backToAdminKeyboard(lang, panelURL(settings)))
		return
	}
	text := i18n.T(lang, i18n.AdminStatsText,
		stats.LinkedAccounts,
		stats.ActivePackages,
		stats.PendingRequests,
		stats.AwaitingPayments,
		stats.UploadedReceipts,
	)
	h.respond(chatID, srcMsgID, text, backToAdminKeyboard(lang, panelURL(settings)))
}

func backToAdminKeyboard(lang, panel string) *tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbdata.AdminMenu),
		),
	}
	if panel != "" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(i18n.T(lang, i18n.BtnOpenPanel), panel),
		))
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return &kb
}

func renderRequestList(lang string, requests []models.TelegramRequest) string {
	var b strings.Builder
	for i, r := range requests {
		if i > 0 {
			b.WriteString("\n\n")
		}
		label := r.DesiredUsername
		if label == "" && r.TargetOcservID != nil {
			label = "user_id=" + strconv.FormatUint(uint64(*r.TargetOcservID), 10)
		}
		if label == "" {
			label = "—"
		}
		note := r.UserMessage
		if note == "" {
			note = "—"
		}
		b.WriteString(i18n.T(lang, i18n.AdminRequestRow,
			r.ID,
			r.Type,
			htmlEscape(label),
			htmlEscape(note),
			r.CreatedAt.Format("2006-01-02 15:04"),
		))
	}
	return b.String()
}

// panelURL builds the admin web panel deep link from TelegramSettings.OcservHost.
// Telegram inline URL buttons require https; missing scheme is normalized.
func panelURL(s *models.TelegramSettings) string {
	if s == nil {
		return ""
	}
	host := strings.TrimSpace(s.OcservHost)
	if host == "" {
		return ""
	}
	host = strings.TrimRight(host, "/")
	switch {
	case strings.HasPrefix(host, "https://"):
		// keep
	case strings.HasPrefix(host, "http://"):
		host = "https://" + strings.TrimPrefix(host, "http://")
	default:
		host = "https://" + host
	}
	return host + "/telegram/requests"
}

func (h *Hub) ShowLanguageMenu(ctx context.Context, chatID int64, lang string, srcMsgID int) {
	kb := languageKeyboard(lang)
	h.respond(chatID, srcMsgID, i18n.T(lang, i18n.BtnLanguage), &kb)
}

func (h *Hub) SetLanguage(ctx context.Context, chatID int64, newLang string, srcMsgID int) {
	if !models.IsTelegramLanguage(newLang) {
		newLang = models.TelegramLanguageEN
	}
	if err := h.deps.Repo.UpdateLanguageForChat(ctx, chatID, newLang); err != nil {
		logger.Warn("ocserv_telegram_bot: failed to update language: %v", err)
	}
	text := i18n.T(newLang, i18n.LanguagePicked) + "\n\n" + i18n.T(newLang, i18n.MainMenu)
	kb := mainMenuKeyboard(newLang)
	h.respond(chatID, srcMsgID, text, &kb)
}

func (h *Hub) ShowHelp(ctx context.Context, chatID int64, lang string, srcMsgID int) {
	settings, _ := h.deps.Repo.Settings(ctx)
	text := i18n.T(lang, i18n.HelpText) + supportFooter(lang, settings)

	if link, label := supportLink(settings); link != "" {
		row1 := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(label, link),
		)
		row2 := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbdata.MainMenu),
		)
		kb := tgbotapi.NewInlineKeyboardMarkup(row1, row2)
		h.respond(chatID, srcMsgID, text, &kb)
		return
	}

	kb := backToMenuKeyboard(lang)
	h.respond(chatID, srcMsgID, text, &kb)
}

// supportLink returns the t.me URL and a localized button label, or empty
// strings when no support_username is configured.
func supportLink(s *models.TelegramSettings) (string, string) {
	if s == nil {
		return "", ""
	}
	handle := strings.TrimPrefix(strings.TrimSpace(s.SupportUsername), "@")
	if handle == "" {
		return "", ""
	}
	return "https://t.me/" + handle, "💬 @" + handle
}

// supportFooter is the localized "for help, contact @handle" trailer.
func supportFooter(lang string, s *models.TelegramSettings) string {
	if s == nil {
		return ""
	}
	handle := strings.TrimPrefix(strings.TrimSpace(s.SupportUsername), "@")
	if handle == "" {
		return ""
	}
	link := `<a href="https://t.me/` + handle + `">@` + handle + `</a>`
	if models.IsTelegramRTL(lang) {
		return "\n\n\u200f<b>پشتیبانی:</b> " + link + " 💬"
	}
	return "\n\n💬 <b>Support:</b> " + link
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
		h.completeLink(ctx, chatID, username, password, lang, fromUsername(m))
		return true

	case session.WaitingUsernameForNew:
		if !validNewUsername(text) {
			h.send(chatID, i18n.T(lang, i18n.AskUsernameNew))
			return true
		}
		sess.BufferDesired = text
		sess.State = session.WaitingPackageForNew
		h.deps.Sessions.Set(chatID, sess)
		h.sendPackages(ctx, chatID, lang, cbdata.PickPackageNew, 0)
		return true

	case session.WaitingNoteForNew:
		note := text
		if note == "/skip" {
			note = ""
		}
		h.finalizeNewRequest(ctx, chatID, sess, note, lang, fromUsername(m))
		return true

	case session.WaitingNoteForRenew:
		note := text
		if note == "/skip" {
			note = ""
		}
		h.finalizeRenewRequest(ctx, chatID, sess, note, lang, fromUsername(m))
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

func (h *Hub) completeLink(ctx context.Context, chatID int64, username, password, lang, tgUsername string) {
	user, err := h.deps.Verifier.Verify(ctx, username, password)
	if err != nil {
		h.deps.Sessions.Reset(chatID)
		h.send(chatID, i18n.T(lang, i18n.AuthFail))
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

	if _, err := h.deps.Repo.UpsertAccount(ctx, chatID, tgUsername, lang, user.ID); err != nil {
		logger.Warn("ocserv_telegram_bot: failed to link account: %v", err)
	}
	h.SyncTelegramUsernameFromAPI(ctx, chatID)
	h.deps.Sessions.Reset(chatID)
	h.send(chatID, i18n.T(lang, i18n.AuthSuccess))
	// Surface a one-line hint so the user immediately knows the linked
	// account needs renewal (locked = quota/expiry exhausted, deactivated =
	// disabled by admin or auto-disabled). Both states can be recovered by
	// renewing through the bot.
	if user.IsLocked || user.DeactivatedAt != nil {
		h.send(chatID, i18n.T(lang, i18n.LinkedLockedHint))
	}
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

	ids := make([]uint, len(accounts))
	for i, a := range accounts {
		ids[i] = a.OcservUserID
	}
	usersByID, err := h.deps.Repo.OcservUsersByIDs(ctx, ids)
	if err != nil {
		logger.Warn("ocserv_telegram_bot: OcservUsersByIDs: %v", err)
		usersByID = make(map[uint]*models.OcservUser)
		for _, a := range accounts {
			u, e := h.deps.Repo.OcservUserByID(ctx, a.OcservUserID)
			if e != nil {
				continue
			}
			usersByID[a.OcservUserID] = u
		}
	}

	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(accounts)+1)
	for _, a := range accounts {
		user := usersByID[a.OcservUserID]
		if user == nil {
			continue
		}
		// Inline keyboard button labels are plain text — no HTML escaping
		// required, but we strip control characters defensively.
		label := "• " + strings.ReplaceAll(user.Username, "\n", " ")
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, cbdata.AccountDetail+strconv.FormatUint(uint64(a.ID), 10)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbdata.MainMenu),
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
	// In the Ocserv data model, Rx is bytes received by the server from the
	// client (the client's upload), and Tx is bytes sent by the server to the
	// client (the client's download). The bot displays everything from the
	// user's point of view, so we map them accordingly.
	downloadGB := float64(user.Tx) / (1 << 30)
	uploadGB := float64(user.Rx) / (1 << 30)
	msg := i18n.T(lang, i18n.UsageText, htmlEscape(user.Username), status, user.TrafficSize, downloadGB, uploadGB, expires)

	idStr := strconv.FormatUint(uint64(accountID), 10)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnBack), cbdata.AccountDetail+idStr),
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
		logger.Warn("ocserv_telegram_bot: failed to delete account: %v", err)
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
	h.sendPackages(ctx, chatID, lang, cbdata.PickPackageRenew, srcMsgID)
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

func (h *Hub) finalizeNewRequest(ctx context.Context, chatID int64, sess *session.Session, note, lang, tgUsername string) {
	pkgID := sess.BufferPackage
	desired := sess.BufferDesired

	req := &models.TelegramRequest{
		ChatID:           chatID,
		TelegramUsername: tgUsername,
		Type:             models.TelegramRequestTypeNew,
		PackageID:        ptrUint(pkgID),
		DesiredUsername:  desired,
		Status:           models.TelegramRequestStatusPending,
		UserMessage:      note,
	}
	created, err := h.deps.Repo.CreateRequest(ctx, req)
	if err != nil {
		logger.Warn("ocserv_telegram_bot: failed to create request: %v", err)
		h.send(chatID, i18n.T(lang, i18n.UnknownCommand))
		return
	}
	h.deps.Sessions.Reset(chatID)
	h.send(chatID, i18n.T(lang, i18n.RequestCreated))
	h.SendMainMenu(ctx, chatID, lang, 0)

	go h.notifyAdminNewRequest(ctx, created)
}

func (h *Hub) finalizeRenewRequest(ctx context.Context, chatID int64, sess *session.Session, note, lang, tgUsername string) {
	pkgID := sess.BufferPackage
	target := sess.BufferTargetID

	req := &models.TelegramRequest{
		ChatID:           chatID,
		TelegramUsername: tgUsername,
		Type:             models.TelegramRequestTypeRenew,
		PackageID:        ptrUint(pkgID),
		TargetOcservID:   ptrUint(target),
		Status:           models.TelegramRequestStatusPending,
		UserMessage:      note,
	}
	created, err := h.deps.Repo.CreateRequest(ctx, req)
	if err != nil {
		logger.Warn("ocserv_telegram_bot: failed to create request: %v", err)
		h.send(chatID, i18n.T(lang, i18n.UnknownCommand))
		return
	}
	h.deps.Sessions.Reset(chatID)
	h.send(chatID, i18n.T(lang, i18n.RequestCreated))
	h.SendMainMenu(ctx, chatID, lang, 0)

	go h.notifyAdminRenewRequest(ctx, created)
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
		logger.Warn("ocserv_telegram_bot: get file url failed: %v", err)
		return
	}

	if err := os.MkdirAll(h.deps.ReceiptDir, 0o750); err != nil {
		logger.Warn("ocserv_telegram_bot: mkdir receipts: %v", err)
		return
	}
	path := filepath.Join(h.deps.ReceiptDir, fmt.Sprintf("req_%d_%d.jpg", pending.ID, time.Now().Unix()))

	if err := downloadFile(fileURL, path); err != nil {
		logger.Warn("ocserv_telegram_bot: download receipt: %v", err)
		return
	}

	if err := h.deps.Repo.AttachReceipt(ctx, pending.ID, path); err != nil {
		logger.Warn("ocserv_telegram_bot: attach receipt: %v", err)
		return
	}

	h.send(chatID, i18n.T(lang, i18n.ReceiptSaved))

	go h.notifyAdminReceipt(ctx, pending, m)
}

// =============================================================================
// Misc
// =============================================================================

// notifyAdmin sends a richly formatted (HTML, parse_mode=HTML) notification to
// the configured admin chat. Falls back silently when no admin chat is set.
func (h *Hub) notifyAdmin(ctx context.Context, body string) {
	settings, err := h.deps.Repo.Settings(ctx)
	if err != nil || settings.AdminChatID == 0 {
		return
	}
	msg := tgbotapi.NewMessage(settings.AdminChatID, body)
	msg.ParseMode = parseModeHTML
	msg.DisableWebPagePreview = true
	if _, err := h.deps.API.Send(msg); err != nil {
		logger.Warn("ocserv_telegram_bot: notifyAdmin failed: %v", err)
	}
}

// notifyAdminNewRequest builds and sends the admin alert for a brand-new
// account request. Uses HTML to embed:
//   - A clickable contact link to the user (works even without @username).
//   - The chosen package details (title, days, GB, price).
//   - The user's note when present.
func (h *Hub) notifyAdminNewRequest(ctx context.Context, req *models.TelegramRequest) {
	pkg := h.lookupPackage(ctx, req.PackageID)
	body := "🆕 <b>New account request</b>\n" +
		"<b>Request:</b> <code>#" + strconv.FormatUint(uint64(req.ID), 10) + "</code>\n\n" +
		formatRequester(req) +
		"<b>Desired username:</b> <code>" + safeText(req.DesiredUsername) + "</code>\n" +
		formatPackage(pkg) +
		formatUserNote(req.UserMessage) +
		"\n👉 <i>Open the dashboard panel to approve and reply.</i>"
	h.notifyAdmin(ctx, body)
}

// notifyAdminRenewRequest builds and sends the admin alert for a renewal
// request. In addition to the package details, it shows the target ocserv
// account (username + current expiry) so the admin can sanity-check the
// renewal without opening the dashboard.
func (h *Hub) notifyAdminRenewRequest(ctx context.Context, req *models.TelegramRequest) {
	pkg := h.lookupPackage(ctx, req.PackageID)
	target := h.lookupOcservUser(ctx, req.TargetOcservID)
	body := "🔄 <b>Renewal request</b>\n" +
		"<b>Request:</b> <code>#" + strconv.FormatUint(uint64(req.ID), 10) + "</code>\n\n" +
		formatRequester(req) +
		formatTargetAccount(target) +
		formatPackage(pkg) +
		formatUserNote(req.UserMessage) +
		"\n👉 <i>Open the dashboard panel to approve and reply.</i>"
	h.notifyAdmin(ctx, body)
}

// notifyAdminReceipt fires when a user has uploaded a payment receipt. The
// receipt photo itself is forwarded to the admin chat so the admin can decide
// at a glance, with a caption pointing back at the request.
func (h *Hub) notifyAdminReceipt(ctx context.Context, req *models.TelegramRequest, m *tgbotapi.Message) {
	settings, err := h.deps.Repo.Settings(ctx)
	if err != nil || settings.AdminChatID == 0 {
		return
	}

	caption := "🧾 <b>Receipt uploaded</b>\n" +
		"<b>Request:</b> <code>#" + strconv.FormatUint(uint64(req.ID), 10) + "</code>\n\n" +
		formatRequester(req) +
		"\n👉 <i>Confirm the payment in the dashboard panel.</i>"

	// Forward the photo when we have it; otherwise fall back to a plain alert
	// so the admin still gets pinged.
	if m != nil && len(m.Photo) > 0 {
		photo := m.Photo[len(m.Photo)-1]
		fwd := tgbotapi.NewPhoto(settings.AdminChatID, tgbotapi.FileID(photo.FileID))
		fwd.Caption = caption
		fwd.ParseMode = parseModeHTML
		if _, err := h.deps.API.Send(fwd); err == nil {
			return
		}
	}
	h.notifyAdmin(ctx, caption)
}

// fromUsername returns the @username of the message sender, or "" when not set
// (Telegram users without a public handle).
func fromUsername(m *tgbotapi.Message) string {
	if m == nil || m.From == nil {
		return ""
	}
	return strings.TrimSpace(m.From.UserName)
}

// formatRequester renders the requester block: a tg:// deep link that opens a
// 1:1 chat with the user (works even without a public @username), plus chat
// id and @handle when available.
func formatRequester(req *models.TelegramRequest) string {
	if req == nil {
		return ""
	}
	chatStr := strconv.FormatInt(req.ChatID, 10)
	link := `<a href="tg://user?id=` + chatStr + `">Open chat</a>`
	out := "<b>From:</b> " + link + " · <code>" + chatStr + "</code>"
	if h := strings.TrimPrefix(strings.TrimSpace(req.TelegramUsername), "@"); h != "" {
		out += " · @" + safeText(h)
	}
	return out + "\n"
}

func formatTargetAccount(user *models.OcservUser) string {
	if user == nil {
		return "<b>Target account:</b> <i>not found</i>\n"
	}
	expires := "—"
	if user.ExpireAt != nil {
		expires = user.ExpireAt.Format("2006-01-02")
	}
	return "<b>Target account:</b> <code>" + safeText(user.Username) + "</code>\n" +
		"<b>Current expiry:</b> " + expires + "\n"
}

func formatPackage(p *models.TelegramPackage) string {
	if p == nil {
		return "<b>Package:</b> <i>not found</i>\n"
	}
	out := "<b>Package:</b> " + safeText(p.Title) +
		" — " + strconv.Itoa(p.Days) + "d · " +
		strconv.Itoa(p.TrafficSizeGB) + "GB"
	if strings.TrimSpace(p.PriceText) != "" {
		out += " · " + safeText(p.PriceText)
	}
	return out + "\n"
}

func formatUserNote(note string) string {
	note = strings.TrimSpace(note)
	if note == "" {
		return ""
	}
	return "<b>Note:</b> " + safeText(note) + "\n"
}

// safeText HTML-escapes user input so it can be safely interpolated into the
// admin notification template.
func safeText(s string) string {
	return html.EscapeString(s)
}

func (h *Hub) lookupPackage(ctx context.Context, id *uint) *models.TelegramPackage {
	if id == nil {
		return nil
	}
	pkg, err := h.deps.Repo.PackageByID(ctx, *id)
	if err != nil {
		return nil
	}
	return pkg
}

func (h *Hub) lookupOcservUser(ctx context.Context, id *uint) *models.OcservUser {
	if id == nil {
		return nil
	}
	user, err := h.deps.Repo.OcservUserByID(ctx, *id)
	if err != nil {
		return nil
	}
	return user
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
