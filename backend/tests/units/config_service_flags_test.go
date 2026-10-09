package units

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	adminapi "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api"
	"github.com/stretchr/testify/require"
)

func TestServiceFlagDefaultsAndBooleanParsing(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_ENABLED", "")
	t.Setenv("CUSTOMER_API_ENABLED", "")
	config.Init(false, "", 0)
	require.False(t, config.Get().TelegramEnabled)
	require.True(t, config.Get().CustomerAPIEnabled)

	t.Setenv("TELEGRAM_BOT_ENABLED", "1")
	t.Setenv("CUSTOMER_API_ENABLED", "false")
	config.Init(false, "", 0)
	require.True(t, config.Get().TelegramEnabled)
	require.False(t, config.Get().CustomerAPIEnabled)

	t.Setenv("TELEGRAM_BOT_ENABLED", "")
	t.Setenv("CUSTOMER_API_ENABLED", "")
	config.Init(false, "", 0)
}

func TestConfigValidateRequiresStrongSecretKey(t *testing.T) {
	for _, secretKey := range []string{"", "too-short"} {
		t.Setenv("SECRET_KEY", secretKey)
		config.Init(false, "", 0)
		require.Error(t, config.Validate())
	}

	t.Setenv("SECRET_KEY", "0123456789abcdef0123456789abcdef")
	config.Init(false, "", 0)
	require.NoError(t, config.Validate())
}

func TestTelegramRoutesFollowServiceFlag(t *testing.T) {
	t.Setenv("AGENT_NODE", "true")
	config.Init(false, "", 0)
	for _, enabled := range []bool{false, true} {
		service, err := adminapi.New(enabled, false)
		require.NoError(t, err)
		e := echo.New()
		service.Register(e.Group(""))
		found := false
		for _, route := range e.Router().Routes() {
			found = found || route.Path == "/telegram/settings"
		}
		require.Equal(t, enabled, found)
	}
}
