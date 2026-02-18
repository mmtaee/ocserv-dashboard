package models

type System struct {
	ID                      uint   `json:"_" gorm:"primaryKey"`
	GoogleCaptchaSecretKey  string `json:"google_captcha_secret" gorm:"type:text"`
	GoogleCaptchaSiteKey    string `json:"google_captcha_site_key" gorm:"type:text"`
	AutoDeleteInactiveUsers bool   `json:"auto_delete_inactive_users" gorm:"type:boolean;default:false"`
	KeepInactiveUserDays    int    `json:"keep_inactive_user_days" gorm:"default:30"`
}
