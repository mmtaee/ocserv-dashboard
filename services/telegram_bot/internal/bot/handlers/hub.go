package handlers

import (
	"context"
	"errors"
	"fmt"
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

const (
	cbMainMenu      = "menu:main"
	cbAddAccount    = "menu:add"
	cbMyAccounts    = "menu:list"
	cbNewOrder      = "menu:order"
	cbHelp          = "menu:help"
	cbLanguage      = "menu:lang"
	cbLangEN        = "lang:en"
	cbLangFA        = "lang:fa"
	cbAccountUsage  = "acc:usage:"
	cbAccountRenew  = "acc:renew:"
	cbAccountRemove = "acc:remove:"
	cbPickPackageNew    = "pkgn:"
	cbPickPackageRenew  = "pkgr:"
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
	if _, err := h.deps.API.Send(msg); err != nil {
		logger.Warn("telegram_bot: send failed: %v", err)
	}
}

func (h *Hub) sendKB(chatID int64, text string, markup tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = markup
	if _, err := h.deps.API.Send(msg); err != nil {
		logger.Warn("telegram_bot: send failed: %v", err)
	}
}

func (h *Hub) deleteMessage(chatID int64, messageID int) {
	cfg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, _ = h.deps.API.Request(cfg)
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

func languageKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("English", cbLangEN),
			tgbotapi.NewInlineKeyboardButtonData("فارسی", cbLangFA),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅", cbMainMenu),
		),
	)
}

func packageKeyboard(packages []models.TelegramPackage, prefix string) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(packages))
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
		tgbotapi.NewInlineKeyboardButtonData("⬅", cbMainMenu),
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
	h.send(chatID, i18n.T(lang, i18n.Welcome))
	h.SendMainMenu(ctx, chatID, lang)
}

func (h *Hub) SendMainMenu(ctx context.Context, chatID int64, lang string) {
	h.sendKB(chatID, i18n.T(lang, i18n.MainMenu), mainMenuKeyboard(lang))
}

func (h *Hub) HandleLanguageMenu(ctx context.Context, chatID int64, lang string) {
	h.sendKB(chatID, i18n.T(lang, i18n.BtnLanguage), languageKeyboard())
}

func (h *Hub) SetLanguage(ctx context.Context, chatID int64, lang string) {
	if err := h.deps.Repo.UpdateLanguageForChat(ctx, chatID, lang); err != nil {
		logger.Warn("telegram_bot: failed to update language: %v", err)
	}
	h.send(chatID, i18n.T(lang, i18n.LanguagePicked))
}

// =============================================================================
// Account linking flow
// =============================================================================

func (h *Hub) StartAddAccount(ctx context.Context, chatID int64) {
	lang := h.LanguageFor(ctx, chatID)
	if !h.deps.Sessions.RegisterAttempt(chatID) {
		h.send(chatID, i18n.T(lang, i18n.RateLimited))
		return
	}
	h.deps.Sessions.Set(chatID, &session.Session{State: session.WaitingUsernameForLink})
	h.send(chatID, i18n.T(lang, i18n.AskUsername))
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
		h.deps.Sessions.Set(chatID, sess)
		sess.State = session.WaitingPasswordForLink
		h.deps.Sessions.Set(chatID, sess)
		h.send(chatID, i18n.T(lang, i18n.AskPassword))
		return true

	case session.WaitingPasswordForLink:
		username := sess.BufferUsername
		password := text
		messageID := m.MessageID
		h.deleteMessage(chatID, messageID)
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
		h.sendPackages(ctx, chatID, lang, cbPickPackageNew)
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
		h.SendMainMenu(ctx, chatID, lang)
		return
	}

	existing, err := h.deps.Repo.AccountsByChatID(ctx, chatID)
	if err == nil {
		for _, a := range existing {
			if a.OcservUserID == user.ID {
				h.deps.Sessions.Reset(chatID)
				h.send(chatID, i18n.T(lang, i18n.AlreadyLinked))
				h.SendMainMenu(ctx, chatID, lang)
				return
			}
		}
	}

	if _, err := h.deps.Repo.UpsertAccount(ctx, chatID, "", lang, user.ID); err != nil {
		logger.Warn("telegram_bot: failed to link account: %v", err)
	}
	h.deps.Sessions.Reset(chatID)
	h.send(chatID, i18n.T(lang, i18n.AuthSuccess))
	h.SendMainMenu(ctx, chatID, lang)
}

// =============================================================================
// My accounts
// =============================================================================

func (h *Hub) SendMyAccounts(ctx context.Context, chatID int64, lang string) {
	accounts, err := h.deps.Repo.AccountsByChatID(ctx, chatID)
	if err != nil || len(accounts) == 0 {
		h.send(chatID, i18n.T(lang, i18n.NoAccounts))
		h.SendMainMenu(ctx, chatID, lang)
		return
	}
	for _, a := range accounts {
		user, err := h.deps.Repo.OcservUserByID(ctx, a.OcservUserID)
		if err != nil {
			continue
		}
		title := user.Username
		row1 := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnUsage), cbAccountUsage+strconv.FormatUint(uint64(a.ID), 10)),
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnRenew), cbAccountRenew+strconv.FormatUint(uint64(a.ID), 10)),
		)
		row2 := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.T(lang, i18n.BtnRemove), cbAccountRemove+strconv.FormatUint(uint64(a.ID), 10)),
		)
		markup := tgbotapi.NewInlineKeyboardMarkup(row1, row2)
		h.sendKB(chatID, "• "+title, markup)
	}
}

func (h *Hub) SendAccountUsage(ctx context.Context, chatID int64, accountID uint, lang string) {
	account, err := h.deps.Repo.AccountByID(ctx, accountID)
	if err != nil || account.ChatID != chatID {
		h.send(chatID, i18n.T(lang, i18n.NotLinked))
		return
	}
	user, err := h.deps.Repo.OcservUserByID(ctx, account.OcservUserID)
	if err != nil {
		h.send(chatID, i18n.T(lang, i18n.NotLinked))
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
	msg := i18n.T(lang, i18n.UsageText, user.Username, status, user.TrafficSize, rxGB, txGB, expires)
	h.send(chatID, msg)
}

func (h *Hub) RemoveAccount(ctx context.Context, chatID int64, accountID uint) {
	account, err := h.deps.Repo.AccountByID(ctx, accountID)
	lang := h.LanguageFor(ctx, chatID)
	if err != nil || account.ChatID != chatID {
		h.send(chatID, i18n.T(lang, i18n.NotLinked))
		return
	}
	if err := h.deps.Repo.DeleteAccount(ctx, accountID); err != nil {
		logger.Warn("telegram_bot: failed to delete account: %v", err)
		h.send(chatID, i18n.T(lang, i18n.NotLinked))
		return
	}
	h.send(chatID, i18n.T(lang, i18n.AccountRemoved))
}

// =============================================================================
// New / Renew flows
// =============================================================================

func (h *Hub) StartNewOrder(ctx context.Context, chatID int64) {
	lang := h.LanguageFor(ctx, chatID)

	pending, err := h.deps.Repo.PendingByChat(ctx, chatID)
	if err == nil && pending != nil {
		h.send(chatID, i18n.T(lang, i18n.RequestExists))
		return
	}

	h.deps.Sessions.Set(chatID, &session.Session{State: session.WaitingUsernameForNew})
	h.send(chatID, i18n.T(lang, i18n.AskUsernameNew))
}

func (h *Hub) StartRenewForAccount(ctx context.Context, chatID int64, accountID uint) {
	lang := h.LanguageFor(ctx, chatID)
	account, err := h.deps.Repo.AccountByID(ctx, accountID)
	if err != nil || account.ChatID != chatID {
		h.send(chatID, i18n.T(lang, i18n.NotLinked))
		return
	}
	pending, err := h.deps.Repo.PendingByChat(ctx, chatID)
	if err == nil && pending != nil {
		h.send(chatID, i18n.T(lang, i18n.RequestExists))
		return
	}
	h.deps.Sessions.Set(chatID, &session.Session{
		State:          session.WaitingPackageForRenew,
		BufferTargetID: account.OcservUserID,
	})
	h.sendPackages(ctx, chatID, lang, cbPickPackageRenew)
}

func (h *Hub) sendPackages(ctx context.Context, chatID int64, lang, prefix string) {
	packages, err := h.deps.Repo.ActivePackages(ctx)
	if err != nil || len(packages) == 0 {
		h.send(chatID, i18n.T(lang, i18n.NoPackages))
		h.SendMainMenu(ctx, chatID, lang)
		return
	}
	h.sendKB(chatID, i18n.T(lang, i18n.PickPackage), packageKeyboard(packages, prefix))
}

func (h *Hub) PickedPackageNew(ctx context.Context, chatID int64, packageID uint) {
	sess := h.deps.Sessions.Get(chatID)
	lang := h.LanguageFor(ctx, chatID)
	if sess.State != session.WaitingPackageForNew {
		h.send(chatID, i18n.T(lang, i18n.SessionTimedOut))
		return
	}
	sess.BufferPackage = packageID
	sess.State = session.WaitingNoteForNew
	h.deps.Sessions.Set(chatID, sess)
	h.send(chatID, i18n.T(lang, i18n.AskMessage))
}

func (h *Hub) PickedPackageRenew(ctx context.Context, chatID int64, packageID uint) {
	sess := h.deps.Sessions.Get(chatID)
	lang := h.LanguageFor(ctx, chatID)
	if sess.State != session.WaitingPackageForRenew {
		h.send(chatID, i18n.T(lang, i18n.SessionTimedOut))
		return
	}
	sess.BufferPackage = packageID
	sess.State = session.WaitingNoteForRenew
	h.deps.Sessions.Set(chatID, sess)
	h.send(chatID, i18n.T(lang, i18n.AskMessage))
}

func (h *Hub) finalizeNewRequest(ctx context.Context, chatID int64, sess *session.Session, note, lang string) {
	pkgID := sess.BufferPackage
	desired := sess.BufferDesired

	req := &models.TelegramRequest{
		ChatID:           chatID,
		Type:             models.TelegramRequestTypeNew,
		PackageID:        ptrUint(pkgID),
		DesiredUsername:  desired,
		Status:           models.TelegramRequestStatusPending,
		UserMessage:      note,
	}
	created, err := h.deps.Repo.CreateRequest(ctx, req)
	if err != nil {
		logger.Warn("telegram_bot: failed to create request: %v", err)
		h.send(chatID, i18n.T(lang, i18n.UnknownCommand))
		return
	}
	h.deps.Sessions.Reset(chatID)
	h.send(chatID, i18n.T(lang, i18n.RequestCreated))

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
