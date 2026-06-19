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

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
	tg18n "github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/service/telegram/i18n"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
)

type TelegramController struct {
	telegramUseCase usecase.TelegramUseCase
	userUseCase     usecase.OcservUserUseCase
	req             *request.Request
	validator       *request.Validator
	userRepo        repository.OcservUserRepository
	config          config.TelegramConfig
}

func NewTelegramController(telegramUseCase usecase.TelegramUseCase, userUseCase usecase.OcservUserUseCase, userRepo repository.OcservUserRepository, cfg config.TelegramConfig) *TelegramController {
	tg18n.Init()

	if err := ensureReceiptDir(cfg); err != nil {
		logger.Warn("telegram: failed to ensure receipt directory: %v", err)
	}

	return &TelegramController{
		telegramUseCase: telegramUseCase,
		userUseCase:     userUseCase,
		req:             &request.Request{},
		validator:       request.NewValidator(),
		userRepo:        userRepo,
		config:          cfg,
	}
}

const (
	telegramHTTPTimeout = 8 * time.Second
)

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

func (ctrl *TelegramController) GetSettings(c echo.Context) error {
	s, err := ctrl.telegramUseCase.GetSettings()
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, settingsToResponse(s))
}

func (ctrl *TelegramController) UpdateSettings(c echo.Context) error {
	var data PatchSettingsData
	if err := ctrl.validator.Validate(c, &data); err != nil {
		return err
	}

	updates := map[string]interface{}{}
	if data.Enabled != nil {
		updates["enabled"] = *data.Enabled
	}
	if data.BotToken != nil {
		updates["bot_token"] = *data.BotToken
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
		return ctrl.req.BadRequest(c, errors.New("no fields to update"))
	}

	s, err := ctrl.telegramUseCase.UpdateSettings(updates)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}

	if data.BotToken != nil && *data.BotToken != "" {
		if uname, err := ctrl.fetchBotUsername(*data.BotToken); err == nil && uname != "" {
			_, _ = ctrl.telegramUseCase.UpdateSettings(map[string]interface{}{
				"bot_username": uname,
			})
			s.BotUsername = uname
		}
	}

	return c.JSON(http.StatusOK, settingsToResponse(s))
}

func (ctrl *TelegramController) Test(c echo.Context) error {
	var data TestData
	_ = ctrl.validator.Validate(c, &data)

	s, err := ctrl.telegramUseCase.GetSettings()
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	if s.BotToken == "" {
		return ctrl.req.BadRequest(c, errors.New("bot token is not set"))
	}
	if s.AdminChatID == 0 {
		return ctrl.req.BadRequest(c, errors.New("admin chat id is not set"))
	}

	msg := data.Message
	if msg == "" {
		msg = "Test message from your dashboard"
	}

	if err := ctrl.sendTelegramMessage(s.BotToken, s.AdminChatID, msg); err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (ctrl *TelegramController) ListPackages(c echo.Context) error {
	includeInactive := c.QueryParam("include_inactive") == "true"
	packages, err := ctrl.telegramUseCase.GetPackages(includeInactive)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, packages)
}

func (ctrl *TelegramController) CreatePackage(c echo.Context) error {
	var data CreatePackageData
	if err := ctrl.validator.Validate(c, &data); err != nil {
		return err
	}
	pkg := &models.TelegramPackage{
		Title:         data.Title,
		Days:          data.Days,
		TrafficSizeGB: data.TrafficSizeGB,
		TrafficType:   data.TrafficType,
		PriceText:     data.PriceText,
		IsActive:      data.IsActive,
	}
	created, err := ctrl.telegramUseCase.CreatePackage(pkg)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusCreated, created)
}

func (ctrl *TelegramController) UpdatePackage(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	var data PatchPackageData
	if err := ctrl.validator.Validate(c, &data); err != nil {
		return err
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
		return ctrl.req.BadRequest(c, errors.New("no fields to update"))
	}

	pkg, err := ctrl.telegramUseCase.UpdatePackage(uint(id), updates)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, pkg)
}

func (ctrl *TelegramController) DeletePackage(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	if err := ctrl.telegramUseCase.DeletePackage(uint(id)); err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

func (ctrl *TelegramController) ListRequests(c echo.Context) error {
	pagination := ctrl.req.Pagination(c)
	q := c.Request().URL.Query()
	if q.Get("order") == "" {
		pagination.Order = "created_at"
	}
	if q.Get("sort") == "" {
		pagination.Sort = "desc"
	}
	status := c.QueryParam("status")
	requestType := c.QueryParam("type")

	requests, total, err := ctrl.telegramUseCase.GetRequests(pagination.Page, pagination.PageSize, pagination.Order, pagination.Sort, status, requestType)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, RequestsResponse{
		Meta:   pagination,
		Result: requests,
	})
}

func (ctrl *TelegramController) GetRequest(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	req, err := ctrl.telegramUseCase.GetRequestByID(uint(id))
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, req)
}

func (ctrl *TelegramController) GetReceipt(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	req, err := ctrl.telegramUseCase.GetRequestByID(uint(id))
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	if req.ReceiptFilePath == "" {
		return ctrl.req.BadRequest(c, errors.New("no receipt uploaded"))
	}
	if _, err := os.Stat(req.ReceiptFilePath); err != nil {
		return ctrl.req.BadRequest(c, errors.New("receipt file not found on disk"))
	}
	return c.File(req.ReceiptFilePath)
}

func (ctrl *TelegramController) DeleteRequest(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	if err := ctrl.telegramUseCase.DeleteRequest(uint(id)); err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (ctrl *TelegramController) Approve(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	var data ApproveData
	_ = ctrl.validator.Validate(c, &data)

	req, err := ctrl.telegramUseCase.GetRequestByID(uint(id))
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	if req.Status != models.TelegramRequestStatusPending {
		return ctrl.req.BadRequest(c, fmt.Errorf("only pending requests can be approved (current=%s)", req.Status))
	}

	var note *string
	if data.AdminNote != "" {
		note = &data.AdminNote
	}
	updated, err := ctrl.telegramUseCase.UpdateRequestStatus(uint(id), models.TelegramRequestStatusAwaitingPayment, note)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}

	go ctrl.notifyAwaitingPayment(updated, &awaitingPaymentOpts{
		CardNumber:  data.CardNumber,
		CardHolder:  data.CardHolder,
		ReplyToUser: data.ReplyToUser,
	})
	return c.JSON(http.StatusOK, updated)
}

func (ctrl *TelegramController) Reject(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	var data RejectData
	_ = ctrl.validator.Validate(c, &data)

	req, err := ctrl.telegramUseCase.GetRequestByID(uint(id))
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	if req.Status == models.TelegramRequestStatusDelivered {
		return ctrl.req.BadRequest(c, errors.New("cannot reject a delivered request"))
	}

	var note *string
	if data.AdminNote != "" {
		note = &data.AdminNote
	}
	updated, err := ctrl.telegramUseCase.UpdateRequestStatus(uint(id), models.TelegramRequestStatusRejected, note)
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}

	go ctrl.notifyRejected(updated)
	return c.JSON(http.StatusOK, updated)
}

func (ctrl *TelegramController) ConfirmPayment(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	var data ConfirmPaymentData
	if err := ctrl.validator.Validate(c, &data); err != nil {
		return err
	}

	req, err := ctrl.telegramUseCase.GetRequestByID(uint(id))
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	if req.Status != models.TelegramRequestStatusPaymentUploaded {
		return ctrl.req.BadRequest(c, fmt.Errorf("payment can only be confirmed after receipt upload (current=%s)", req.Status))
	}
	if req.PackageID == nil {
		return ctrl.req.BadRequest(c, errors.New("request has no package"))
	}

	pkg, err := ctrl.telegramUseCase.GetPackageByID(*req.PackageID)
	if err != nil {
		return ctrl.req.BadRequest(c, fmt.Errorf("package not found: %w", err))
	}

	settings, err := ctrl.telegramUseCase.GetSettings()
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}

	switch req.Type {
	case models.TelegramRequestTypeNew:
		return ctrl.deliverNewAccount(c, req, pkg, settings, &data)
	case models.TelegramRequestTypeRenew:
		return ctrl.deliverRenewal(c, req, pkg, settings, &data)
	default:
		return ctrl.req.BadRequest(c, fmt.Errorf("unknown request type: %s", req.Type))
	}
}

func (ctrl *TelegramController) AccountsForOcservUser(c echo.Context) error {
	userIDStr := c.QueryParam("ocserv_user_id")
	if userIDStr == "" {
		return ctrl.req.BadRequest(c, errors.New("ocserv_user_id query parameter is required"))
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	accounts, err := ctrl.telegramUseCase.GetAccountsForOcservUser(uint(userID))
	if err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusOK, accounts)
}

func (ctrl *TelegramController) DeleteAccount(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}
	if err := ctrl.telegramUseCase.DeleteAccount(uint(id)); err != nil {
		return ctrl.req.InternalServerError(c, err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

func (ctrl *TelegramController) deliverNewAccount(
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

	user, err := ctrl.userUseCase.CreateUser(username, password, group, pkg.TrafficType, gigabytesToBytes(pkg.TrafficSizeGB), fmt.Sprintf("created via telegram bot (request #%d)", req.ID), nil, 1, &expireAt)
	if err != nil {
		return ctrl.req.InternalServerError(c, fmt.Errorf("failed to create ocserv user: %w", err))
	}

	if err := linkTelegramAccount(c.Request().Context(), req.ChatID, req.TelegramUsername, settings.DefaultLanguage, user.ID); err != nil {
	}

	if data.AdminNote != "" {
		note := data.AdminNote
		_, _ = ctrl.telegramUseCase.UpdateRequestStatus(req.ID, models.TelegramRequestStatusPaymentUploaded, &note)
	}
	if err := ctrl.telegramUseCase.MarkDelivered(req.ID, &user.ID); err != nil {
		return ctrl.req.InternalServerError(c, err)
	}

	go ctrl.notifyDelivery(req.ChatID, settings, formatNewAccountMessage(settings, user, password, expireAt))
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "delivered",
		"username": user.Username,
	})
}

func (ctrl *TelegramController) deliverRenewal(
	c echo.Context,
	req *models.TelegramRequest,
	pkg *models.TelegramPackage,
	settings *models.TelegramSettings,
	data *ConfirmPaymentData,
) error {
	if req.TargetOcservID == nil {
		return ctrl.req.BadRequest(c, errors.New("renewal request has no target user"))
	}

	user, err := ctrl.findOcservUserByID(c.Request().Context(), *req.TargetOcservID)
	if err != nil {
		return ctrl.req.BadRequest(c, fmt.Errorf("target ocserv user not found: %w", err))
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
	user.TrafficSize = gigabytesToBytes(pkg.TrafficSizeGB)

	if err := ctrl.userRepo.UpdateUnrestricted(user); err != nil {
		return ctrl.req.InternalServerError(c, fmt.Errorf("failed to renew ocserv user: %w", err))
	}

	if data.AdminNote != "" {
		note := data.AdminNote
		_, _ = ctrl.telegramUseCase.UpdateRequestStatus(req.ID, models.TelegramRequestStatusPaymentUploaded, &note)
	}
	if err := ctrl.telegramUseCase.MarkDelivered(req.ID, &user.ID); err != nil {
		return ctrl.req.InternalServerError(c, err)
	}

	go ctrl.notifyDelivery(req.ChatID, settings, formatRenewalMessage(settings, user, newExpire))
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "delivered",
		"username": user.Username,
	})
}

func (ctrl *TelegramController) findOcservUserByID(ctx context.Context, id uint) (*models.OcservUser, error) {
	return ctrl.userRepo.FindByIDUnrestricted(id)
}

type awaitingPaymentOpts struct {
	CardNumber  string
	CardHolder  string
	ReplyToUser string
}

func (ctrl *TelegramController) notifyAwaitingPayment(req *models.TelegramRequest, opts *awaitingPaymentOpts) {
	settings, err := ctrl.telegramUseCase.GetSettings()
	if err != nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	lang := ctrl.resolveNotifyLang(context.Background(), req.ChatID, settings)
	var pkg *models.TelegramPackage
	if req.PackageID != nil && *req.PackageID > 0 {
		if p, err := ctrl.telegramUseCase.GetPackageByID(*req.PackageID); err == nil {
			pkg = p
		}
	}
	msg := formatAwaitingPaymentMessage(lang, settings, opts, pkg)
	msgID, err := ctrl.sendTelegramHTMLMessageWithID(settings.BotToken, req.ChatID, msg)
	if err != nil || msgID <= 0 {
		return
	}
	_ = ctrl.telegramUseCase.SetAwaitingPaymentMessageID(req.ID, msgID)
}

func (ctrl *TelegramController) resolveNotifyLang(ctx context.Context, chatID int64, settings *models.TelegramSettings) string {
	if l, err := ctrl.telegramUseCase.PreferredLanguageForChat(chatID); err == nil && strings.TrimSpace(l) != "" {
		return strings.TrimSpace(l)
	}
	if settings != nil && settings.DefaultLanguage != "" {
		return settings.DefaultLanguage
	}
	return models.TelegramLanguageEN
}

func (ctrl *TelegramController) notifyRejected(req *models.TelegramRequest) {
	settings, err := ctrl.telegramUseCase.GetSettings()
	if err != nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	if req.AwaitingPaymentMessageID != nil && *req.AwaitingPaymentMessageID > 0 {
		ctrl.deleteTelegramMessage(settings.BotToken, req.ChatID, *req.AwaitingPaymentMessageID)
		_ = ctrl.telegramUseCase.ClearAwaitingPaymentMessageID(req.ID)
	}
	msg := formatRejectedMessage(settings, req.AdminNote)
	_ = ctrl.sendTelegramHTMLMessage(settings.BotToken, req.ChatID, msg)
}

func (ctrl *TelegramController) notifyDelivery(chatID int64, settings *models.TelegramSettings, message string) {
	if settings == nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	_ = ctrl.sendTelegramHTMLMessage(settings.BotToken, chatID, message)
}

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
		expireAt.Format("2006-01-02"), bytesToGigabytes(user.TrafficSize), support,
	)
}

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
		htmlEsc(user.Username), newExpire.Format("2006-01-02"), bytesToGigabytes(user.TrafficSize), support,
	)
}

func (ctrl *TelegramController) fetchBotUsername(token string) (string, error) {
	endpoint := fmt.Sprintf("%s/bot%s/getMe", ctrl.config.APIBase, token)
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

func htmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func (ctrl *TelegramController) sendTelegramMessage(token string, chatID int64, text string) error {
	_, err := ctrl.sendTelegramMessageWithMode(token, chatID, text, "")
	return err
}

func (ctrl *TelegramController) sendTelegramHTMLMessage(token string, chatID int64, text string) error {
	_, err := ctrl.sendTelegramHTMLMessageWithID(token, chatID, text)
	return err
}

func (ctrl *TelegramController) sendTelegramHTMLMessageWithID(token string, chatID int64, text string) (int64, error) {
	return ctrl.sendTelegramMessageWithMode(token, chatID, text, "HTML")
}

func (ctrl *TelegramController) sendTelegramMessageWithMode(token string, chatID int64, text, parseMode string) (int64, error) {
	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", ctrl.config.APIBase, token)
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

func (ctrl *TelegramController) deleteTelegramMessage(token string, chatID, messageID int64) {
	endpoint := fmt.Sprintf("%s/bot%s/deleteMessage", ctrl.config.APIBase, token)
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
	return infra.DB.
		WithContext(ctx).
		Where("chat_id = ? AND ocserv_user_id = ?", chatID, ocservUserID).
		FirstOrCreate(account).Error
}

func ensureReceiptDir(cfg config.TelegramConfig) error {
	return os.MkdirAll(filepath.Clean(cfg.ReceiptsDir), 0o750)
}

func defaultNotifyLang(settings *models.TelegramSettings) string {
	if settings != nil && strings.TrimSpace(settings.DefaultLanguage) != "" {
		return settings.DefaultLanguage
	}
	return models.TelegramLanguageEN
}

func gigabytesToBytes(gb int) int64 {
	return int64(gb) * (1 << 30)
}

func bytesToGigabytes(bytes int64) int {
	return int(bytes / (1 << 30))
}
