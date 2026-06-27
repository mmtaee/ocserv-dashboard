package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mmtaee/ocserv-dashboard/api/internal/models"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/captcha"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/crypto"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
)

type SystemUsecaseInterface interface {
	DashboardRelease() (current, latest string, err error)
	SetupSystem(ctx context.Context, data SetupSystemData) (user *models.User, system *models.System, token string, err error)
	ResetAdminPassword(ctx context.Context, data ResetAdminPasswordData) (user *models.User, token string, err error)
	SystemInit(ctx context.Context) (googleCaptchaSiteKey string, telegramBotEnabled bool, err error)
	System(ctx context.Context) (*GetSystemResponse, error)
	SystemUpdate(ctx context.Context, data PatchSystemUpdateData, userUID string) (*GetSystemResponse, error)
	Login(ctx context.Context, data LoginData) (user *models.User, token string, err error)
	CreateUser(ctx context.Context, data CreateUserData) (*models.User, error)
	Users(ctx context.Context, pagination *request.Pagination) ([]models.User, int64, error)
	ChangeUserPasswordByAdmin(ctx context.Context, userTargetID string, data ChangeUserPasswordData) error
	DeleteUser(ctx context.Context, deleteUserID, userUID string) error
	ChangePasswordBySelf(ctx context.Context, userUID string, data ChangePasswordBySelfData) error
	Profile(ctx context.Context, userUID string) (*models.User, error)
	UsersLookup(ctx context.Context) ([]models.UsersLookup, error)
}

type SystemUsecase struct {
	systemRepo      repository.SystemRepositoryInterface
	userRepo        repository.UserRepositoryInterface
	captchaVerifier captcha.GoogleCaptchaInterface
	cryptoRepo      crypto.CustomPasswordInterface
}

func NewSystemUsecase(
	systemRepo repository.SystemRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	captchaVerifier captcha.GoogleCaptchaInterface,
	cryptoRepo crypto.CustomPasswordInterface,
) *SystemUsecase {
	return &SystemUsecase{
		systemRepo:      systemRepo,
		userRepo:        userRepo,
		captchaVerifier: captchaVerifier,
		cryptoRepo:      cryptoRepo,
	}
}

func (uc *SystemUsecase) DashboardRelease() (current, latest string, err error) {
	current = os.Getenv("CURRENT_RELEASE")

	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"https://api.github.com/repos/mmtaee/ocserv-dashboard/releases/latest",
		nil,
	)
	if err != nil {
		return "", "", errors.New("failed to create latest release request")
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ocserv-dashboard")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", errors.New("failed to fetch latest release")
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logger.Error("error on close io.ReadCloser: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", "", errors.New("failed to fetch latest release")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", errors.New("failed to read latest release")
	}

	var gh struct {
		TagName string `json:"tag_name"`
	}

	if err = json.Unmarshal(body, &gh); err != nil {
		return "", "", errors.New("failed to parse latest release")
	}

	latest = strings.TrimSpace(gh.TagName)
	return current, latest, nil
}

func (uc *SystemUsecase) SetupSystem(ctx context.Context, data SetupSystemData) (adminUser *models.User, system *models.System, token string, err error) {
	if _, err := uc.systemRepo.System(ctx); err == nil {
		return nil, nil, "", errors.New("the system is already configured")
	}

	passwd := uc.cryptoRepo.CreatePassword(data.Password)

	adminUser = &models.User{
		Username: strings.ToLower(data.Username),
		Password: passwd.Hash,
		Salt:     passwd.Salt,
		IsAdmin:  true,
	}

	inactiveDays := data.KeepInactiveUserDays
	if inactiveDays < 1 {
		inactiveDays = 1
	}
	clientProfileServerAddress := strings.TrimSpace(data.ClientProfileServerAddress)
	if clientProfileServerAddress != "" {
		if _, err := user.NormalizeProfileServerAddress(clientProfileServerAddress); err != nil {
			return nil, nil, "", err
		}
	}

	clientProfileConnectionName := strings.TrimSpace(data.ClientProfileConnectionName)
	if clientProfileConnectionName != "" {
		if _, err := user.NormalizeProfileConnectionName(clientProfileConnectionName); err != nil {
			return nil, nil, "", err
		}
	}

	clientProfileServerPort := data.ClientProfileServerPort
	if clientProfileServerPort == 0 {
		clientProfileServerPort = 443
	}
	if _, err := user.NormalizeProfileServerPort(clientProfileServerPort); err != nil {
		return nil, nil, "", err
	}

	system = &models.System{
		GoogleCaptchaSiteKey:        data.GoogleCaptchaSiteKey,
		GoogleCaptchaSecretKey:      data.GoogleCaptchaSecretKey,
		AutoDeleteInactiveUsers:     data.AutoDeleteInactiveUsers,
		KeepInactiveUserDays:        inactiveDays,
		ClientProfileServerAddress:  clientProfileServerAddress,
		ClientProfileServerPort:     clientProfileServerPort,
		ClientProfileConnectionName: clientProfileConnectionName,
	}
	newUser, newSystem, err := uc.systemRepo.SystemSetup(ctx, adminUser, system)
	if err != nil {
		return nil, nil, "", err
	}

	token, err = uc.userRepo.CreateToken(ctx, newUser, true)
	if err != nil {
		return nil, nil, "", err
	}

	return newUser, newSystem, token, nil
}

func (uc *SystemUsecase) ResetAdminPassword(ctx context.Context, data ResetAdminPasswordData) (user *models.User, token string, err error) {
	if config.Get().SecretKey != data.SecretKey {
		return nil, "", errors.New("the secret key is invalid")
	}

	user, err = uc.userRepo.GetByUsername(ctx, data.Username)
	if err != nil {
		return nil, "", errors.New("username not found")
	}

	passwd := uc.cryptoRepo.CreatePassword(data.NewPassword)
	if err = uc.userRepo.ChangePassword(ctx, user.UID, passwd.Hash, passwd.Salt); err != nil {
		return nil, "", err
	}

	token, err = uc.userRepo.CreateToken(ctx, user, true)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (uc *SystemUsecase) SystemInit(ctx context.Context) (googleCaptchaSiteKey string, telegramBotEnabled bool, err error) {
	cfg, err := uc.systemRepo.System(ctx)
	if err != nil {
		if errors.Is(err, errors.New("record not found")) {
			return "", false, nil
		}
		return "", false, err
	}

	return cfg.GoogleCaptchaSiteKey, os.Getenv("TELEGRAM_BOT_ENABLED") == "true", nil
}

func (uc *SystemUsecase) System(ctx context.Context) (*GetSystemResponse, error) {
	cfg, err := uc.systemRepo.System(ctx)
	if err != nil {
		if errors.Is(err, errors.New("record not found")) {
			return nil, nil
		}
		return nil, err
	}
	return &GetSystemResponse{
		GoogleCaptchaSiteKey:        cfg.GoogleCaptchaSiteKey,
		GoogleCaptchaSecretKey:      cfg.GoogleCaptchaSecretKey,
		AutoDeleteInactiveUsers:     cfg.AutoDeleteInactiveUsers,
		KeepInactiveUserDays:        cfg.KeepInactiveUserDays,
		ClientProfileServerAddress:  cfg.ClientProfileServerAddress,
		ClientProfileServerPort:     cfg.ClientProfileServerPort,
		ClientProfileConnectionName: cfg.ClientProfileConnectionName,
	}, nil
}

func (uc *SystemUsecase) SystemUpdate(ctx context.Context, data PatchSystemUpdateData, userUID string) (*GetSystemResponse, error) {
	system := models.System{}

	if data.GoogleCaptchaSiteKey != nil {
		system.GoogleCaptchaSiteKey = *data.GoogleCaptchaSiteKey
	}
	if data.GoogleCaptchaSecretKey != nil {
		system.GoogleCaptchaSecretKey = *data.GoogleCaptchaSecretKey
	}
	if data.AutoDeleteInactiveUsers != nil {
		system.AutoDeleteInactiveUsers = *data.AutoDeleteInactiveUsers
	}
	if data.KeepInactiveUserDays != nil {
		inactiveDays := *data.KeepInactiveUserDays
		if inactiveDays < 1 {
			inactiveDays = 1
		}
		system.KeepInactiveUserDays = inactiveDays
	}
	if data.ClientProfileServerAddress != nil {
		clientProfileServerAddress := strings.TrimSpace(*data.ClientProfileServerAddress)
		if clientProfileServerAddress != "" {
			if _, err := user.NormalizeProfileServerAddress(clientProfileServerAddress); err != nil {
				return nil, err
			}
		}
		system.ClientProfileServerAddress = clientProfileServerAddress
	}

	if data.ClientProfileServerPort != nil {
		if _, err := user.NormalizeProfileServerPort(*data.ClientProfileServerPort); err != nil {
			return nil, err
		}
		system.ClientProfileServerPort = *data.ClientProfileServerPort
	}

	if data.ClientProfileConnectionName != nil {
		clientProfileConnectionName := strings.TrimSpace(*data.ClientProfileConnectionName)
		if clientProfileConnectionName != "" {
			if _, err := user.NormalizeProfileConnectionName(clientProfileConnectionName); err != nil {
				return nil, err
			}
		}
		system.ClientProfileConnectionName = clientProfileConnectionName
	}

	updateCtx := context.WithValue(ctx, "userUID", userUID)
	updatedConfig, err := uc.systemRepo.SystemUpdate(updateCtx, &system)
	if err != nil {
		return nil, err
	}
	return &GetSystemResponse{
		GoogleCaptchaSiteKey:        updatedConfig.GoogleCaptchaSiteKey,
		GoogleCaptchaSecretKey:      updatedConfig.GoogleCaptchaSecretKey,
		AutoDeleteInactiveUsers:     updatedConfig.AutoDeleteInactiveUsers,
		KeepInactiveUserDays:        updatedConfig.KeepInactiveUserDays,
		ClientProfileServerAddress:  updatedConfig.ClientProfileServerAddress,
		ClientProfileServerPort:     updatedConfig.ClientProfileServerPort,
		ClientProfileConnectionName: updatedConfig.ClientProfileConnectionName,
	}, nil
}

func (uc *SystemUsecase) Login(ctx context.Context, data LoginData) (user *models.User, token string, err error) {
	system, err := uc.systemRepo.System(ctx)
	if err != nil {
		return nil, "", err
	}

	if secretKey := system.GoogleCaptchaSecretKey; secretKey != "" {
		uc.captchaVerifier.SetSecretKey(secretKey)
		uc.captchaVerifier.Verify(data.Token)
		if !uc.captchaVerifier.IsValid() {
			return nil, "", errors.New("captcha challenge failed")
		}
	}

	user, err = uc.userRepo.GetByUsername(ctx, data.Username)
	if err != nil {
		return nil, "", errors.New("invalid username or password")
	}

	if ok := uc.cryptoRepo.CheckPassword(data.Password, user.Password, user.Salt); !ok {
		return nil, "", errors.New("invalid username or password")
	}

	token, err = uc.userRepo.CreateToken(ctx, user, data.RememberMe)
	if err != nil {
		return nil, "", err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), "userUID", user.UID), 10*time.Second)
		ctx = context.WithValue(ctx, "username", user.Username)
		defer cancel()

		now := time.Now()
		user.LastLogin = &now
		_ = uc.userRepo.UpdateLastLogin(ctx, user)
	}()

	return user, token, nil
}

func (uc *SystemUsecase) CreateUser(ctx context.Context, data CreateUserData) (*models.User, error) {
	passwd := uc.cryptoRepo.CreatePassword(data.Password)

	user := &models.User{
		Username: strings.ToLower(data.Username),
		Password: passwd.Hash,
		Salt:     passwd.Salt,
		IsAdmin:  false,
	}

	newUser, err := uc.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

func (uc *SystemUsecase) Users(ctx context.Context, pagination *request.Pagination) ([]models.User, int64, error) {
	return uc.userRepo.Users(ctx, pagination)
}

func (uc *SystemUsecase) ChangeUserPasswordByAdmin(ctx context.Context, userTargetID string, data ChangeUserPasswordData) error {
	passwd := uc.cryptoRepo.CreatePassword(data.Password)
	return uc.userRepo.ChangePassword(ctx, userTargetID, passwd.Hash, passwd.Salt)
}

func (uc *SystemUsecase) DeleteUser(ctx context.Context, deleteUserID, userUID string) error {
	updateCtx := context.WithValue(ctx, "userUID", userUID)
	return uc.userRepo.DeleteUser(updateCtx, deleteUserID)
}

func (uc *SystemUsecase) ChangePasswordBySelf(ctx context.Context, userUID string, data ChangePasswordBySelfData) error {
	user, err := uc.userRepo.GetByUID(ctx, userUID)
	if err != nil {
		return err
	}
	if ok := uc.cryptoRepo.CheckPassword(data.OldPassword, user.Password, user.Salt); !ok {
		return errors.New("invalid old password")
	}

	passwd := uc.cryptoRepo.CreatePassword(data.NewPassword)
	return uc.userRepo.ChangePassword(ctx, userUID, passwd.Hash, passwd.Salt)
}

func (uc *SystemUsecase) Profile(ctx context.Context, userUID string) (*models.User, error) {
	return uc.userRepo.GetByUID(ctx, userUID)
}

func (uc *SystemUsecase) UsersLookup(ctx context.Context) ([]models.UsersLookup, error) {
	return uc.userRepo.UsersLookup(ctx)
}

type SetupSystemData struct {
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

type DashboardReleaseResponse struct {
	Current string `json:"current" validate:"required"`
	Latest  string `json:"latest" validate:"required"`
}

type ResetAdminPasswordData struct {
	Username    string `json:"username" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=4,max=16"`
	SecretKey   string `json:"secret_key" validate:"required,min=16,max=64"`
}

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
	Username   string `json:"username" validate:"required,min=2,max=16" example:"john_doe"`
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
}

type UsersResponse struct {
	Meta   request.Meta  `json:"meta" validate:"required"`
	Result []models.User `json:"result" validate:"omitempty"`
}

type ChangeUserPasswordData struct {
	Password string `json:"password" validate:"required,min=4,max=16"`
}

type ChangePasswordBySelfData struct {
	OldPassword string `json:"old_password" validate:"required,min=4,max=16"`
	NewPassword string `json:"new_password" validate:"required,min=4,max=16"`
}
