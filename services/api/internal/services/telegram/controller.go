package telegram

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/database"
)

const (
	// ReceiptStorageRoot is the root path on disk where receipt photos are stored
	// for telegram payment workflows. The telegram_bot service writes here too.
	ReceiptStorageRoot = "/opt/ocserv_dashboard/uploads/receipts"

	telegramAPIBase    = "https://api.telegram.org"
	telegramHTTPTimeout = 8 * time.Second
)

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
		msg = "Test message from Ocserv Dashboard"
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

	go ctl.notifyAwaitingPayment(updated)
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

	go ctl.notifyDelivery(req.ChatID, settings, formatRenewalMessage(user, newExpire))
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

func (ctl *Controller) notifyAwaitingPayment(req *models.TelegramRequest) {
	settings, err := ctl.repo.Settings(context.Background())
	if err != nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	msg := "Your request has been approved. Please send the payment receipt as a photo to this chat."
	if isFa(settings) {
		msg = "درخواست شما تایید شد. لطفاً تصویر رسید پرداخت را به همین چت ارسال کنید."
	}
	_ = sendTelegramMessage(settings.BotToken, req.ChatID, msg)
}

func (ctl *Controller) notifyRejected(req *models.TelegramRequest) {
	settings, err := ctl.repo.Settings(context.Background())
	if err != nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	msg := "Your request was rejected by the administrator."
	if req.AdminNote != "" {
		msg = msg + "\nNote: " + req.AdminNote
	}
	if isFa(settings) {
		msg = "درخواست شما توسط ادمین رد شد."
		if req.AdminNote != "" {
			msg = msg + "\nتوضیح: " + req.AdminNote
		}
	}
	_ = sendTelegramMessage(settings.BotToken, req.ChatID, msg)
}

func (ctl *Controller) notifyDelivery(chatID int64, settings *models.TelegramSettings, message string) {
	if settings == nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	_ = sendTelegramMessage(settings.BotToken, chatID, message)
}

func formatNewAccountMessage(settings *models.TelegramSettings, user *models.OcservUser, plainPassword string, expireAt time.Time) string {
	host := settings.OcservHost
	if host == "" {
		host = "<server>"
	}
	if isFa(settings) {
		return fmt.Sprintf(
			"اکانت شما آماده شد:\n"+
				"سرور: %s\n"+
				"نام کاربری: %s\n"+
				"رمز عبور: %s\n"+
				"اعتبار تا: %s\n"+
				"حجم: %d GB",
			host, user.Username, plainPassword, expireAt.Format("2006-01-02"), user.TrafficSize,
		)
	}
	return fmt.Sprintf(
		"Your account is ready:\n"+
			"Server: %s\n"+
			"Username: %s\n"+
			"Password: %s\n"+
			"Expires: %s\n"+
			"Quota: %d GB",
		host, user.Username, plainPassword, expireAt.Format("2006-01-02"), user.TrafficSize,
	)
}

func formatRenewalMessage(user *models.OcservUser, newExpire time.Time) string {
	return fmt.Sprintf(
		"Account %s renewed successfully.\nNew expire date: %s",
		user.Username, newExpire.Format("2006-01-02"),
	)
}

func isFa(s *models.TelegramSettings) bool {
	return s != nil && s.DefaultLanguage == models.TelegramLanguageFA
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

func sendTelegramMessage(token string, chatID int64, text string) error {
	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", telegramAPIBase, token)
	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("text", text)
	client := &http.Client{Timeout: telegramHTTPTimeout}
	resp, err := client.PostForm(endpoint, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram sendMessage status=%d body=%s", resp.StatusCode, string(body))
	}
	return nil
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
	return os.MkdirAll(filepath.Clean(ReceiptStorageRoot), 0o750)
}
