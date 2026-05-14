package telegram

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/database"
	tg18n "github.com/mmtaee/ocserv-dashboard/api/internal/services/telegram/i18n"
)

const (
	telegramAPIBase     = "https://api.telegram.org"
	telegramHTTPTimeout = 8 * time.Second
)

func receiptStorageRoot() string {
	if d := strings.TrimSpace(os.Getenv("TELEGRAM_RECEIPTS_DIR")); d != "" {
		return filepath.Clean(d)
	}
	return "/opt/ocserv_dashboard/uploads/receipts"
}

func defaultNotifyLang(settings *models.TelegramSettings) string {
	if settings != nil && strings.TrimSpace(settings.DefaultLanguage) != "" {
		return settings.DefaultLanguage
	}
	return models.TelegramLanguageEN
}

type Controller struct {
	request        request.CustomRequestInterface
	repo           repository.TelegramRepositoryInterface
	ocservUserRepo repository.OcservUserRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request:        request.NewCustomRequest(),
		repo:           repository.NewTelegramRepository(),
		ocservUserRepo: repository.NewtOcservUserRepository(),
	}
}

// =============================================================================
// Settings
// =============================================================================

func (ctl *Controller) GetSettings(c echo.Context) error {
	s, err := ctl.repo.Settings(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, settingsToResponse(s))
}

func (ctl *Controller) UpdateSettings(c echo.Context) error {
	var data PatchSettingsData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	updates := map[string]interface{}{}
	if data.Enabled != nil {
		updates["enabled"] = *data.Enabled
	}
	if data.BotToken != nil {
		updates["bot_token"] = *data.BotToken
		// reset cached username; the bot service will refresh it via getMe.
		updates["bot_username"] = ""
	}
	if data.AdminChatID != nil {
		updates["admin_chat_id"] = *data.AdminChatID
	}
	if data.LowQuotaThresholdMB != nil {
		updates["low_quota_threshold_mb"] = *data.LowQuotaThresholdMB
	}
	if data.DefaultLanguage != nil {
		updates["default_language"] = *data.DefaultLanguage
	}
	if data.OcservHost != nil {
		updates["ocserv_host"] = *data.OcservHost
	}
	if data.CardNumber != nil {
		updates["card_number"] = *data.CardNumber
	}
	if data.CardHolder != nil {
		updates["card_holder"] = *data.CardHolder
	}
	if data.SupportUsername != nil {
		updates["support_username"] = strings.TrimPrefix(strings.TrimSpace(*data.SupportUsername), "@")
	}
	if len(updates) == 0 {
		return ctl.request.BadRequest(c, errors.New("no fields to update"))
	}

	s, err := ctl.repo.UpdateSettings(c.Request().Context(), updates)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	// best-effort: refresh bot username from telegram getMe
	if data.BotToken != nil && *data.BotToken != "" {
		if uname, err := fetchBotUsername(*data.BotToken); err == nil && uname != "" {
			_, _ = ctl.repo.UpdateSettings(c.Request().Context(), map[string]interface{}{
				"bot_username": uname,
			})
			s.BotUsername = uname
		}
	}

	return c.JSON(http.StatusOK, settingsToResponse(s))
}

func (ctl *Controller) Test(c echo.Context) error {
	var data TestData
	_ = ctl.request.DoValidate(c, &data)

	s, err := ctl.repo.Settings(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if s.BotToken == "" {
		return ctl.request.BadRequest(c, errors.New("bot token is not set"))
	}
	if s.AdminChatID == 0 {
		return ctl.request.BadRequest(c, errors.New("admin chat id is not set"))
	}

	msg := data.Message
	if msg == "" {
		msg = "Test message from your dashboard"
	}

	if err := sendTelegramMessage(s.BotToken, s.AdminChatID, msg); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// =============================================================================
// Packages
// =============================================================================

func (ctl *Controller) ListPackages(c echo.Context) error {
	includeInactive := c.QueryParam("include_inactive") == "true"
	packages, err := ctl.repo.Packages(c.Request().Context(), includeInactive)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, packages)
}

func (ctl *Controller) CreatePackage(c echo.Context) error {
	var data CreatePackageData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	pkg := &models.TelegramPackage{
		Title:         data.Title,
		Days:          data.Days,
		TrafficSizeGB: data.TrafficSizeGB,
		TrafficType:   data.TrafficType,
		PriceText:     data.PriceText,
		IsActive:      data.IsActive,
	}
	created, err := ctl.repo.CreatePackage(c.Request().Context(), pkg)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusCreated, created)
}

func (ctl *Controller) UpdatePackage(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	var data PatchPackageData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	updates := map[string]interface{}{}
	if data.Title != nil {
		updates["title"] = *data.Title
	}
	if data.Days != nil {
		updates["days"] = *data.Days
	}
	if data.TrafficSizeGB != nil {
		updates["traffic_size_gb"] = *data.TrafficSizeGB
	}
	if data.TrafficType != nil {
		updates["traffic_type"] = *data.TrafficType
	}
	if data.PriceText != nil {
		updates["price_text"] = *data.PriceText
	}
	if data.IsActive != nil {
		updates["is_active"] = *data.IsActive
	}
	if len(updates) == 0 {
		return ctl.request.BadRequest(c, errors.New("no fields to update"))
	}

	pkg, err := ctl.repo.UpdatePackage(c.Request().Context(), uint(id), updates)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, pkg)
}

func (ctl *Controller) DeletePackage(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.repo.DeletePackage(c.Request().Context(), uint(id)); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

// =============================================================================
// Requests
// =============================================================================

func (ctl *Controller) ListRequests(c echo.Context) error {
	pagination := ctl.request.Pagination(c)
	// Default listing matches historical behavior (newest first) when the client omits order/sort.
	q := c.Request().URL.Query()
	if q.Get("order") == "" {
		pagination.Order = "created_at"
	}
	if q.Get("sort") == "" {
		pagination.Sort = "DESC"
	}
	status := c.QueryParam("status")
	requestType := c.QueryParam("type")

	requests, total, err := ctl.repo.Requests(c.Request().Context(), pagination, status, requestType)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, RequestsResponse{
		Meta: request.Meta{
			Page:         pagination.Page,
			PageSize:     pagination.PageSize,
			TotalRecords: total,
		},
		Result: requests,
	})
}

func (ctl *Controller) GetRequest(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	req, err := ctl.repo.RequestByID(c.Request().Context(), uint(id))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, req)
}

func (ctl *Controller) GetReceipt(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	req, err := ctl.repo.RequestByID(c.Request().Context(), uint(id))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if req.ReceiptFilePath == "" {
		return ctl.request.BadRequest(c, errors.New("no receipt uploaded"))
	}
	if _, err := os.Stat(req.ReceiptFilePath); err != nil {
		return ctl.request.BadRequest(c, errors.New("receipt file not found on disk"))
	}
	return c.File(req.ReceiptFilePath)
}

func (ctl *Controller) DeleteRequest(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.repo.DeleteRequest(c.Request().Context(), uint(id)); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (ctl *Controller) Approve(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	var data ApproveData
	_ = ctl.request.DoValidate(c, &data)

	req, err := ctl.repo.RequestByID(c.Request().Context(), uint(id))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if req.Status != models.TelegramRequestStatusPending {
		return ctl.request.BadRequest(c, fmt.Errorf("only pending requests can be approved (current=%s)", req.Status))
	}

	var note *string
	if data.AdminNote != "" {
		note = &data.AdminNote
	}
	updated, err := ctl.repo.UpdateRequestStatus(c.Request().Context(), uint(id), models.TelegramRequestStatusAwaitingPayment, note)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	go ctl.notifyAwaitingPayment(updated, &awaitingPaymentOpts{
		CardNumber:  data.CardNumber,
		CardHolder:  data.CardHolder,
		ReplyToUser: data.ReplyToUser,
	})
	return c.JSON(http.StatusOK, updated)
}

func (ctl *Controller) Reject(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	var data RejectData
	_ = ctl.request.DoValidate(c, &data)

	req, err := ctl.repo.RequestByID(c.Request().Context(), uint(id))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if req.Status == models.TelegramRequestStatusDelivered {
		return ctl.request.BadRequest(c, errors.New("cannot reject a delivered request"))
	}

	var note *string
	if data.AdminNote != "" {
		note = &data.AdminNote
	}
	updated, err := ctl.repo.UpdateRequestStatus(c.Request().Context(), uint(id), models.TelegramRequestStatusRejected, note)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	go ctl.notifyRejected(updated)
	return c.JSON(http.StatusOK, updated)
}

func (ctl *Controller) ConfirmPayment(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	var data ConfirmPaymentData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	req, err := ctl.repo.RequestByID(c.Request().Context(), uint(id))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if req.Status != models.TelegramRequestStatusPaymentUploaded {
		return ctl.request.BadRequest(c, fmt.Errorf("payment can only be confirmed after receipt upload (current=%s)", req.Status))
	}
	if req.PackageID == nil {
		return ctl.request.BadRequest(c, errors.New("request has no package"))
	}

	pkg, err := ctl.repo.PackageByID(c.Request().Context(), *req.PackageID)
	if err != nil {
		return ctl.request.BadRequest(c, fmt.Errorf("package not found: %w", err))
	}

	settings, err := ctl.repo.Settings(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	switch req.Type {
	case models.TelegramRequestTypeNew:
		return ctl.deliverNewAccount(c, req, pkg, settings, &data)
	case models.TelegramRequestTypeRenew:
		return ctl.deliverRenewal(c, req, pkg, settings, &data)
	default:
		return ctl.request.BadRequest(c, fmt.Errorf("unknown request type: %s", req.Type))
	}
}

// =============================================================================
// Linked accounts
// =============================================================================

func (ctl *Controller) AccountsForOcservUser(c echo.Context) error {
	uid := c.QueryParam("ocserv_user_uid")
	if uid == "" {
		return ctl.request.BadRequest(c, errors.New("ocserv_user_uid query parameter is required"))
	}
	user, err := ctl.ocservUserRepo.GetByUID(c.Request().Context(), uid)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	accounts, err := ctl.repo.AccountsForOcservUser(c.Request().Context(), user.ID)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, accounts)
}

func (ctl *Controller) DeleteAccount(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.repo.DeleteAccount(c.Request().Context(), uint(id)); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

// =============================================================================
// Internal helpers
// =============================================================================

func settingsToResponse(s *models.TelegramSettings) SettingsResponse {
	return SettingsResponse{
		Enabled:             s.Enabled,
		BotToken:            s.BotToken,
		BotUsername:         s.BotUsername,
		AdminChatID:         s.AdminChatID,
		LowQuotaThresholdMB: s.LowQuotaThresholdMB,
		DefaultLanguage:     s.DefaultLanguage,
		OcservHost:          s.OcservHost,
		CardNumber:          s.CardNumber,
		CardHolder:          s.CardHolder,
		SupportUsername:     s.SupportUsername,
	}
}

func (ctl *Controller) deliverNewAccount(
	c echo.Context,
	req *models.TelegramRequest,
	pkg *models.TelegramPackage,
	settings *models.TelegramSettings,
	data *ConfirmPaymentData,
) error {
	username := data.OverrideUsername
	if username == "" {
		username = req.DesiredUsername
	}
	if username == "" {
		username = generateUsername()
	}

	password := data.OverridePassword
	if password == "" {
		password = generatePassword()
	}

	owner := data.Owner
	if owner == "" {
		owner = "telegram"
	}
	group := data.Group
	if group == "" {
		group = "defaults"
	}

	expireAt := time.Now().AddDate(0, 0, pkg.Days)

	user := &models.OcservUser{
		Owner:       owner,
		Group:       group,
		Username:    username,
		Password:    password,
		ExpireAt:    &expireAt,
		TrafficType: pkg.TrafficType,
		TrafficSize: pkg.TrafficSizeGB,
		Description: fmt.Sprintf("created via telegram bot (request #%d)", req.ID),
	}

	created, err := ctl.ocservUserRepo.Create(c.Request().Context(), user)
	if err != nil {
		return ctl.request.BadRequest(c, fmt.Errorf("failed to create ocserv user: %w", err))
	}

	// link telegram account to the new ocserv user
	if err := linkTelegramAccount(c.Request().Context(), req.ChatID, req.TelegramUsername, settings.DefaultLanguage, created.ID); err != nil {
		// non-fatal, just log via admin note path
		_ = err
	}

	if data.AdminNote != "" {
		_, _ = ctl.repo.UpdateRequestStatus(c.Request().Context(), req.ID, models.TelegramRequestStatusPaymentUploaded, &data.AdminNote)
	}
	if err := ctl.repo.MarkDelivered(c.Request().Context(), req.ID, &created.ID); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	go ctl.notifyDelivery(req.ChatID, settings, formatNewAccountMessage(settings, created, password, expireAt))
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "delivered",
		"username": created.Username,
	})
}

func (ctl *Controller) deliverRenewal(
	c echo.Context,
	req *models.TelegramRequest,
	pkg *models.TelegramPackage,
	settings *models.TelegramSettings,
	data *ConfirmPaymentData,
) error {
	if req.TargetOcservID == nil {
		return ctl.request.BadRequest(c, errors.New("renewal request has no target user"))
	}

	user, err := ctl.findOcservUserByID(c.Request().Context(), *req.TargetOcservID)
	if err != nil {
		return ctl.request.BadRequest(c, fmt.Errorf("target ocserv user not found: %w", err))
	}

	now := time.Now()
	base := now
	if user.ExpireAt != nil && user.ExpireAt.After(now) {
		base = *user.ExpireAt
	}
	newExpire := base.AddDate(0, 0, pkg.Days)

	user.ExpireAt = &newExpire
	user.DeactivatedAt = nil
	user.IsLocked = false
	user.Rx = 0
	user.Tx = 0
	user.TrafficType = pkg.TrafficType
	user.TrafficSize = pkg.TrafficSizeGB

	if _, err := ctl.ocservUserRepo.Update(c.Request().Context(), user); err != nil {
		return ctl.request.BadRequest(c, fmt.Errorf("failed to renew ocserv user: %w", err))
	}

	if data.AdminNote != "" {
		_, _ = ctl.repo.UpdateRequestStatus(c.Request().Context(), req.ID, models.TelegramRequestStatusPaymentUploaded, &data.AdminNote)
	}
	if err := ctl.repo.MarkDelivered(c.Request().Context(), req.ID, &user.ID); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	go ctl.notifyDelivery(req.ChatID, settings, formatRenewalMessage(settings, user, newExpire))
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "delivered",
		"username": user.Username,
	})
}

func (ctl *Controller) findOcservUserByID(ctx context.Context, id uint) (*models.OcservUser, error) {
	var user models.OcservUser
	if err := database.GetConnection().
		WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// =============================================================================
// Notifications (best-effort fire-and-forget)
// =============================================================================

// awaitingPaymentOpts carries optional payment instructions chosen when approving a request.
type awaitingPaymentOpts struct {
	CardNumber  string
	CardHolder  string
	ReplyToUser string
}

func (ctl *Controller) notifyAwaitingPayment(req *models.TelegramRequest, opts *awaitingPaymentOpts) {
	settings, err := ctl.repo.Settings(context.Background())
	if err != nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	lang := ctl.resolveNotifyLang(context.Background(), req.ChatID, settings)
	var pkg *models.TelegramPackage
	if req.PackageID != nil && *req.PackageID > 0 {
		if p, err := ctl.repo.PackageByID(context.Background(), *req.PackageID); err == nil {
			pkg = p
		}
	}
	msg := formatAwaitingPaymentMessage(lang, settings, opts, pkg)
	msgID, err := sendTelegramHTMLMessageWithID(settings.BotToken, req.ChatID, msg)
	if err != nil || msgID <= 0 {
		return
	}
	_ = ctl.repo.SetAwaitingPaymentMessageID(context.Background(), req.ID, msgID)
}

func (ctl *Controller) resolveNotifyLang(ctx context.Context, chatID int64, settings *models.TelegramSettings) string {
	if l, err := ctl.repo.PreferredLanguageForChat(ctx, chatID); err == nil && strings.TrimSpace(l) != "" {
		return strings.TrimSpace(l)
	}
	if settings != nil && settings.DefaultLanguage != "" {
		return settings.DefaultLanguage
	}
	return models.TelegramLanguageEN
}

func (ctl *Controller) notifyRejected(req *models.TelegramRequest) {
	settings, err := ctl.repo.Settings(context.Background())
	if err != nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	if req.AwaitingPaymentMessageID != nil && *req.AwaitingPaymentMessageID > 0 {
		deleteTelegramMessage(settings.BotToken, req.ChatID, *req.AwaitingPaymentMessageID)
		_ = ctl.repo.ClearAwaitingPaymentMessageID(context.Background(), req.ID)
	}
	msg := formatRejectedMessage(settings, req.AdminNote)
	_ = sendTelegramHTMLMessage(settings.BotToken, req.ChatID, msg)
}

func (ctl *Controller) notifyDelivery(chatID int64, settings *models.TelegramSettings, message string) {
	if settings == nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	_ = sendTelegramHTMLMessage(settings.BotToken, chatID, message)
}

// packageSummaryBlock appends plan title, deposit hint (PriceText), duration and quota for the awaiting-payment notice.
func packageSummaryBlock(lang string, pkg *models.TelegramPackage) string {
	if pkg == nil {
		return ""
	}
	title := htmlEsc(pkg.Title)
	price := strings.TrimSpace(pkg.PriceText)
	if price == "" {
		price = tg18n.T(lang, "pkg_price_placeholder")
	} else {
		price = htmlEsc(price)
	}
	return tg18n.T(lang, "pkg_summary", title, price, pkg.Days, pkg.TrafficSizeGB)
}

func formatAwaitingPaymentMessage(lang string, settings *models.TelegramSettings, opts *awaitingPaymentOpts, pkg *models.TelegramPackage) string {
	cardNum := ""
	cardHold := ""
	if settings != nil {
		cardNum = settings.CardNumber
		cardHold = settings.CardHolder
	}
	if opts != nil {
		if strings.TrimSpace(opts.CardNumber) != "" {
			cardNum = strings.TrimSpace(opts.CardNumber)
		}
		if strings.TrimSpace(opts.CardHolder) != "" {
			cardHold = strings.TrimSpace(opts.CardHolder)
		}
	}

	cardLine := ""
	if cardNum != "" {
		holder := cardHold
		if holder == "" {
			holder = "—"
		}
		cardLine = tg18n.T(lang, "awaiting_card_line", htmlEsc(cardNum), htmlEsc(holder))
	}

	replyBlock := ""
	if opts != nil && strings.TrimSpace(opts.ReplyToUser) != "" {
		reply := strings.TrimSpace(opts.ReplyToUser)
		replyBlock = tg18n.T(lang, "awaiting_reply_prefix") + htmlEsc(reply)
	}

	missingCard := ""
	if cardNum == "" {
		missingCard = tg18n.T(lang, "awaiting_missing_card")
	}

	receiptLine := tg18n.T(lang, "awaiting_receipt_line")

	pkgBlock := packageSummaryBlock(lang, pkg)
	support := supportLine(settings)
	intro := tg18n.T(lang, "awaiting_intro")
	closeTag := tg18n.T(lang, "awaiting_close")
	return intro + pkgBlock + replyBlock + cardLine + receiptLine + missingCard + closeTag + support
}

func formatRejectedMessage(settings *models.TelegramSettings, adminNote string) string {
	lang := defaultNotifyLang(settings)
	msg := tg18n.T(lang, "rejected_title")
	if adminNote != "" {
		msg += tg18n.T(lang, "rejected_reason", htmlEsc(adminNote))
	}
	msg += tg18n.T(lang, "rejected_close")
	return msg
}

func formatNewAccountMessage(settings *models.TelegramSettings, user *models.OcservUser, plainPassword string, expireAt time.Time) string {
	host := settings.OcservHost
	if host == "" {
		host = "—"
	}
	support := supportLine(settings)
	lang := defaultNotifyLang(settings)
	return tg18n.T(lang, "new_account",
		htmlEsc(host), htmlEsc(user.Username), htmlEsc(plainPassword),
		expireAt.Format("2006-01-02"), user.TrafficSize, support,
	)
}

// supportLine returns the localized "for help, contact @support" line, or
// an empty string if no support_username is configured. The leading "\n\n"
// is included so callers can append unconditionally.
func supportLine(settings *models.TelegramSettings) string {
	if settings == nil {
		return ""
	}
	handle := strings.TrimPrefix(strings.TrimSpace(settings.SupportUsername), "@")
	if handle == "" {
		return ""
	}
	link := `<a href="https://t.me/` + handle + `">@` + handle + `</a>`
	lang := defaultNotifyLang(settings)
	return tg18n.T(lang, "support_suffix", link)
}

func formatRenewalMessage(settings *models.TelegramSettings, user *models.OcservUser, newExpire time.Time) string {
	support := supportLine(settings)
	lang := defaultNotifyLang(settings)
	return tg18n.T(lang, "renewal",
		htmlEsc(user.Username), newExpire.Format("2006-01-02"), user.TrafficSize, support,
	)
}

// =============================================================================
// Telegram low-level helpers
// =============================================================================

func fetchBotUsername(token string) (string, error) {
	endpoint := fmt.Sprintf("%s/bot%s/getMe", telegramAPIBase, token)
	client := &http.Client{Timeout: telegramHTTPTimeout}
	resp, err := client.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("telegram getMe returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return parseUsernameFromGetMe(body), nil
}

func parseUsernameFromGetMe(body []byte) string {
	const key = `"username":"`
	idx := -1
	for i := 0; i+len(key) < len(body); i++ {
		match := true
		for j := 0; j < len(key); j++ {
			if body[i+j] != key[j] {
				match = false
				break
			}
		}
		if match {
			idx = i + len(key)
			break
		}
	}
	if idx == -1 {
		return ""
	}
	end := idx
	for end < len(body) && body[end] != '"' {
		end++
	}
	if end >= len(body) {
		return ""
	}
	return string(body[idx:end])
}

// htmlEsc escapes special HTML characters for safe inclusion in HTML parse-mode messages.
func htmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func sendTelegramMessage(token string, chatID int64, text string) error {
	_, err := sendTelegramMessageWithMode(token, chatID, text, "")
	return err
}

func sendTelegramHTMLMessage(token string, chatID int64, text string) error {
	_, err := sendTelegramHTMLMessageWithID(token, chatID, text)
	return err
}

func sendTelegramHTMLMessageWithID(token string, chatID int64, text string) (int64, error) {
	return sendTelegramMessageWithMode(token, chatID, text, "HTML")
}

func sendTelegramMessageWithMode(token string, chatID int64, text, parseMode string) (int64, error) {
	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", telegramAPIBase, token)
	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("text", text)
	form.Set("disable_web_page_preview", "true")
	if parseMode != "" {
		form.Set("parse_mode", parseMode)
	}
	client := &http.Client{Timeout: telegramHTTPTimeout}
	resp, err := client.PostForm(endpoint, form)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("telegram sendMessage status=%d body=%s", resp.StatusCode, string(body))
	}
	var envelope struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil || !envelope.OK {
		return 0, fmt.Errorf("telegram sendMessage: invalid or unsuccessful response: %s", string(body))
	}
	return envelope.Result.MessageID, nil
}

func deleteTelegramMessage(token string, chatID, messageID int64) {
	endpoint := fmt.Sprintf("%s/bot%s/deleteMessage", telegramAPIBase, token)
	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("message_id", strconv.FormatInt(messageID, 10))
	client := &http.Client{Timeout: telegramHTTPTimeout}
	resp, err := client.PostForm(endpoint, form)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
}

func generateUsername() string {
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	return "tg_" + hex.EncodeToString(buf)
}

func generatePassword() string {
	buf := make([]byte, 6)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func linkTelegramAccount(ctx context.Context, chatID int64, username, language string, ocservUserID uint) error {
	if language == "" {
		language = models.TelegramLanguageEN
	}
	account := &models.TelegramAccount{
		ChatID:           chatID,
		TelegramUsername: username,
		Language:         language,
		OcservUserID:     ocservUserID,
	}
	return database.GetConnection().
		WithContext(ctx).
		Where("chat_id = ? AND ocserv_user_id = ?", chatID, ocservUserID).
		FirstOrCreate(account).Error
}

// EnsureReceiptDir is invoked at startup to make sure the receipt storage directory exists.
func EnsureReceiptDir() error {
	return os.MkdirAll(filepath.Clean(receiptStorageRoot()), 0o750)
}
