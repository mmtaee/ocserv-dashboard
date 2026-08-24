package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// TelegramSettings holds the singleton configuration for the Telegram bot.
// Only one row is expected to exist; callers should upsert by ID=1.
type TelegramSettings struct {
	ID                  uint             `json:"-" gorm:"primaryKey"`
	Enabled             bool             `json:"enabled" gorm:"default:false"`
	BotToken            string           `json:"bot_token" gorm:"type:varchar(255)"`
	BotUsername         string           `json:"bot_username" gorm:"type:varchar(64)"`
	AdminChatID         int64            `json:"admin_chat_id" gorm:"default:0"`
	LowQuotaThresholdMB int              `json:"low_quota_threshold_mb" gorm:"default:200"`
	DefaultLanguage     TelegramLanguage `json:"default_language" gorm:"type:varchar(8);default:'en'" enums:"en,fa,ar,ru,zh-cn,zh-tw,it" validate:"required"`
	OcservHost          string           `json:"ocserv_host" gorm:"type:varchar(255)"`
	CardNumber          string           `json:"card_number" gorm:"type:varchar(64)"`
	CardHolder          string           `json:"card_holder" gorm:"type:varchar(128)"`
	// SupportUsername is a Telegram username (no @, no URL) shown to users as a
	// contact for client setup/help. Rendered as @handle and an inline t.me/...
	// button by the bot.
	SupportUsername string    `json:"support_username" gorm:"type:varchar(64)"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TelegramAccount links a Telegram chat to an OcservUser. A single chat can be
// linked to many ocserv users (the same admin may track several VPN accounts).
type TelegramAccount struct {
	ID                     uint             `json:"id" gorm:"primaryKey;autoIncrement"`
	ChatID                 int64            `json:"chat_id" gorm:"index;not null"`
	TelegramUsername       string           `json:"telegram_username" gorm:"type:varchar(64)"`
	Language               TelegramLanguage `json:"language" gorm:"type:varchar(8);default:'en'" enums:"en,fa,ar,ru,zh-cn,zh-tw,it" validate:"required"`
	OcservUserID           uint             `json:"ocserv_user_id" gorm:"index;not null;constraint:OnDelete:CASCADE"`
	CreatedAt              time.Time        `json:"created_at" gorm:"autoCreateTime"`
	LastLowQuotaNotifiedAt *time.Time       `json:"last_low_quota_notified_at"`
}

// TelegramPackage describes a sellable plan that bot users can pick when
// requesting a new account or a renewal.
type TelegramPackage struct {
	ID            uint        `json:"id" gorm:"primaryKey;autoIncrement"`
	Title         string      `json:"title" gorm:"type:varchar(128);not null"`
	Days          int         `json:"days" gorm:"not null"`
	TrafficSizeGB int         `json:"traffic_size_gb" gorm:"not null"`
	TrafficType   TrafficType `json:"traffic_type" gorm:"type:varchar(32);not null;default:'TotallyTransmit'" enums:"Free,MonthlyTransmit,MonthlyReceive,MonthlyRxTx,TotallyTransmit,TotallyReceive,TotallyRxTx" validate:"required"`
	PriceText     string      `json:"price_text" gorm:"type:varchar(64)"`
	IsActive      bool        `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// TelegramRequest tracks the lifecycle of a new-account or renewal request
// submitted by a Telegram bot user.
type TelegramRequest struct {
	ID                 uint                  `json:"id" gorm:"primaryKey;autoIncrement"`
	ChatID             int64                 `json:"chat_id" gorm:"index;not null"`
	TelegramUsername   string                `json:"telegram_username" gorm:"type:varchar(64)"`
	Type               TelegramRequestType   `json:"type" gorm:"type:varchar(16);not null" enums:"new,renew" validate:"required"`
	PackageID          *uint                 `json:"package_id"`
	TargetOcservUserID *uint                 `json:"target_ocserv_user_id"`
	DesiredUsername    string                `json:"desired_username" gorm:"type:varchar(64)"`
	Status             TelegramRequestStatus `json:"status" gorm:"type:varchar(32);index;default:'pending'" enums:"pending,awaiting_payment,payment_uploaded,approved,rejected,delivered" validate:"required"`
	ReceiptFilePath    string                `json:"receipt_file_path" gorm:"type:varchar(255)"`
	UserMessage        string                `json:"user_message" gorm:"type:text"`
	AdminNote          string                `json:"admin_note" gorm:"type:text"`
	DeliveredAt        *time.Time            `json:"delivered_at"`
	// AwaitingPaymentMessageID is the Telegram message_id of the bot message sent when
	// the request moved to awaiting_payment (may contain card details). Used to delete
	// that message if the request is later rejected.
	AwaitingPaymentMessageID *int64    `json:"awaiting_payment_message_id"`
	CreatedAt                time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (settings *TelegramSettings) BeforeCreate(_ *gorm.DB) error {
	if settings.DefaultLanguage == "" {
		settings.DefaultLanguage = TelegramLanguageEN
	}
	return settings.validateEnums()
}

func (settings *TelegramSettings) BeforeUpdate(_ *gorm.DB) error {
	return settings.validateEnums()
}

func (settings *TelegramSettings) validateEnums() error {
	if settings.DefaultLanguage == "" {
		return nil
	}
	if !settings.DefaultLanguage.IsValid() {
		return fmt.Errorf("invalid TelegramLanguage: %s", settings.DefaultLanguage)
	}
	return nil
}

func (account *TelegramAccount) BeforeCreate(_ *gorm.DB) error {
	if account.Language == "" {
		account.Language = TelegramLanguageEN
	}
	return account.validateEnums()
}

func (account *TelegramAccount) BeforeUpdate(_ *gorm.DB) error {
	return account.validateEnums()
}

func (account *TelegramAccount) validateEnums() error {
	if account.Language == "" {
		return nil
	}
	if !account.Language.IsValid() {
		return fmt.Errorf("invalid TelegramLanguage: %s", account.Language)
	}
	return nil
}

func (pkg *TelegramPackage) BeforeCreate(_ *gorm.DB) error {
	if pkg.TrafficType == "" {
		pkg.TrafficType = TotallyTransmit
	}
	return pkg.validateEnums()
}

func (pkg *TelegramPackage) BeforeUpdate(_ *gorm.DB) error {
	return pkg.validateEnums()
}

func (pkg *TelegramPackage) validateEnums() error {
	if pkg.TrafficType == "" {
		return nil
	}
	if !pkg.TrafficType.IsValid() {
		return fmt.Errorf("invalid TrafficType: %s", pkg.TrafficType)
	}
	return nil
}

func (request *TelegramRequest) BeforeCreate(_ *gorm.DB) error {
	if request.Status == "" {
		request.Status = TelegramRequestStatusPending
	}
	return request.validateEnums(true)
}

func (request *TelegramRequest) BeforeUpdate(_ *gorm.DB) error {
	return request.validateEnums(false)
}

func (request *TelegramRequest) validateEnums(requireType bool) error {
	if (requireType || request.Type != "") && !request.Type.IsValid() {
		return fmt.Errorf("invalid TelegramRequestType: %s", request.Type)
	}
	if request.Status != "" && !request.Status.IsValid() {
		return fmt.Errorf("invalid TelegramRequestStatus: %s", request.Status)
	}
	return nil
}
