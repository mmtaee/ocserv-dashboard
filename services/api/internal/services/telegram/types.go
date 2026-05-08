package telegram

import (
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/common/models"
)

// SettingsResponse mirrors TelegramSettings but exposes only what the frontend needs.
type SettingsResponse struct {
	Enabled             bool   `json:"enabled"`
	BotToken            string `json:"bot_token"`
	BotUsername         string `json:"bot_username"`
	AdminChatID         int64  `json:"admin_chat_id"`
	LowQuotaThresholdMB int    `json:"low_quota_threshold_mb"`
	DefaultLanguage     string `json:"default_language"`
	OcservHost          string `json:"ocserv_host"`
}

// PatchSettingsData accepts partial updates from admin UI.
type PatchSettingsData struct {
	Enabled             *bool   `json:"enabled"`
	BotToken            *string `json:"bot_token"`
	AdminChatID         *int64  `json:"admin_chat_id"`
	LowQuotaThresholdMB *int    `json:"low_quota_threshold_mb" validate:"omitempty,min=10,max=10240"`
	DefaultLanguage     *string `json:"default_language" validate:"omitempty,oneof=en fa"`
	OcservHost          *string `json:"ocserv_host"`
}

type TestData struct {
	Message string `json:"message"`
}

type CreatePackageData struct {
	Title         string `json:"title" validate:"required,min=2,max=128"`
	Days          int    `json:"days" validate:"required,min=1,max=3650"`
	TrafficSizeGB int    `json:"traffic_size_gb" validate:"min=0,max=100000"`
	TrafficType   string `json:"traffic_type" validate:"required,oneof=Free MonthlyTransmit MonthlyReceive TotallyTransmit TotallyReceive"`
	PriceText     string `json:"price_text" validate:"omitempty,max=64"`
	IsActive      bool   `json:"is_active"`
}

type PatchPackageData struct {
	Title         *string `json:"title" validate:"omitempty,min=2,max=128"`
	Days          *int    `json:"days" validate:"omitempty,min=1,max=3650"`
	TrafficSizeGB *int    `json:"traffic_size_gb" validate:"omitempty,min=0,max=100000"`
	TrafficType   *string `json:"traffic_type" validate:"omitempty,oneof=Free MonthlyTransmit MonthlyReceive TotallyTransmit TotallyReceive"`
	PriceText     *string `json:"price_text" validate:"omitempty,max=64"`
	IsActive      *bool   `json:"is_active"`
}

type RequestsResponse struct {
	Meta   request.Meta             `json:"meta"`
	Result []models.TelegramRequest `json:"result"`
}

type ApproveData struct {
	AdminNote string `json:"admin_note" validate:"omitempty,max=1024"`
}

type RejectData struct {
	AdminNote string `json:"admin_note" validate:"omitempty,max=1024"`
}

// ConfirmPaymentData lets the admin override the auto-generated username
// for new account requests when needed.
type ConfirmPaymentData struct {
	OverrideUsername string `json:"override_username" validate:"omitempty,min=3,max=64"`
	OverridePassword string `json:"override_password" validate:"omitempty,min=4,max=64"`
	Owner            string `json:"owner" validate:"omitempty,max=16"`
	Group            string `json:"group" validate:"omitempty,max=16"`
	AdminNote        string `json:"admin_note" validate:"omitempty,max=1024"`
}
