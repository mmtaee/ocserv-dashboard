package system

import (
	"github.com/mmtaee/ocserv-users-management/api/internal/models"
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
)

type GetSystemInitResponse struct {
	GoogleCaptchaSiteKey string `json:"google_captcha_site_key" validate:"omitempty"`
}

type GetSystemResponse struct {
	GoogleCaptchaSiteKey    string `json:"google_captcha_site_key" validate:"omitempty"`
	GoogleCaptchaSecretKey  string `json:"google_captcha_secret_key" validate:"omitempty"`
	AutoDeleteInactiveUsers bool   `json:"auto_delete_inactive_users" validate:"omitempty"`
	KeepInactiveUserDays    int    `json:"keep_inactive_user_days" validate:"omitempty"`
}

type PatchSystemUpdateData struct {
	GoogleCaptchaSiteKey    *string `json:"google_captcha_site_key" validate:"required"`
	GoogleCaptchaSecretKey  *string `json:"google_captcha_secret_key" validate:"required"`
	AutoDeleteInactiveUsers *bool   `json:"auto_delete_inactive_users" validate:"required"`
	KeepInactiveUserDays    *int    `json:"keep_inactive_user_days" validate:"required"`
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
	Password string `json:"password" validate:"required"`
	//Admin    bool   `json:"admin"`
}

type UsersResponse struct {
	Meta   request.Meta  `json:"meta" validate:"required"`
	Result []models.User `json:"result" validate:"omitempty"`
}

type ChangeUserPassword struct {
	Password string `json:"password" validate:"required"`
}

type ChangeUserPasswordBySelf struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type SetupSystem struct {
	Username                string `json:"username" validate:"required,min=2,max=16"`
	Password                string `json:"password" validate:"required,min=4,max=16"`
	GoogleCaptchaSiteKey    string `json:"google_captcha_site_key" validate:"omitempty"`
	GoogleCaptchaSecretKey  string `json:"google_captcha_secret_key" validate:"omitempty"`
	AutoDeleteInactiveUsers bool   `json:"auto_delete_inactive_users" validate:"omitempty"`
	KeepInactiveUserDays    int    `json:"keep_inactive_user_days" validate:"omitempty"`
}

type SetupSystemResponse struct {
	User   models.User   `json:"user" validate:"required"`
	System models.System `json:"system" validate:"required"`
	Token  string        `json:"token" validate:"required"`
}
