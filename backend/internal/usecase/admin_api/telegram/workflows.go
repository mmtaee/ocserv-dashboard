package telegram

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"os"
	"strings"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/authz"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	tg18n "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/telegram/i18n"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

func (u *Usecase) GetSettings(ctx context.Context) (*SettingsResponse, error) {
	settings, err := u.Repository.Settings(ctx)
	if err != nil {
		return nil, err
	}
	return settingsResponse(settings), nil
}

func (u *Usecase) PatchSettings(ctx context.Context, input PatchSettingsData) (*SettingsResponse, error) {
	updates := map[string]interface{}{}
	if input.Enabled != nil {
		updates["enabled"] = *input.Enabled
	}
	if input.BotToken != nil {
		updates["bot_token"] = *input.BotToken
		updates["bot_username"] = ""
	}
	if input.AdminChatID != nil {
		updates["admin_chat_id"] = *input.AdminChatID
	}
	if input.LowQuotaThresholdMB != nil {
		updates["low_quota_threshold_mb"] = *input.LowQuotaThresholdMB
	}
	if input.DefaultLanguage != nil {
		updates["default_language"] = *input.DefaultLanguage
	}
	if input.OcservHost != nil {
		updates["ocserv_host"] = *input.OcservHost
	}
	if input.CardNumber != nil {
		updates["card_number"] = *input.CardNumber
	}
	if input.CardHolder != nil {
		updates["card_holder"] = *input.CardHolder
	}
	if input.SupportUsername != nil {
		updates["support_username"] = strings.TrimPrefix(strings.TrimSpace(*input.SupportUsername), "@")
	}
	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}
	settings, err := u.Repository.UpdateSettings(ctx, updates)
	if err != nil {
		return nil, err
	}
	if input.BotToken != nil && *input.BotToken != "" {
		if username, err := u.client.Username(ctx, *input.BotToken); err == nil && username != "" {
			if updated, updateErr := u.Repository.UpdateSettings(ctx, map[string]interface{}{"bot_username": username}); updateErr == nil {
				settings = updated
			}
		}
	}
	return settingsResponse(settings), nil
}

func (u *Usecase) Test(ctx context.Context, message string) error {
	settings, err := u.Repository.Settings(ctx)
	if err != nil {
		return err
	}
	if settings.BotToken == "" {
		return errors.New("bot token is not set")
	}
	if settings.AdminChatID == 0 {
		return errors.New("admin chat id is not set")
	}
	if message == "" {
		message = "Test message from your dashboard"
	}
	_, err = u.client.Send(ctx, settings.BotToken, settings.AdminChatID, message, false)
	return err
}

func (u *Usecase) ListPackages(ctx context.Context, includeInactive bool) ([]models.TelegramPackage, error) {
	return u.Repository.Packages(ctx, includeInactive)
}

func (u *Usecase) CreatePackage(ctx context.Context, input CreatePackageData) (*models.TelegramPackage, error) {
	return u.Repository.CreatePackage(ctx, &models.TelegramPackage{Title: input.Title, Days: input.Days, TrafficSizeGB: input.TrafficSizeGB, TrafficType: input.TrafficType, PriceText: input.PriceText, IsActive: input.IsActive})
}

func (u *Usecase) PatchPackage(ctx context.Context, id uint, input PatchPackageData) (*models.TelegramPackage, error) {
	updates := map[string]interface{}{}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Days != nil {
		updates["days"] = *input.Days
	}
	if input.TrafficSizeGB != nil {
		updates["traffic_size_gb"] = *input.TrafficSizeGB
	}
	if input.TrafficType != nil {
		updates["traffic_type"] = *input.TrafficType
	}
	if input.PriceText != nil {
		updates["price_text"] = *input.PriceText
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}
	return u.Repository.UpdatePackage(ctx, id, updates)
}

func (u *Usecase) ListRequests(ctx context.Context, pagination *request.Pagination, status, requestType string) ([]models.TelegramRequest, int64, error) {
	if pagination.Order == "" {
		pagination.Order = "created_at"
	}
	if pagination.Sort == "" {
		pagination.Sort = "DESC"
	}
	return u.Repository.Requests(ctx, pagination, status, requestType)
}

func (u *Usecase) ReceiptPath(ctx context.Context, id uint) (string, error) {
	paymentRequest, err := u.Repository.RequestByID(ctx, id)
	if err != nil {
		return "", err
	}
	if paymentRequest.ReceiptFilePath == "" {
		return "", errors.New("no receipt uploaded")
	}
	if _, err := os.Stat(paymentRequest.ReceiptFilePath); err != nil {
		return "", errors.New("receipt file not found on disk")
	}
	return paymentRequest.ReceiptFilePath, nil
}

func (u *Usecase) DeletePaymentRequest(ctx context.Context, id uint) error {
	paymentRequest, err := u.Repository.RequestByID(ctx, id)
	if err != nil {
		return err
	}
	switch paymentRequest.Status {
	case models.TelegramRequestStatusPending, models.TelegramRequestStatusAwaitingPayment, models.TelegramRequestStatusPaymentUploaded:
		return fmt.Errorf("cannot delete an active request (status=%s)", paymentRequest.Status)
	}
	if paymentRequest.ReceiptFilePath != "" {
		_ = os.Remove(paymentRequest.ReceiptFilePath)
	}
	return u.Repository.DeleteRequest(ctx, id)
}

func (u *Usecase) Approve(ctx context.Context, id uint, input ApproveData) (*models.TelegramRequest, error) {
	paymentRequest, err := u.Repository.RequestByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if paymentRequest.Status != models.TelegramRequestStatusPending {
		return nil, fmt.Errorf("only pending requests can be approved (current=%s)", paymentRequest.Status)
	}
	note := optionalString(input.AdminNote)
	updated, err := u.Repository.UpdateRequestStatus(ctx, id, models.TelegramRequestStatusAwaitingPayment, note)
	if err != nil {
		return nil, err
	}
	u.notifyAwaitingPayment(ctx, updated, input)
	return updated, nil
}

func (u *Usecase) Reject(ctx context.Context, id uint, input RejectData) (*models.TelegramRequest, error) {
	paymentRequest, err := u.Repository.RequestByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if paymentRequest.Status == models.TelegramRequestStatusDelivered {
		return nil, errors.New("cannot reject a delivered request")
	}
	updated, err := u.Repository.UpdateRequestStatus(ctx, id, models.TelegramRequestStatusRejected, optionalString(input.AdminNote))
	if err != nil {
		return nil, err
	}
	u.notifyRejected(ctx, updated)
	return updated, nil
}

func (u *Usecase) ConfirmPayment(ctx context.Context, principal authz.Principal, id uint, input ConfirmPaymentData) (map[string]interface{}, error) {
	paymentRequest, err := u.Repository.RequestByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if paymentRequest.Status != models.TelegramRequestStatusPaymentUploaded {
		return nil, fmt.Errorf("payment can only be confirmed after receipt upload (current=%s)", paymentRequest.Status)
	}
	if paymentRequest.PackageID == nil {
		return nil, errors.New("request has no package")
	}
	plan, err := u.Repository.PackageByID(ctx, *paymentRequest.PackageID)
	if err != nil {
		return nil, fmt.Errorf("package not found: %w", err)
	}
	settings, err := u.Repository.Settings(ctx)
	if err != nil {
		return nil, err
	}
	switch paymentRequest.Type {
	case models.TelegramRequestTypeNew:
		return u.deliverNew(ctx, principal.UserID, paymentRequest, plan, settings, input)
	case models.TelegramRequestTypeRenew:
		return u.deliverRenewal(ctx, paymentRequest, plan, settings, input)
	default:
		return nil, fmt.Errorf("unknown request type: %s", paymentRequest.Type)
	}
}

func (u *Usecase) Accounts(ctx context.Context, principal authz.Principal, id uint) ([]models.TelegramAccount, error) {
	if id == 0 {
		return nil, errors.New("ocserv_user_id query parameter is required")
	}
	user, err := u.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := principal.RequireOwner(user.OwnerID); err != nil {
		return nil, err
	}
	return u.Repository.AccountsForOcservUser(ctx, user.ID)
}

func (u *Usecase) deliverNew(ctx context.Context, ownerID uint, req *models.TelegramRequest, plan *models.TelegramPackage, settings *models.TelegramSettings, input ConfirmPaymentData) (map[string]interface{}, error) {
	username := input.OverrideUsername
	if username == "" {
		username = req.DesiredUsername
	}
	if username == "" {
		username = randomValue("tg_", 4)
	}
	password := input.OverridePassword
	if password == "" {
		password = randomValue("", 6)
	}
	group := input.Group
	if group == "" {
		group = "defaults"
	}
	expiresAt := time.Now().AddDate(0, 0, plan.Days)
	user, err := u.users.Create(ctx, &models.OcservUser{OwnerID: ownerID, Group: group, Username: username, Password: password, ExpireAt: &expiresAt, TrafficType: plan.TrafficType, TrafficSize: int64(plan.TrafficSizeGB) << 30, Description: fmt.Sprintf("created via telegram bot (request #%d)", req.ID)})
	if err != nil {
		return nil, fmt.Errorf("failed to create ocserv user: %w", err)
	}
	language := settings.DefaultLanguage
	if language == "" {
		language = models.TelegramLanguageEN
	}
	_ = u.Repository.LinkAccount(ctx, &models.TelegramAccount{ChatID: req.ChatID, TelegramUsername: req.TelegramUsername, Language: language, OcservUserID: user.ID})
	if input.AdminNote != "" {
		_, _ = u.Repository.UpdateRequestStatus(ctx, req.ID, models.TelegramRequestStatusPaymentUploaded, &input.AdminNote)
	}
	if err := u.Repository.MarkDelivered(ctx, req.ID, &user.ID); err != nil {
		return nil, err
	}
	u.notify(ctx, req.ChatID, settings, formatNewAccount(settings, user, password, expiresAt))
	return map[string]interface{}{"status": "delivered", "username": user.Username}, nil
}

func (u *Usecase) deliverRenewal(ctx context.Context, req *models.TelegramRequest, plan *models.TelegramPackage, settings *models.TelegramSettings, input ConfirmPaymentData) (map[string]interface{}, error) {
	if req.TargetOcservUserID == nil {
		return nil, errors.New("renewal request has no target user")
	}
	user, err := u.Repository.OcservUserByID(ctx, *req.TargetOcservUserID)
	if err != nil {
		return nil, fmt.Errorf("target ocserv user not found: %w", err)
	}
	base := time.Now()
	if user.ExpireAt != nil && user.ExpireAt.After(base) {
		base = *user.ExpireAt
	}
	expiresAt := base.AddDate(0, 0, plan.Days)
	user.ExpireAt, user.DeactivatedAt, user.IsLocked, user.Rx, user.Tx = &expiresAt, nil, false, 0, 0
	user.TrafficType, user.TrafficSize = plan.TrafficType, int64(plan.TrafficSizeGB)<<30
	if _, err := u.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to renew ocserv user: %w", err)
	}
	if input.AdminNote != "" {
		_, _ = u.Repository.UpdateRequestStatus(ctx, req.ID, models.TelegramRequestStatusPaymentUploaded, &input.AdminNote)
	}
	if err := u.Repository.MarkDelivered(ctx, req.ID, &user.ID); err != nil {
		return nil, err
	}
	u.notify(ctx, req.ChatID, settings, formatRenewal(settings, user, expiresAt))
	return map[string]interface{}{"status": "delivered", "username": user.Username}, nil
}

func (u *Usecase) notifyAwaitingPayment(ctx context.Context, req *models.TelegramRequest, input ApproveData) {
	settings, err := u.Repository.Settings(ctx)
	if err != nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	language, _ := u.Repository.PreferredLanguageForChat(ctx, req.ChatID)
	if language == "" {
		language = defaultLanguage(settings)
	}
	card := input.CardNumber
	if card == "" {
		card = settings.CardNumber
	}
	holder := input.CardHolder
	if holder == "" {
		holder = settings.CardHolder
	}
	message := tg18n.T(language, "awaiting_intro")
	if input.ReplyToUser != "" {
		message += tg18n.T(language, "awaiting_reply_prefix") + html.EscapeString(input.ReplyToUser)
	}
	if card != "" {
		if holder == "" {
			holder = "—"
		}
		message += tg18n.T(language, "awaiting_card_line", html.EscapeString(card), html.EscapeString(holder))
	}
	message += tg18n.T(language, "awaiting_receipt_line") + tg18n.T(language, "awaiting_close") + support(settings)
	messageID, err := u.client.Send(ctx, settings.BotToken, req.ChatID, message, true)
	if err == nil && messageID > 0 {
		_ = u.Repository.SetAwaitingPaymentMessageID(ctx, req.ID, messageID)
	}
}

func (u *Usecase) notifyRejected(ctx context.Context, req *models.TelegramRequest) {
	settings, err := u.Repository.Settings(ctx)
	if err != nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	if req.AwaitingPaymentMessageID != nil {
		_ = u.client.Delete(ctx, settings.BotToken, req.ChatID, *req.AwaitingPaymentMessageID)
		_ = u.Repository.ClearAwaitingPaymentMessageID(ctx, req.ID)
	}
	message := tg18n.T(defaultLanguage(settings), "rejected_title")
	if req.AdminNote != "" {
		message += tg18n.T(defaultLanguage(settings), "rejected_reason", html.EscapeString(req.AdminNote))
	}
	message += tg18n.T(defaultLanguage(settings), "rejected_close")
	u.notify(ctx, req.ChatID, settings, message)
}

func (u *Usecase) notify(ctx context.Context, chatID int64, settings *models.TelegramSettings, message string) {
	if settings != nil && settings.Enabled && settings.BotToken != "" {
		_, _ = u.client.Send(ctx, settings.BotToken, chatID, message, true)
	}
}

func settingsResponse(value *models.TelegramSettings) *SettingsResponse {
	return &SettingsResponse{Enabled: value.Enabled, BotToken: value.BotToken, BotUsername: value.BotUsername, AdminChatID: value.AdminChatID, LowQuotaThresholdMB: value.LowQuotaThresholdMB, DefaultLanguage: value.DefaultLanguage, OcservHost: value.OcservHost, CardNumber: value.CardNumber, CardHolder: value.CardHolder, SupportUsername: value.SupportUsername}
}
func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
func randomValue(prefix string, size int) string {
	value := make([]byte, size)
	_, _ = rand.Read(value)
	return prefix + hex.EncodeToString(value)
}
func defaultLanguage(settings *models.TelegramSettings) string {
	if settings != nil && strings.TrimSpace(settings.DefaultLanguage) != "" {
		return settings.DefaultLanguage
	}
	return models.TelegramLanguageEN
}
func support(settings *models.TelegramSettings) string {
	if settings == nil {
		return ""
	}
	handle := strings.TrimPrefix(strings.TrimSpace(settings.SupportUsername), "@")
	if handle == "" {
		return ""
	}
	return tg18n.T(defaultLanguage(settings), "support_suffix", `<a href="https://t.me/`+handle+`">@`+handle+`</a>`)
}
func formatNewAccount(settings *models.TelegramSettings, user *models.OcservUser, password string, expires time.Time) string {
	host := settings.OcservHost
	if host == "" {
		host = "—"
	}
	return tg18n.T(defaultLanguage(settings), "new_account", html.EscapeString(host), html.EscapeString(user.Username), html.EscapeString(password), expires.Format("2006-01-02"), int(user.TrafficSize>>30), support(settings))
}
func formatRenewal(settings *models.TelegramSettings, user *models.OcservUser, expires time.Time) string {
	return tg18n.T(defaultLanguage(settings), "renewal", html.EscapeString(user.Username), expires.Format("2006-01-02"), int(user.TrafficSize>>30), support(settings))
}
