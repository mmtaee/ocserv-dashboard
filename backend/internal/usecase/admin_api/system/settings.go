package system

import (
	"context"
	"errors"
	"strings"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	ocservuser "github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/user"
	"gorm.io/gorm"
)

func (u *Usecase) Init(ctx context.Context) (*GetSystemInitResponse, error) {
	settings, err := u.systems.System(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &GetSystemInitResponse{
		GoogleCaptchaSiteKey: settings.GoogleCaptchaSiteKey,
		TelegramBotEnabled:   u.telegramEnabled,
		FirstInit:            settings.FirstInit,
	}, nil
}

func (u *Usecase) Settings(ctx context.Context) (*GetSystemResponse, error) {
	settings, err := u.systems.System(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return systemResponse(settings), nil
}

func (u *Usecase) Update(ctx context.Context, userID uint, input PatchSystemUpdateData) (*GetSystemResponse, error) {
	settings := &models.System{}
	if input.GoogleCaptchaSiteKey != nil {
		settings.GoogleCaptchaSiteKey = *input.GoogleCaptchaSiteKey
	}
	if input.GoogleCaptchaSecretKey != nil {
		settings.GoogleCaptchaSecretKey = *input.GoogleCaptchaSecretKey
	}
	if input.AutoDeleteInactiveUsers != nil {
		settings.AutoDeleteInactiveUsers = *input.AutoDeleteInactiveUsers
	}
	if input.KeepInactiveUserDays != nil {
		settings.KeepInactiveUserDays = *input.KeepInactiveUserDays
		if settings.KeepInactiveUserDays < 1 {
			settings.KeepInactiveUserDays = 1
		}
	}
	var err error
	if input.ClientProfileServerAddress != nil {
		settings.ClientProfileServerAddress = strings.TrimSpace(*input.ClientProfileServerAddress)
		if settings.ClientProfileServerAddress != "" {
			_, err = ocservuser.NormalizeProfileServerAddress(settings.ClientProfileServerAddress)
		}
	}
	if err == nil && input.ClientProfileServerPort != nil {
		settings.ClientProfileServerPort = *input.ClientProfileServerPort
		_, err = ocservuser.NormalizeProfileServerPort(settings.ClientProfileServerPort)
	}
	if err == nil && input.ClientProfileConnectionName != nil {
		settings.ClientProfileConnectionName = strings.TrimSpace(*input.ClientProfileConnectionName)
		if settings.ClientProfileConnectionName != "" {
			_, err = ocservuser.NormalizeProfileConnectionName(settings.ClientProfileConnectionName)
		}
	}
	if err != nil {
		return nil, err
	}
	settings.FirstInit = true
	ctx = context.WithValue(ctx, "userID", userID)
	updated, err := u.systems.SystemUpdate(ctx, settings)
	if err != nil {
		return nil, err
	}
	return systemResponse(updated), nil
}

func systemResponse(settings *models.System) *GetSystemResponse {
	return &GetSystemResponse{
		FirstInit:            settings.FirstInit,
		GoogleCaptchaSiteKey: settings.GoogleCaptchaSiteKey, GoogleCaptchaSecretKey: settings.GoogleCaptchaSecretKey,
		AutoDeleteInactiveUsers: settings.AutoDeleteInactiveUsers, KeepInactiveUserDays: settings.KeepInactiveUserDays,
		ClientProfileServerAddress: settings.ClientProfileServerAddress, ClientProfileServerPort: settings.ClientProfileServerPort,
		ClientProfileConnectionName: settings.ClientProfileConnectionName,
	}
}
