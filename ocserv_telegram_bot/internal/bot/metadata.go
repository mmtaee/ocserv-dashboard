package bot

import (
	_ "embed"
	"encoding/json"
	"os"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
)

//go:embed metadata_locales.json
var embeddedMetadataLocales []byte

type metadataLocalesFile struct {
	Commands          map[string][]metadataCmd `json:"commands"`
	LongDescriptions  map[string]string        `json:"long_descriptions"`
	ShortDescriptions map[string]string        `json:"short_descriptions"`
}

type metadataCmd struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

var (
	metadataLocales   metadataLocalesFile
	metadataLocalesMu sync.Once
)

func loadMetadataLocales() {
	metadataLocalesMu.Do(func() {
		if err := json.Unmarshal(embeddedMetadataLocales, &metadataLocales); err != nil {
			logger.Error("ocserv_telegram_bot: parse embedded metadata_locales.json: %v", err)
			return
		}
		if p := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_METADATA_LOCALES_PATH")); p != "" {
			if b, err := os.ReadFile(p); err == nil {
				var extra metadataLocalesFile
				if err := json.Unmarshal(b, &extra); err != nil {
					logger.Warn("ocserv_telegram_bot: optional metadata locales %s: %v", p, err)
					return
				}
				mergeMetadataLocales(&extra)
			}
		}
	})
}

func mergeMetadataLocales(extra *metadataLocalesFile) {
	if extra.Commands != nil {
		if metadataLocales.Commands == nil {
			metadataLocales.Commands = map[string][]metadataCmd{}
		}
		for lang, cmds := range extra.Commands {
			metadataLocales.Commands[strings.ToLower(lang)] = cmds
		}
	}
	if extra.LongDescriptions != nil {
		if metadataLocales.LongDescriptions == nil {
			metadataLocales.LongDescriptions = map[string]string{}
		}
		for k, v := range extra.LongDescriptions {
			metadataLocales.LongDescriptions[strings.ToLower(k)] = v
		}
	}
	if extra.ShortDescriptions != nil {
		if metadataLocales.ShortDescriptions == nil {
			metadataLocales.ShortDescriptions = map[string]string{}
		}
		for k, v := range extra.ShortDescriptions {
			metadataLocales.ShortDescriptions[strings.ToLower(k)] = v
		}
	}
}

func toBotCommands(cmds []metadataCmd) []tgbotapi.BotCommand {
	out := make([]tgbotapi.BotCommand, 0, len(cmds))
	for _, c := range cmds {
		out = append(out, tgbotapi.BotCommand{Command: c.Command, Description: c.Description})
	}
	return out
}

// applyBotMetadata pushes a localised set of commands, descriptions and the
// default menu button to BotFather every time the bot connects with a new
// token. Strings are loaded from metadata_locales.json (embedded); override or
// extend with TELEGRAM_BOT_METADATA_LOCALES_PATH. See docs/telegram-translations.md.
//
// All calls are best-effort and idempotent. Telegram silently ignores updates
// that match the current value, so re-applying on every (re)connect is safe.
func applyBotMetadata(api *tgbotapi.BotAPI) {
	if api == nil {
		return
	}
	loadMetadataLocales()

	for lang, cmds := range metadataLocales.Commands {
		if len(cmds) == 0 {
			continue
		}
		bc := toBotCommands(cmds)
		cfg := tgbotapi.NewSetMyCommandsWithScopeAndLanguage(
			tgbotapi.NewBotCommandScopeAllPrivateChats(), lang, bc...,
		)
		if _, err := api.Request(cfg); err != nil {
			logger.Warn("ocserv_telegram_bot: SetMyCommands(%s) failed: %v", lang, err)
		}
	}

	defaultCmds := toBotCommands(metadataLocales.Commands["en"])
	if len(defaultCmds) > 0 {
		if _, err := api.Request(tgbotapi.NewSetMyCommands(defaultCmds...)); err != nil {
			logger.Warn("ocserv_telegram_bot: SetMyCommands(default) failed: %v", err)
		}
	}

	for lang, desc := range metadataLocales.LongDescriptions {
		params := tgbotapi.Params{
			"description":   desc,
			"language_code": lang,
		}
		if _, err := api.MakeRequest("setMyDescription", params); err != nil {
			logger.Warn("ocserv_telegram_bot: setMyDescription(%s) failed: %v", lang, err)
		}
	}

	for lang, short := range metadataLocales.ShortDescriptions {
		params := tgbotapi.Params{
			"short_description": short,
			"language_code":     lang,
		}
		if _, err := api.MakeRequest("setMyShortDescription", params); err != nil {
			logger.Warn("ocserv_telegram_bot: setMyShortDescription(%s) failed: %v", lang, err)
		}
	}

	menuParams := tgbotapi.Params{
		"menu_button": `{"type":"commands"}`,
	}
	if _, err := api.MakeRequest("setChatMenuButton", menuParams); err != nil {
		logger.Warn("ocserv_telegram_bot: setChatMenuButton failed: %v", err)
	}
}
