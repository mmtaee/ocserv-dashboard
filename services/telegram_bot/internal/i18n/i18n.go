package i18n

import "github.com/mmtaee/ocserv-dashboard/common/models"

// Key is the catalog of translatable bot strings. Adding a new string requires
// providing translations in every language map below.
type Key string

const (
	Welcome           Key = "welcome"
	BotDisabled       Key = "bot_disabled"
	MainMenu          Key = "main_menu"
	BtnAddAccount     Key = "btn_add_account"
	BtnMyAccounts     Key = "btn_my_accounts"
	BtnNewOrder       Key = "btn_new_order"
	BtnHelp           Key = "btn_help"
	BtnLanguage       Key = "btn_language"
	BtnCancel         Key = "btn_cancel"
	BtnBack           Key = "btn_back"
	BtnUsage          Key = "btn_usage"
	BtnRenew          Key = "btn_renew"
	BtnRemove         Key = "btn_remove"
	AskUsername       Key = "ask_username"
	AskPassword       Key = "ask_password"
	AskUsernameNew    Key = "ask_username_new"
	AskMessage        Key = "ask_message"
	AskReceipt        Key = "ask_receipt"
	AuthSuccess       Key = "auth_success"
	AuthFail          Key = "auth_fail"
	AuthLocked        Key = "auth_locked"
	AlreadyLinked     Key = "already_linked"
	NoAccounts        Key = "no_accounts"
	NoPackages        Key = "no_packages"
	PickPackage       Key = "pick_package"
	PickAccountRenew  Key = "pick_account_renew"
	RequestCreated    Key = "request_created"
	RequestExists     Key = "request_exists"
	WaitForApproval   Key = "wait_for_approval"
	NotApprovedYet    Key = "not_approved_yet"
	ReceiptSaved      Key = "receipt_saved"
	OnlyPhoto         Key = "only_photo"
	HelpText          Key = "help_text"
	UsageText         Key = "usage_text"
	AccountRemoved    Key = "account_removed"
	NotLinked         Key = "not_linked"
	UnknownCommand    Key = "unknown_command"
	LowQuotaWarning   Key = "low_quota_warning"
	LanguagePicked    Key = "language_picked"
	SessionTimedOut   Key = "session_timed_out"
	OcservDeactivated Key = "ocserv_deactivated"
	RateLimited       Key = "rate_limited"
)

var catalog = map[string]map[Key]string{
	models.TelegramLanguageEN: en,
	models.TelegramLanguageFA: fa,
}

// T returns the translation for the given language, falling back to English
// when the language is missing or the key is not translated.
func T(lang string, key Key, args ...interface{}) string {
	if lang == "" {
		lang = models.TelegramLanguageEN
	}
	bundle, ok := catalog[lang]
	if !ok {
		bundle = en
	}
	value, ok := bundle[key]
	if !ok {
		if fallback, ok2 := en[key]; ok2 {
			value = fallback
		} else {
			value = string(key)
		}
	}
	if len(args) == 0 {
		return value
	}
	return sprintf(value, args...)
}

// sprintf is a tiny wrapper over fmt.Sprintf isolated for testability.
func sprintf(format string, args ...interface{}) string {
	return fmtSprintf(format, args...)
}
