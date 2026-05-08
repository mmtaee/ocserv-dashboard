package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
)

// applyBotMetadata pushes a localised set of commands, descriptions and the
// default menu button to BotFather every time the bot connects with a new
// token. This is what powers:
//
//   - The "/" command picker that appears as users type, in both EN and FA.
//   - The "What can this bot do?" intro box that shows on first launch.
//   - The short profile description used in the bot's profile page and when
//     the bot is shared.
//   - The menu button next to the input field, which we lock to the commands
//     list so users always have one tap access to /start, /help, /settings…
//
// All calls are best-effort and idempotent. Telegram silently ignores updates
// that match the current value, so re-applying on every (re)connect is safe.
func applyBotMetadata(api *tgbotapi.BotAPI) {
	if api == nil {
		return
	}

	commandsByLang := map[string][]tgbotapi.BotCommand{
		"en": {
			{Command: "start", Description: "Open the main menu"},
			{Command: "help", Description: "Show help and supported commands"},
			{Command: "settings", Description: "Bot settings"},
			{Command: "language", Description: "Change bot language"},
			{Command: "cancel", Description: "Cancel the current operation"},
		},
		"fa": {
			{Command: "start", Description: "نمایش منوی اصلی"},
			{Command: "help", Description: "راهنما و دستورها"},
			{Command: "settings", Description: "تنظیمات ربات"},
			{Command: "language", Description: "تغییر زبان"},
			{Command: "cancel", Description: "لغو عملیات فعلی"},
		},
	}
	for lang, cmds := range commandsByLang {
		cfg := tgbotapi.NewSetMyCommandsWithScopeAndLanguage(
			tgbotapi.NewBotCommandScopeAllPrivateChats(), lang, cmds...,
		)
		if _, err := api.Request(cfg); err != nil {
			logger.Warn("telegram_bot: SetMyCommands(%s) failed: %v", lang, err)
		}
	}

	// Default scope (no language) — fallback for any unknown locale.
	defaultCmds := commandsByLang["en"]
	if _, err := api.Request(tgbotapi.NewSetMyCommands(defaultCmds...)); err != nil {
		logger.Warn("telegram_bot: SetMyCommands(default) failed: %v", err)
	}

	// setMyDescription / setMyShortDescription / setChatMenuButton aren't
	// surfaced as helpers in telegram-bot-api v5.5.1, so we hit the raw
	// methods through MakeRequest. The Bot API returns 400 on no-op writes
	// after an identical call, which is harmless.
	descriptions := map[string]string{
		"en": "Manage your Ocserv VPN account directly from Telegram. Check your remaining quota and expiry, request renewals, order new accounts and upload payment receipts. You will be alerted automatically when your traffic runs low.",
		"fa": "اکانت VPN خود را مستقیم از تلگرام مدیریت کنید: مشاهدهٔ مصرف و تاریخ انقضا، درخواست تمدید، سفارش اکانت جدید و ارسال تصویر رسید پرداخت. هنگام کاهش حجم به‌صورت خودکار مطلع می‌شوید.",
	}
	for lang, desc := range descriptions {
		params := tgbotapi.Params{
			"description":   desc,
			"language_code": lang,
		}
		if _, err := api.MakeRequest("setMyDescription", params); err != nil {
			logger.Warn("telegram_bot: setMyDescription(%s) failed: %v", lang, err)
		}
	}

	shortDescriptions := map[string]string{
		"en": "Self-service Ocserv VPN account management on Telegram.",
		"fa": "مدیریت سلف-سرویس اکانت VPN از طریق تلگرام.",
	}
	for lang, short := range shortDescriptions {
		params := tgbotapi.Params{
			"short_description": short,
			"language_code":     lang,
		}
		if _, err := api.MakeRequest("setMyShortDescription", params); err != nil {
			logger.Warn("telegram_bot: setMyShortDescription(%s) failed: %v", lang, err)
		}
	}

	// Lock the menu button to the canonical commands list so users always
	// see /start, /help, /settings… one tap away from the input field.
	menuParams := tgbotapi.Params{
		"menu_button": `{"type":"commands"}`,
	}
	if _, err := api.MakeRequest("setChatMenuButton", menuParams); err != nil {
		logger.Warn("telegram_bot: setChatMenuButton failed: %v", err)
	}
}
