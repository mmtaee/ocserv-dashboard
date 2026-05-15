// Package i18n holds Telegram bot UI strings loaded from JSON locale files (one file per language,
// like web/src/locales). Override or extend with TELEGRAM_BOT_I18N_PATH pointing to a directory
// containing the same *.json files. See docs/telegram-translations.md.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mmtaee/ocserv-dashboard/common/models"
)

//go:embed locales/*.json
var embeddedLocales embed.FS

// Key is the catalog of translatable bot strings. Adding a new string requires a key here
// and an entry in each locales/<code>.json (and optional TELEGRAM_BOT_I18N_PATH overlay).
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
	LinkedLockedHint  Key = "linked_locked_hint"
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

	AdminWelcome     Key = "admin_welcome"
	AdminMenu        Key = "admin_menu"
	BtnAdminPending  Key = "btn_admin_pending"
	BtnAdminReceipts Key = "btn_admin_receipts"
	BtnAdminStats    Key = "btn_admin_stats"
	BtnAdminUserView Key = "btn_admin_user_view"
	BtnAdminBack     Key = "btn_admin_back"
	BtnOpenPanel     Key = "btn_open_panel"
	AdminNoPending   Key = "admin_no_pending"
	AdminNoReceipts  Key = "admin_no_receipts"
	AdminStatsText   Key = "admin_stats_text"
	AdminRequestRow  Key = "admin_request_row"
)

var (
	mu    sync.RWMutex
	store map[string]map[string]string
	once  sync.Once
)

// Init loads embedded locale files and optional TELEGRAM_BOT_I18N_PATH directory merge.
func Init() {
	once.Do(func() {
		store = make(map[string]map[string]string)
		if err := loadLocaleFS(embeddedLocales, "locales"); err != nil {
			panic("i18n: embedded locales: " + err.Error())
		}
		if p := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_I18N_PATH")); p != "" {
			_ = loadLocaleDir(p)
		}
	})
}

func loadLocaleFS(fsys fs.FS, root string) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			return nil
		}
		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		lang := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		return mergeLangFile(lang, b)
	})
}

func loadLocaleDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		lang := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		_ = mergeLangFile(lang, b)
	}
	return nil
}

func mergeLangFile(lang string, b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	lang = strings.ToLower(strings.TrimSpace(lang))
	mu.Lock()
	defer mu.Unlock()
	if store[lang] == nil {
		store[lang] = make(map[string]string)
	}
	for k, v := range m {
		store[lang][k] = v
	}
	return nil
}

// T returns the translation for the given language, falling back to English
// when the language is missing or the key is not translated.
func T(lang string, key Key, args ...interface{}) string {
	Init()
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		lang = models.TelegramLanguageEN
	}
	k := string(key)

	mu.RLock()
	defer mu.RUnlock()

	value, ok := lookup(lang, k)
	if !ok && lang != models.TelegramLanguageEN {
		value, ok = lookup(models.TelegramLanguageEN, k)
	}
	if !ok {
		return k
	}
	if len(args) == 0 {
		return value
	}
	return fmt.Sprintf(value, args...)
}

func lookup(lang, key string) (string, bool) {
	bundle, ok := store[lang]
	if !ok {
		return "", false
	}
	v, ok := bundle[key]
	return v, ok
}
