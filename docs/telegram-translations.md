# Telegram strings and bot metadata

You can override bundled English/Persian text without rebuilding binaries.

## API service (HTML messages to customers)

- **Embedded defaults:** `services/api/internal/services/telegram/i18n/default.json` (`en`, `fa`, `ar`, `ru`, `zh-cn`, `zh-tw`, `it`)
- **Optional overlay:** set `TELEGRAM_I18N_PATH` to a JSON file with the same top-level keys. Values you omit keep the embedded default. Restart the API after changes.

Keys include notification templates such as `pkg_*`, `awaiting_*`, `rejected_*`, `new_account`, `renewal`, `support_suffix`, and related fragments used by `telegram/controller.go`.

## Standalone Telegram bot (conversation UI)

Bot menus, prompts, and button labels (everything under `services/telegram_bot/internal/i18n` used via `i18n.T`).

Same layout as the web dashboard: **one JSON file per language** under `services/telegram_bot/internal/i18n/locales/` (`en.json`, `fa.json`, `ar.json`, `ru.json`, `zh-cn.json`, `zh-tw.json`, `it.json`). Each file is a flat map of key → string. Keys match the `Key` constants in `i18n.go` (e.g. `welcome`, `btn_back`, `usage_text`). Missing keys fall back to English.

- **Optional overlay:** set `TELEGRAM_BOT_I18N_PATH` to a **directory** containing the same `*.json` files (merged over embedded defaults). Restart the bot after changes.

Supported codes are defined in `services/common/models/telegram_languages.go` (kept in sync with `VITE_I18N_LANGUAGES` / `web/src/locales/*.json`).

## Standalone Telegram bot (BotFather metadata)

- **Embedded defaults:** `services/telegram_bot/internal/bot/metadata_locales.json`
- **Optional overlay:** set `TELEGRAM_BOT_METADATA_LOCALES_PATH` to a JSON file with the same structure (`en` / `fa` objects with `commands`, `long_description`, `short_description`). Restart the bot after changes.

## Dashboard

The home dashboard shows a read-only snapshot of Telegram settings (`enabled`, whether a bot token is stored, optional `bot_username` from settings). It does not call the Telegram API.
