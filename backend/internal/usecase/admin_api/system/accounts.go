package system

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

var ErrInvalidCredentials = errors.New("invalid username or password")
var ErrInvalidOldPassword = errors.New("invalid old password")

func (u *Usecase) Login(ctx context.Context, input LoginData, userAgent string) (*UserLoginResponse, error) {
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
	token, err := u.createSession(ctx, user.ID, userAgent, input.RememberMe)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	user.LastLogin = &now
	_ = u.users.UpdateLastLogin(ctx, user)
	return &UserLoginResponse{User: user, Token: token}, nil
}

func (u *Usecase) ResetPassword(ctx context.Context, input ResetAdminPassword, userAgent string) (*ResetPasswordResponse, error) {
	if u.secretKey != input.SecretKey {
		return nil, errors.New("the secret key is invalid")
	}
	user, err := u.users.GetByUsername(ctx, input.Username)
	if err != nil {
		return nil, errors.New("username not found")
	}
	password := u.passwords.CreatePassword(input.NewPassword)
	if err := u.users.ChangePassword(ctx, user.ID, password.Hash, password.Salt); err != nil {
		return nil, err
	}
	token, err := u.createSession(ctx, user.ID, userAgent, true)
	if err != nil {
		return nil, err
	}
	return &ResetPasswordResponse{User: user, Token: token}, nil
}

func (u *Usecase) createSession(ctx context.Context, userID uint, userAgent string, rememberMe bool) (string, error) {
	token, err := crypto.GenerateSecureToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	if rememberMe {
		expiresAt = expiresAt.AddDate(0, 1, 0)
	}
	session := &models.UserToken{
		UserID: userID, Token: crypto.HashToken(token), UserAgent: normalizeUserAgent(userAgent), ExpireAt: expiresAt,
	}
	if err := u.sessions.Create(ctx, session); err != nil {
		return "", err
	}
	return token, nil
}

func normalizeUserAgent(value string) string {
	characters := []rune(strings.TrimSpace(value))
	if len(characters) > 512 {
		characters = characters[:512]
	}
	return string(characters)
}

func (u *Usecase) CreateUser(ctx context.Context, input CreateUserData) (*models.User, error) {
	password := u.passwords.CreatePassword(input.Password)
	return u.users.CreateUser(ctx, &models.User{Username: strings.ToLower(input.Username), Password: password.Hash, Salt: password.Salt})
}

func (u *Usecase) Users(ctx context.Context, pagination *request.Pagination) ([]models.User, int64, error) {
	return u.users.Users(ctx, pagination)
}

func (u *Usecase) ChangeUserPassword(ctx context.Context, id uint, passwordValue string) error {
	password := u.passwords.CreatePassword(passwordValue)
	return u.users.ChangePassword(ctx, id, password.Hash, password.Salt)
}

func (u *Usecase) DeleteUser(ctx context.Context, actorID, targetID uint) error {
	return u.users.DeleteUser(context.WithValue(ctx, "userID", actorID), targetID)
}

func (u *Usecase) ChangeOwnPassword(ctx context.Context, id uint, input ChangeUserPasswordBySelf) error {
	user, err := u.users.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !u.passwords.CheckPassword(input.OldPassword, user.Password, user.Salt) {
		return ErrInvalidOldPassword
	}
	return u.ChangeUserPassword(ctx, id, input.NewPassword)
}

func (u *Usecase) Profile(ctx context.Context, id uint) (*models.User, error) {
	return u.users.GetByID(ctx, id)
}

func (u *Usecase) UsersLookup(ctx context.Context) (*[]models.UsersLookup, error) {
	return u.users.UsersLookup(ctx)
}
