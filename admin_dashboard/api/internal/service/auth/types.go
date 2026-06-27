package auth

import "github.com/mmtaee/ocserv-dashboard/core/models"

// LoginRequest represents the request body for login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the response for successful login
type LoginResponse struct {
	Token string                 `json:"token"`
	Admin *models.Administrator `json:"admin"`
}

// ChangePasswordRequest represents the request body for changing password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ChangePasswordResponse represents the response for successful password change
type ChangePasswordResponse struct {
	Token string                 `json:"token"`
	Admin *models.Administrator `json:"admin"`
}
