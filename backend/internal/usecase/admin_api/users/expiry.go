package ocservuser

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

var ErrInvalidExpiryConfiguration = errors.New("invalid expiry configuration")

func configureNewExpiry(account *models.OcservUser, input CreateOcservUserData, now time.Time) error {
	mode := models.ExpiryMode(strings.TrimSpace(string(input.ExpiryMode)))
	if mode == "" {
		switch {
		case input.Unlimited:
			mode = models.ExpiryModeUnlimited
		case input.ExpireDaysAfterFirstConnection != nil:
			mode = models.ExpiryModeFirstConnection
		case input.ExpireAt != "":
			mode = models.ExpiryModeFixed
		default:
			mode = models.ExpiryModeFixed
			defaultExpiry := now.UTC().AddDate(0, 0, 30)
			account.ExpiryMode = mode
			account.ExpireAt = &defaultExpiry
			return nil
		}
	}

	var expireAt *string
	if input.ExpireAt != "" {
		expireAt = &input.ExpireAt
	}
	return setExpiryConfiguration(account, mode, expireAt, input.ExpireDaysAfterFirstConnection, input.Unlimited, false)
}

func applyExpiryUpdate(account *models.OcservUser, input UpdateOcservUserData) error {
	if input.ExpiryMode == nil && input.ExpireAt == nil && input.ExpireDaysAfterFirstConnection == nil && !input.Unlimited && !input.ResetFirstConnection {
		return nil
	}

	mode := models.ExpiryMode("")
	if input.ExpiryMode != nil {
		mode = models.ExpiryMode(strings.TrimSpace(string(*input.ExpiryMode)))
		if mode == "" {
			return fmt.Errorf("%w: expiry_mode cannot be empty", ErrInvalidExpiryConfiguration)
		}
	} else {
		switch {
		case input.Unlimited:
			mode = models.ExpiryModeUnlimited
		case input.ExpireDaysAfterFirstConnection != nil || input.ResetFirstConnection:
			mode = models.ExpiryModeFirstConnection
		case input.ExpireAt != nil:
			mode = models.ExpiryModeFixed
		}
	}

	return setExpiryConfiguration(
		account,
		mode,
		input.ExpireAt,
		input.ExpireDaysAfterFirstConnection,
		input.Unlimited,
		input.ResetFirstConnection,
	)
}

func setExpiryConfiguration(
	account *models.OcservUser,
	mode models.ExpiryMode,
	expireAt *string,
	days *int,
	unlimited bool,
	resetFirstConnection bool,
) error {
	if unlimited && mode != models.ExpiryModeUnlimited {
		return fmt.Errorf("%w: unlimited conflicts with expiry_mode %q", ErrInvalidExpiryConfiguration, mode)
	}

	switch mode {
	case models.ExpiryModeUnlimited:
		if expireAt != nil || days != nil || resetFirstConnection {
			return fmt.Errorf("%w: unlimited mode cannot include expiry date, first-connection days, or reset", ErrInvalidExpiryConfiguration)
		}
		account.ExpiryMode = models.ExpiryModeUnlimited
		account.ExpireAt = nil
		account.ExpireDaysAfterFirstConnection = nil
		account.FirstConnectedAt = nil
		return nil

	case models.ExpiryModeFixed:
		if days != nil || resetFirstConnection {
			return fmt.Errorf("%w: fixed mode cannot include first-connection options", ErrInvalidExpiryConfiguration)
		}
		if expireAt == nil || strings.TrimSpace(*expireAt) == "" {
			return fmt.Errorf("%w: expire_at is required for fixed mode", ErrInvalidExpiryConfiguration)
		}
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*expireAt))
		if err != nil {
			return fmt.Errorf("%w: expire_at must use YYYY-MM-DD", ErrInvalidExpiryConfiguration)
		}
		account.ExpiryMode = models.ExpiryModeFixed
		account.ExpireAt = &parsed
		account.ExpireDaysAfterFirstConnection = nil
		account.FirstConnectedAt = nil
		return nil

	case models.ExpiryModeFirstConnection:
		if expireAt != nil || unlimited {
			return fmt.Errorf("%w: first_connection mode cannot include expire_at or unlimited", ErrInvalidExpiryConfiguration)
		}
		configuredDays := account.ExpireDaysAfterFirstConnection
		if days != nil {
			configuredDays = days
		}
		if configuredDays == nil || *configuredDays < 1 {
			return fmt.Errorf("%w: expire_days_after_first_connection must be greater than zero", ErrInvalidExpiryConfiguration)
		}
		if account.ExpiryMode == models.ExpiryModeFirstConnection && account.FirstConnectedAt != nil && days != nil &&
			account.ExpireDaysAfterFirstConnection != nil && *days != *account.ExpireDaysAfterFirstConnection && !resetFirstConnection {
			return fmt.Errorf("%w: reset_first_connection is required to change days after the first connection", ErrInvalidExpiryConfiguration)
		}
		if account.ExpiryMode != models.ExpiryModeFirstConnection || resetFirstConnection {
			account.FirstConnectedAt = nil
			account.ExpireAt = nil
		}
		account.ExpiryMode = models.ExpiryModeFirstConnection
		account.ExpireDaysAfterFirstConnection = configuredDays
		return nil

	default:
		return fmt.Errorf("%w: unsupported expiry_mode %q", ErrInvalidExpiryConfiguration, mode)
	}
}
