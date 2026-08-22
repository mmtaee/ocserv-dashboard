package models

import "strings"

// Language codes are in telegram.go (TelegramLanguageEN, …). This file lists them for UI/bot pickers.

// TelegramLangCallbackPrefix is the inline-keyboard prefix for language selection (lang:en, lang:fa, …).
const TelegramLangCallbackPrefix = "lang:"

// TelegramLanguageOption is one supported bot/dashboard Telegram language.
type TelegramLanguageOption struct {
	Code  string
	Label string
}

// TelegramLanguages lists supported languages in display order (same set as VITE_I18N_LANGUAGES).
var TelegramLanguages = []TelegramLanguageOption{
	{Code: TelegramLanguageEN, Label: "English"},
	{Code: TelegramLanguageFA, Label: "فارسی"},
	{Code: TelegramLanguageAR, Label: "العربية"},
	{Code: TelegramLanguageRU, Label: "Русский"},
	{Code: TelegramLanguageZHCN, Label: "中文(简体)"},
	{Code: TelegramLanguageZHTW, Label: "中文(繁體)"},
	{Code: TelegramLanguageIT, Label: "Italiano"},
}

var telegramLanguageCodes map[string]struct{}

func init() {
	telegramLanguageCodes = make(map[string]struct{}, len(TelegramLanguages))
	for _, l := range TelegramLanguages {
		telegramLanguageCodes[strings.ToLower(l.Code)] = struct{}{}
	}
}

// IsTelegramLanguage reports whether code is a supported Telegram language.
func IsTelegramLanguage(code string) bool {
	_, ok := telegramLanguageCodes[strings.ToLower(strings.TrimSpace(code))]
	return ok
}

// IsTelegramRTL reports whether bot/API HTML messages should use RTL formatting for this language.
func IsTelegramRTL(code string) bool {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case TelegramLanguageFA, TelegramLanguageAR:
		return true
	default:
		return false
	}
}

// TelegramLangCallback builds callback_data for a language button.
func TelegramLangCallback(code string) string {
	return TelegramLangCallbackPrefix + strings.ToLower(strings.TrimSpace(code))
}

// TelegramLangFromCallback parses callback_data; second return is false if invalid.
func TelegramLangFromCallback(data string) (string, bool) {
	if !strings.HasPrefix(data, TelegramLangCallbackPrefix) {
		return "", false
	}
	code := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(data, TelegramLangCallbackPrefix)))
	if !IsTelegramLanguage(code) {
		return "", false
	}
	return code, true
}
