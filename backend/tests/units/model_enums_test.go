package units

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	telegramusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/telegram"
	ocservuser "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/users"
	"github.com/stretchr/testify/require"
)

func TestModelEnumsValidateAllowedValues(t *testing.T) {
	tests := []struct {
		name    string
		valid   func() bool
		invalid func() bool
	}{
		{"expiry mode", models.ExpiryModeFixed.IsValid, models.ExpiryMode("rolling").IsValid},
		{"traffic type", models.MonthlyRxTx.IsValid, models.TrafficType("Daily").IsValid},
		{"session event", models.EventHandshake.IsValid, models.OcservUserSessionEvent("login").IsValid},
		{"certificate status", models.OcservUserCertificateStatusActive.IsValid, models.OcservUserCertificateStatus("revoked").IsValid},
		{"telegram language", models.TelegramLanguageFA.IsValid, models.TelegramLanguage("de").IsValid},
		{"telegram request type", models.TelegramRequestTypeRenew.IsValid, models.TelegramRequestType("cancel").IsValid},
		{"telegram request status", models.TelegramRequestStatusDelivered.IsValid, models.TelegramRequestStatus("cancelled").IsValid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.True(t, test.valid())
			require.False(t, test.invalid())
		})
	}
}

func TestEnumModelHooksRejectUnsupportedValues(t *testing.T) {
	require.Error(t, (&models.OcservAgent{AddressType: models.AgentAddressType("socket")}).BeforeSave(nil))
	require.Error(t, (&models.OcservUser{
		ExpiryMode:  models.ExpiryMode("rolling"),
		TrafficType: models.Free,
	}).BeforeCreate(nil))
	require.Error(t, (&models.OcservUserSessionLog{
		Event: models.OcservUserSessionEvent("login"),
	}).BeforeSave(nil))
	require.Error(t, (&models.TelegramSettings{
		DefaultLanguage: models.TelegramLanguage("de"),
	}).BeforeCreate(nil))
	require.Error(t, (&models.TelegramAccount{
		Language: models.TelegramLanguage("de"),
	}).BeforeCreate(nil))
	require.Error(t, (&models.TelegramPackage{
		TrafficType: models.TrafficType("Daily"),
	}).BeforeCreate(nil))
	require.Error(t, (&models.TelegramRequest{
		Type:   models.TelegramRequestType("cancel"),
		Status: models.TelegramRequestStatusPending,
	}).BeforeCreate(nil))
	require.Error(t, (&models.TelegramRequest{
		Type:   models.TelegramRequestTypeNew,
		Status: models.TelegramRequestStatus("cancelled"),
	}).BeforeCreate(nil))

	// Map-based GORM updates use an empty model receiver. They must remain valid
	// while the repository validates the enum value contained in the update map.
	require.NoError(t, (&models.TelegramRequest{}).BeforeUpdate(nil))
}

func TestEnumAPIDTOValidationRejectsUnsupportedValues(t *testing.T) {
	validate := validator.New()

	user := ocservuser.CreateOcservUserData{
		Group:       "defaults",
		Username:    "enum-user",
		Password:    "secret",
		ExpiryMode:  models.ExpiryModeFixed,
		ExpireAt:    "2027-01-01",
		TrafficType: models.TrafficType("Daily"),
		Config:      &models.OcservUserConfig{},
	}
	require.Error(t, validate.Struct(user))

	language := models.TelegramLanguage("de")
	require.Error(t, validate.Struct(telegramusecase.PatchSettingsData{DefaultLanguage: &language}))

	plan := telegramusecase.CreatePackageData{
		Title:       "Invalid plan",
		Days:        30,
		TrafficType: models.TrafficType("Daily"),
	}
	require.Error(t, validate.Struct(plan))
}

func TestTypedEnumsPreserveJSONStrings(t *testing.T) {
	value := models.TelegramRequest{
		Type:   models.TelegramRequestTypeRenew,
		Status: models.TelegramRequestStatusAwaitingPayment,
	}
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	require.JSONEq(t, `{"id":0,"chat_id":0,"telegram_username":"","type":"renew","package_id":null,"target_ocserv_user_id":null,"desired_username":"","status":"awaiting_payment","receipt_file_path":"","user_message":"","admin_note":"","delivered_at":null,"awaiting_payment_message_id":null,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`, string(encoded))
}
