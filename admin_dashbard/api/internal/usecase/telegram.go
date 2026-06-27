package usecase

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

	tg18n "github.com/mmtaee/ocserv-dashboard/api/internal/services/telegram/i18n"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
)

type TelegramUsecaseInterface interface {
	GetSettings() (*SettingsResponse, error)
	UpdateSettings(data PatchSettingsData) (*SettingsResponse, error)
	Test(data TestData) error
	ListPackages(includeInactive bool) ([]models.TelegramPackage, error)
	CreatePackage(data CreatePackageData) (*models.TelegramPackage, error)
	UpdatePackage(id uint, data PatchPackageData) (*models.TelegramPackage, error)
	DeletePackage(id uint) error
	ListRequests(pagination *request.Pagination, status, reqType string) ([]models.TelegramRequest, int64, error)
	GetRequest(id uint) (*models.TelegramRequest, error)
	GetReceipt(id uint) (string, error)
	DeleteRequest(id uint) error
	Approve(id uint, data ApproveData) (*models.TelegramRequest, error)
	Reject(id uint, data RejectData) (*models.TelegramRequest, error)
	ConfirmPayment(id uint, data ConfirmPaymentData) (map[string]interface{}, error)
	AccountsForOcservUser(uid string) ([]models.TelegramAccount, error)
	DeleteAccount(id uint) error
}

type TelegramUsecase struct {
	repo           repository.TelegramRepositoryInterface
	ocservUserRepo repository.OcservUserRepositoryInterface
}

func NewTelegramUsecase(
	repo repository.TelegramRepositoryInterface,
	ocservUserRepo repository.OcservUserRepositoryInterface,
) *TelegramUsecase {
	return &TelegramUsecase{
		repo:           repo,
		ocservUserRepo: ocservUserRepo,
	}
}

func (uc *TelegramUsecase) GetSettings() (*SettingsResponse, error) {
	s, err := uc.repo.Settings(context.Background())
	if err != nil {
		return nil, err
	}
	return settingsToResponse(s), nil
}

func (uc *TelegramUsecase) UpdateSettings(data PatchSettingsData) (*SettingsResponse, error) {
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
		return nil, errors.New("no fields to update")
	}

	s, err := uc.repo.UpdateSettings(context.Background(), updates)
	if err != nil {
		return nil, err
	}

	// best-effort: refresh bot username from telegram getMe
	if data.BotToken != nil && *data.BotToken != "" {
		if uname, err := fetchBotUsername(*data.BotToken); err == nil && uname != "" {
			_, _ = uc.repo.UpdateSettings(context.Background(), map[string]interface{}{
				"bot_username": uname,
			})
			s.BotUsername = uname
		}
	}

	return settingsToResponse(s), nil
}

func (uc *TelegramUsecase) Test(data TestData) error {
	s, err := uc.repo.Settings(context.Background())
	if err != nil {
		return err
	}
	if s.BotToken == "" {
		return errors.New("bot token is not set")
	}
	if s.AdminChatID == 0 {
		return errors.New("admin chat id is not set")
	}

	msg := data.Message
	if msg == "" {
		msg = "Test message from your dashboard"
	}

	return sendTelegramMessage(s.BotToken, s.AdminChatID, msg)
}

func (uc *TelegramUsecase) ListPackages(includeInactive bool) ([]models.TelegramPackage, error) {
	return uc.repo.Packages(context.Background(), includeInactive)
}

func (uc *TelegramUsecase) CreatePackage(data CreatePackageData) (*models.TelegramPackage, error) {
	pkg := &models.TelegramPackage{
		Title:         data.Title,
		Days:          data.Days,
		TrafficSizeGB: data.TrafficSizeGB,
		TrafficType:   data.TrafficType,
		PriceText:     data.PriceText,
		IsActive:      data.IsActive,
	}
	return uc.repo.CreatePackage(context.Background(), pkg)
}

func (uc *TelegramUsecase) UpdatePackage(id uint, data PatchPackageData) (*models.TelegramPackage, error) {
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
		return nil, errors.New("no fields to update")
	}

	return uc.repo.UpdatePackage(context.Background(), id, updates)
}

func (uc *TelegramUsecase) DeletePackage(id uint) error {
	return uc.repo.DeletePackage(context.Background(), id)
}

func (uc *TelegramUsecase) ListRequests(pagination *request.Pagination, status, reqType string) ([]models.TelegramRequest, int64, error) {
	return uc.repo.Requests(context.Background(), pagination, status, reqType)
}

func (uc *TelegramUsecase) GetRequest(id uint) (*models.TelegramRequest, error) {
	return uc.repo.RequestByID(context.Background(), id)
}

func (uc *TelegramUsecase) GetReceipt(id uint) (string, error) {
	req, err := uc.repo.RequestByID(context.Background(), id)
	if err != nil {
		return "", err
	}
	if req.ReceiptFilePath == "" {
		return "", errors.New("no receipt uploaded")
	}
	if _, err := os.Stat(req.ReceiptFilePath); err != nil {
		return "", errors.New("receipt file not found on disk")
	}
	return req.ReceiptFilePath, nil
}

func (uc *TelegramUsecase) DeleteRequest(id uint) error {
	return uc.repo.DeleteRequest(context.Background(), id)
}

func (uc *TelegramUsecase) Approve(id uint, data ApproveData) (*models.TelegramRequest, error) {
	req, err := uc.repo.RequestByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if req.Status != models.TelegramRequestStatusPending {
		return nil, fmt.Errorf("only pending requests can be approved (current=%s)", req.Status)
	}

	var note *string
	if data.AdminNote != "" {
		note = &data.AdminNote
	}
	updated, err := uc.repo.UpdateRequestStatus(context.Background(), id, models.TelegramRequestStatusAwaitingPayment, note)
	if err != nil {
		return nil, err
	}

	go uc.notifyAwaitingPayment(updated, &awaitingPaymentOpts{
		CardNumber:  data.CardNumber,
		CardHolder:  data.CardHolder,
		ReplyToUser: data.ReplyToUser,
	})

	return updated, nil
}

func (uc *TelegramUsecase) Reject(id uint, data RejectData) (*models.TelegramRequest, error) {
	req, err := uc.repo.RequestByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if req.Status == models.TelegramRequestStatusDelivered {
		return nil, errors.New("cannot reject a delivered request")
	}

	var note *string
	if data.AdminNote != "" {
		note = &data.AdminNote
	}
	updated, err := uc.repo.UpdateRequestStatus(context.Background(), id, models.TelegramRequestStatusRejected, note)
	if err != nil {
		return nil, err
	}

	go uc.notifyRejected(updated)

	return updated, nil
}

func (uc *TelegramUsecase) ConfirmPayment(id uint, data ConfirmPaymentData) (map[string]interface{}, error) {
	req, err := uc.repo.RequestByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if req.Status != models.TelegramRequestStatusPaymentUploaded {
		return nil, fmt.Errorf("payment can only be confirmed after receipt upload (current=%s)", req.Status)
	}
	if req.PackageID == nil {
		return nil, errors.New("request has no package")
	}

	pkg, err := uc.repo.PackageByID(context.Background(), *req.PackageID)
	if err != nil {
		return nil, fmt.Errorf("package not found: %w", err)
	}

	settings, err := uc.repo.Settings(context.Background())
	if err != nil {
		return nil, err
	}

	switch req.Type {
	case models.TelegramRequestTypeNew:
		return uc.deliverNewAccount(req, pkg, settings, &data)
	case models.TelegramRequestTypeRenew:
		return uc.deliverRenewal(req, pkg, settings, &data)
	default:
		return nil, fmt.Errorf("unknown request type: %s", req.Type)
	}
}

func (uc *TelegramUsecase) AccountsForOcservUser(uid string) ([]models.TelegramAccount, error) {
	user, err := uc.ocservUserRepo.GetByUID(context.Background(), uid)
	if err != nil {
		return nil, err
	}
	return uc.repo.AccountsForOcservUser(context.Background(), user.ID)
}

func (uc *TelegramUsecase) DeleteAccount(id uint) error {
	return uc.repo.DeleteAccount(context.Background(), id)
}

func (uc *TelegramUsecase) notifyAwaitingPayment(req *models.TelegramRequest, opts *awaitingPaymentOpts) {
	settings, err := uc.repo.Settings(context.Background())
	if err != nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	lang := uc.resolveNotifyLang(context.Background(), req.ChatID, settings)
	var pkg *models.TelegramPackage
	if req.PackageID != nil && *req.PackageID > 0 {
		if p, err := uc.repo.PackageByID(context.Background(), *req.PackageID); err == nil {
			pkg = p
		}
	}
	msg := formatAwaitingPaymentMessage(lang, settings, opts, pkg)
	msgID, err := sendTelegramHTMLMessageWithID(settings.BotToken, req.ChatID, msg)
	if err != nil || msgID <= 0 {
		return
	}
	_ = uc.repo.SetAwaitingPaymentMessageID(context.Background(), req.ID, msgID)
}

func (uc *TelegramUsecase) resolveNotifyLang(ctx context.Context, chatID int64, settings *models.TelegramSettings) string {
	if l, err := uc.repo.PreferredLanguageForChat(ctx, chatID); err == nil && strings.TrimSpace(l) != "" {
		return strings.TrimSpace(l)
	}
	if settings != nil && settings.DefaultLanguage != "" {
		return settings.DefaultLanguage
	}
	return models.TelegramLanguageEN
}

func (uc *TelegramUsecase) notifyRejected(req *models.TelegramRequest) {
	settings, err := uc.repo.Settings(context.Background())
	if err != nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	if req.AwaitingPaymentMessageID != nil && *req.AwaitingPaymentMessageID > 0 {
		deleteTelegramMessage(settings.BotToken, req.ChatID, *req.AwaitingPaymentMessageID)
		_ = uc.repo.ClearAwaitingPaymentMessageID(context.Background(), req.ID)
	}
	msg := formatRejectedMessage(settings, req.AdminNote)
	_ = sendTelegramHTMLMessage(settings.BotToken, req.ChatID, msg)
}

func (uc *TelegramUsecase) notifyDelivery(chatID int64, settings *models.TelegramSettings, message string) {
	if settings == nil || settings.BotToken == "" || !settings.Enabled {
		return
	}
	_ = sendTelegramHTMLMessage(settings.BotToken, chatID, message)
}

func (uc *TelegramUsecase) deliverNewAccount(
	req *models.TelegramRequest,
	pkg *models.TelegramPackage,
	settings *models.TelegramSettings,
	data *ConfirmPaymentData,
) (map[string]interface{}, error) {
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
		TrafficSize: gigabytesToBytes(pkg.TrafficSizeGB),
		Description: fmt.Sprintf("created via telegram bot (request #%d)", req.ID),
	}

	created, err := uc.ocservUserRepo.Create(context.Background(), user)
	if err != nil {
		return nil, fmt.Errorf("failed to create ocserv user: %w", err)
	}

	// link telegram account to the new ocserv user
	if err := linkTelegramAccount(context.Background(), req.ChatID, req.TelegramUsername, settings.DefaultLanguage, created.ID); err != nil {
		// non-fatal, just log via admin note path
		_ = err
	}

	if data.AdminNote != "" {
		_, _ = uc.repo.UpdateRequestStatus(context.Background(), req.ID, models.TelegramRequestStatusPaymentUploaded, &data.AdminNote)
	}
	if err := uc.repo.MarkDelivered(context.Background(), req.ID, &created.ID); err != nil {
		return nil, err
	}

	go uc.notifyDelivery(req.ChatID, settings, formatNewAccountMessage(settings, created, password, expireAt))
	return map[string]interface{}{
		"status":   "delivered",
		"username": created.Username,
	}, nil
}

func (uc *TelegramUsecase) deliverRenewal(
	req *models.TelegramRequest,
	pkg *models.TelegramPackage,
	settings *models.TelegramSettings,
	data *ConfirmPaymentData,
) (map[string]interface{}, error) {
	if req.TargetOcservID == nil {
		return nil, errors.New("renewal request has no target user")
	}

	user, err := uc.findOcservUserByID(context.Background(), *req.TargetOcservID)
	if err != nil {
		return nil, fmt.Errorf("target ocserv user not found: %w", err)
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

	if _, err := uc.ocservUserRepo.Update(context.Background(), user); err != nil {
		return nil, fmt.Errorf("failed to renew ocserv user: %w", err)
	}

	if data.AdminNote != "" {
		_, _ = uc.repo.UpdateRequestStatus(context.Background(), req.ID, models.TelegramRequestStatusPaymentUploaded, &data.AdminNote)
	}
	if err := uc.repo.MarkDelivered(context.Background(), req.ID, &user.ID); err != nil {
		return nil, err
	}

	go uc.notifyDelivery(req.ChatID, settings, formatRenewalMessage(settings, user, newExpire))
	return map[string]interface{}{
		"status":   "delivered",
		"username": user.Username,
	}, nil
}

func (uc *TelegramUsecase) findOcservUserByID(ctx context.Context, id uint) (*models.OcservUser, error) {
	var user models.OcservUser
	if err := database.GetConnection().
		WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

type awaitingPaymentOpts struct {
	CardNumber  string
	CardHolder  string
	ReplyToUser string
}

// ==================== TYPES ====================
type SettingsResponse struct {
	Enabled             bool   `json:"enabled"`
	BotToken            string `json:"bot_token"`
	BotUsername         string `json:"bot_username"`
	AdminChatID         int64  `json:"admin_chat_id"`
	LowQuotaThresholdMB int    `json:"low_quota_threshold_mb"`
	DefaultLanguage     string `json:"default_language"`
	OcservHost          string `json:"ocserv_host"`
	CardNumber          string `json:"card_number"`
	CardHolder          string `json:"card_holder"`
	SupportUsername     string `json:"support_username"`
}

type PatchSettingsData struct {
	Enabled             *bool   `json:"enabled"`
	BotToken            *string `json:"bot_token"`
	AdminChatID         *int64  `json:"admin_chat_id"`
	LowQuotaThresholdMB *int    `json:"low_quota_threshold_mb" validate:"omitempty,min=10,max=10240"`
	DefaultLanguage     *string `json:"default_language" validate:"omitempty,oneof=en fa ar ru zh-cn zh-tw it"`
	OcservHost          *string `json:"ocserv_host"`
	CardNumber          *string `json:"card_number" validate:"omitempty,max=64"`
	CardHolder          *string `json:"card_holder" validate:"omitempty,max=128"`
	SupportUsername     *string `json:"support_username" validate:"omitempty,max=64"`
}

type TestData struct {
	Message string `json:"message"`
}

type CreatePackageData struct {
	Title         string `json:"title" validate:"required,min=2,max=128"`
	Days          int    `json:"days" validate:"required,min=1,max=3650"`
	TrafficSizeGB int    `json:"traffic_size_gb" validate:"min=0,max=100000"`
	TrafficType   string `json:"traffic_type" validate:"required,oneof=Free MonthlyTransmit MonthlyReceive MonthlyRxTx TotallyTransmit TotallyReceive TotallyRxTx"`
	PriceText     string `json:"price_text" validate:"omitempty,max=64"`
	IsActive      bool   `json:"is_active"`
}

type PatchPackageData struct {
	Title         *string `json:"title" validate:"omitempty,min=2,max=128"`
	Days          *int    `json:"days" validate:"omitempty,min=1,max=3650"`
	TrafficSizeGB *int    `json:"traffic_size_gb" validate:"omitempty,min=0,max=100000"`
	TrafficType   *string `json:"traffic_type" validate:"omitempty,oneof=Free MonthlyTransmit MonthlyReceive MonthlyRxTx TotallyTransmit TotallyReceive TotallyRxTx"`
	PriceText     *string `json:"price_text" validate:"omitempty,max=64"`
	IsActive      *bool   `json:"is_active"`
}

type RequestsResponse struct {
	Meta   request.Meta             `json:"meta"`
	Result []models.TelegramRequest `json:"result"`
}

type ApproveData struct {
	AdminNote string `json:"admin_note" validate:"omitempty,max=1024"`
	CardNumber  string `json:"card_number" validate:"omitempty,max=64"`
	CardHolder  string `json:"card_holder" validate:"omitempty,max=128"`
	ReplyToUser string `json:"reply_to_user" validate:"omitempty,max=1024"`
}

type RejectData struct {
	AdminNote string `json:"admin_note" validate:"omitempty,max=1024"`
}

type ConfirmPaymentData struct {
	OverrideUsername string `json:"override_username" validate:"omitempty,min=3,max=64"`
	OverridePassword string `json:"override_password" validate:"omitempty,min=4,max=64"`
	Owner            string `json:"owner" validate:"omitempty,max=16"`
	Group            string `json:"group" validate:"omitempty,max=16"`
	AdminNote        string `json:"admin_note" validate:"omitempty,max=1024"`
}

// ==================== HELPERS ====================
func settingsToResponse(s *models.TelegramSettings) *SettingsResponse {
	return &SettingsResponse{
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

const telegramAPIBase = "https://api.telegram.org"
const telegramHTTPTimeout = 8 * time.Second

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

func ensureReceiptDir() error {
	return os.MkdirAll(filepath.Clean(receiptStorageRoot()), 0o750)
}

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

func gigabytesToBytes(gb int) int64 {
	return int64(gb) * (1 << 30)
}

func bytesToGigabytes(bytes int64) int {
	return int(bytes / (1 << 30))
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
