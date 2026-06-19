package system

// UpdateSystemRequest represents the request body for updating system config
type UpdateSystemRequest struct {
	GoogleCaptchaSecretKey      string `json:"google_captcha_secret"`
	GoogleCaptchaSiteKey        string `json:"google_captcha_site_key"`
	AutoDeleteInactiveUsers     bool   `json:"auto_delete_inactive_users"`
	KeepInactiveUserDays        int    `json:"keep_inactive_user_days"`
	ClientProfileServerAddress  string `json:"client_profile_server_address"`
	ClientProfileServerPort     int    `json:"client_profile_server_port"`
	ClientProfileConnectionName string `json:"client_profile_connection_name"`
}
