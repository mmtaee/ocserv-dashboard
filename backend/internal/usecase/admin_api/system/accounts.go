package system

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

var ErrInvalidCredentials = errors.New("invalid username or password")
var ErrInvalidOldPassword = errors.New("invalid old password")

func (u *Usecase) Login(ctx context.Context, input LoginData) (*UserLoginResponse, error) {
	settings, err := u.systems.System(ctx)
	if err != nil {
		return nil, err
	}
	if settings.GoogleCaptchaSecretKey != "" {
		u.captchaMu.Lock()
		u.captcha.SetSecretKey(settings.GoogleCaptchaSecretKey)
		u.captcha.Verify(input.Token)
		valid := u.captcha.IsValid()
		u.captchaMu.Unlock()
		if !valid {
			return nil, errors.New("captcha challenge failed")
		}
	}
	user, err := u.users.GetByUsername(ctx, input.Username)
	if err != nil || !u.passwords.CheckPassword(input.Password, user.Password, user.Salt) {
		return nil, ErrInvalidCredentials
	}
	token, err := u.users.CreateToken(ctx, user, input.RememberMe)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	user.LastLogin = &now
	_ = u.users.UpdateLastLogin(ctx, user)
	return &UserLoginResponse{User: user, Token: token}, nil
}

func (u *Usecase) ResetPassword(ctx context.Context, input ResetAdminPassword) (*ResetPasswordResponse, error) {
	if u.secretKey != input.SecretKey {
		return nil, errors.New("the secret key is invalid")
	}
	user, err := u.users.GetByUsername(ctx, input.Username)
	if err != nil {
		return nil, errors.New("username not found")
	}
	password := u.passwords.CreatePassword(input.NewPassword)
	if err := u.users.ChangePassword(ctx, user.UID, password.Hash, password.Salt); err != nil {
		return nil, err
	}
	token, err := u.users.CreateToken(ctx, user, true)
	if err != nil {
		return nil, err
	}
	return &ResetPasswordResponse{User: user, Token: token}, nil
}

func (u *Usecase) CreateUser(ctx context.Context, input CreateUserData) (*models.User, error) {
	password := u.passwords.CreatePassword(input.Password)
	return u.users.CreateUser(ctx, &models.User{Username: strings.ToLower(input.Username), Password: password.Hash, Salt: password.Salt})
}

func (u *Usecase) Users(ctx context.Context, pagination *request.Pagination) ([]models.User, int64, error) {
	return u.users.Users(ctx, pagination)
}

func (u *Usecase) ChangeUserPassword(ctx context.Context, uid, passwordValue string) error {
	password := u.passwords.CreatePassword(passwordValue)
	return u.users.ChangePassword(ctx, uid, password.Hash, password.Salt)
}

func (u *Usecase) DeleteUser(ctx context.Context, actorUID, targetUID string) error {
	return u.users.DeleteUser(context.WithValue(ctx, "userUID", actorUID), targetUID)
}

func (u *Usecase) ChangeOwnPassword(ctx context.Context, uid string, input ChangeUserPasswordBySelf) error {
	user, err := u.users.GetByUID(ctx, uid)
	if err != nil {
		return err
	}
	if !u.passwords.CheckPassword(input.OldPassword, user.Password, user.Salt) {
		return ErrInvalidOldPassword
	}
	return u.ChangeUserPassword(ctx, uid, input.NewPassword)
}

func (u *Usecase) Profile(ctx context.Context, uid string) (*models.User, error) {
	return u.users.GetByUID(ctx, uid)
}

func (u *Usecase) UsersLookup(ctx context.Context) (*[]models.UsersLookup, error) {
	return u.users.UsersLookup(ctx)
}
