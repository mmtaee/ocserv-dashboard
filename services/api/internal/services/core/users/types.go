package system

import (
	"github.com/mmtaee/ocserv-users-management/api/internal/models"
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
)

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

type PermissionData struct {
	Service models.UserService `json:"service" validate:"required"`
	Action  models.UserAction  `json:"action" validate:"required"`
}

type CreateUserData struct {
	Username    string           `json:"username" validate:"required"`
	Password    string           `json:"password" validate:"required"`
	Role        *models.UserRole `json:"role" validate:"omitempty"`
	AdminID     *uint            `json:"admin_id" validate:"omitempty"`
	Permissions []PermissionData `json:"permissions" validate:"omitempty"`
}

type CreateUserResponse struct {
	User       *models.User        `json:"user" validate:"required"`
	Permission []models.Permission `json:"permission" validate:"omitempty"`
}

type UpdatePermissionData struct {
	Permissions []PermissionData `json:"permissions,omitempty" validate:"omitempty" description:"empty to remove all permissions"`
}
