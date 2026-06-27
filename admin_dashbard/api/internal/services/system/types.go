package system

import (
	"github.com/mmtaee/ocserv-dashboard/api/internal/models"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
)

type GetSystemInitResponse struct {
	GoogleCaptchaSiteKey string `json:"google_captcha_site_key" validate:"omitempty"`
	TelegramBotEnabled   bool   `json:"telegram_bot_enabled" validate:"omitempty"`
}

type GetSystemResponse struct {
	GoogleCaptchaSiteKey        string `json:"google_captcha_site_key" validate:"omitempty"`
	GoogleCaptchaSecretKey      string `json:"google_captcha_secret_key" validate:"omitempty"`
	AutoDeleteInactiveUsers     bool   `json:"auto_delete_inactive_users" validate:"omitempty"`
	KeepInactiveUserDays        int    `json:"keep_inactive_user_days" validate:"omitempty"`
	ClientProfileServerAddress  string `json:"client_profile_server_address" validate:"omitempty"`
	ClientProfileServerPort     int    `json:"client_profile_server_port" validate:"omitempty"`
	ClientProfileConnectionName string `json:"client_profile_connection_name" validate:"omitempty"`
}

type PatchSystemUpdateData struct {
	GoogleCaptchaSiteKey        *string `json:"google_captcha_site_key" validate:"required"`
	GoogleCaptchaSecretKey      *string `json:"google_captcha_secret_key" validate:"required"`
	AutoDeleteInactiveUsers     *bool   `json:"auto_delete_inactive_users" validate:"required"`
	KeepInactiveUserDays        *int    `json:"keep_inactive_user_days" validate:"required"`
	ClientProfileServerAddress  *string `json:"client_profile_server_address" validate:"required"`
	ClientProfileServerPort     *int    `json:"client_profile_server_port" validate:"required"`
	ClientProfileConnectionName *string `json:"client_profile_connection_name" validate:"required"`
}

type LoginData struct {
	Username   string `json:"username" validate:"required,min=2,max=16" example:"john_doe" `
	Password   string `json:"password" validate:"required,min=2,max=16" example:"doe123456"`
	RememberMe bool   `json:"remember_me" desc:"remember for a month"`
	Token      string `json:"token" desc:"captcha v2 token"`
}

type UserLoginResponse struct {
	User  *models.User `json:"user" validate:"required"`
	Token string       `json:"token" validate:"required"`
}

type CreateUserData struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=4,max=16"`
	//Admin    bool   `json:"admin"`
}

type UsersResponse struct {
	Meta   request.Meta  `json:"meta" validate:"required"`
	Result []models.User `json:"result" validate:"omitempty"`
}

type ChangeUserPassword struct {
	Password string `json:"password" validate:"required,min=4,max=16"`
}

type ChangeUserPasswordBySelf struct {
	OldPassword string `json:"old_password" validate:"required,min=4,max=16"`
	NewPassword string `json:"new_password" validate:"required,min=4,max=16"`
}

type SetupSystem struct {
	Username                    string `json:"username" validate:"required,min=2,max=16"`
	Password                    string `json:"password" validate:"required,min=4,max=16"`
	GoogleCaptchaSiteKey        string `json:"google_captcha_site_key" validate:"omitempty"`
	GoogleCaptchaSecretKey      string `json:"google_captcha_secret_key" validate:"omitempty"`
	AutoDeleteInactiveUsers     bool   `json:"auto_delete_inactive_users" validate:"omitempty"`
	KeepInactiveUserDays        int    `json:"keep_inactive_user_days" validate:"omitempty"`
	ClientProfileServerAddress  string `json:"client_profile_server_address" validate:"omitempty"`
	ClientProfileServerPort     int    `json:"client_profile_server_port" validate:"omitempty"`
	ClientProfileConnectionName string `json:"client_profile_connection_name" validate:"omitempty"`
}

type SetupSystemResponse struct {
	User   models.User   `json:"user" validate:"required"`
	System models.System `json:"system" validate:"required"`
	Token  string        `json:"token" validate:"required"`
}

type ResetPasswordResponse struct {
	User  *models.User `json:"user" validate:"required"`
	Token string       `json:"token" validate:"required"`
}
type DashboardRelease struct {
	Current string `json:"current" validate:"required"`
	Latest  string `json:"latest" validate:"required"`
}

type ResetAdminPassword struct {
	Username    string `json:"username" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=4,max=16"`
	SecretKey   string `json:"secret_key" validate:"required,min=16,max=64"`
}
