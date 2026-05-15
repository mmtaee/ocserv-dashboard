# Telegram Bot

## Features

### Customer Self-Service
- **Link existing account**: customers send their VPN username/password to the bot to associate it with their Telegram chat. One chat can manage many accounts.
- **Usage check**: customers see remaining quota, expiry date and online status for each linked account on demand.
- **Renewal requests**: customers pick a package and submit a renewal request. The admin is notified in Telegram.
- **New account orders**: new customers can order a fresh account by picking a package and a desired username.
- **Receipt-based payment workflow**: after admin approval the customer is asked to upload a payment receipt (photo). The admin reviews the receipt in the dashboard, confirms the payment and the bot automatically delivers the credentials (or extends the existing account).
- **Low-quota warnings**: customers receive an automatic warning message when remaining quota drops below a configurable threshold (default 200 MB).
- **Multi-language**: bot conversations and notifications are available in English and Persian, selectable per chat.

### Admin Dashboard Pages (under `Telegram` section in sidebar)
- **Settings**: enable/disable the bot, paste the BotFather token, set admin chat ID, low-quota threshold, default language and Ocserv host.
- **Packages**: CRUD for the plans customers can pick (title, days, traffic size, traffic type, price text, active flag).
- **Requests**: review pending requests, view uploaded receipts, approve, reject or confirm payment, with optional admin notes.
- **Linked accounts**: each Ocserv user detail page lists every Telegram chat linked to that account, with a one-click unlink action.

---

## 🤖 Configuring the Telegram Bot

The dashboard ships with an integrated Telegram bot service (`telegram_bot`) that runs alongside `api`, `log_stream` and `user_expiry`. Configuration lives entirely in the database — there is no need to edit `.env` or restart anything manually after a token change.

### Installation Options

During installation, you will be asked whether you want to install the Telegram bot service. You can also manually control this by setting `TELEGRAM_BOT_ENABLED` in your `.env` file:

- **Enable Telegram bot (default):** `TELEGRAM_BOT_ENABLED=true`
- **Disable Telegram bot:** `TELEGRAM_BOT_ENABLED=false`

For Docker deployments, we use a modular approach with multiple compose files:
- `docker-compose.yml` - base compose file with all services except Telegram bot
- `docker-compose-telegram.yml` - optional compose file that adds only the Telegram bot service

To use both together:
```bash
docker compose -f docker-compose.yml -f docker-compose-telegram.yml up -d
```

### Configuration Steps

1. Create a new bot with [@BotFather](https://t.me/BotFather) and copy the token.
2. Open the dashboard, navigate to **Telegram → Settings** and paste the token, set your admin chat ID, low-quota threshold, default language and the Ocserv host that customers will see when they receive new credentials.
3. Toggle **Bot enabled** on and save. Within ~30 seconds the bot service detects the change, connects to Telegram and starts polling for updates.
4. Define one or more sellable plans in **Telegram → Packages**.
5. Send `/start` to your bot in Telegram. Customer flows (link account / view usage / order new / renew / upload receipt) are now active.
6. Incoming requests appear in real time under **Telegram → Requests** where you approve them, review uploaded receipts and confirm payment. The bot delivers the resulting credentials (or renewal confirmation) automatically.

Receipt photos uploaded by customers are stored under `/opt/ocserv_dashboard/uploads/receipts/` (created by the installer with `0750` permissions).

---

## Custom Telegram Copy & Bot Metadata

You can override bundled English/Persian text without rebuilding binaries.

### API Service (HTML Messages to Customers)
- **Embedded defaults:** `services/api/internal/services/telegram/i18n/default.json` (`en`, `fa`, `ar`, `ru`, `zh-cn`, `zh-tw`, `it`)
- **Optional overlay:** set `TELEGRAM_I18N_PATH` to a JSON file with the same top-level keys. Values you omit keep the embedded default. Restart the API after changes.

Keys include notification templates such as `pkg_*`, `awaiting_*`, `rejected_*`, `new_account`, `renewal`, `support_suffix`, and related fragments used by `telegram/controller.go`.

### Standalone Telegram Bot (Conversation UI)
Bot menus, prompts, and button labels (everything under `services/telegram_bot/internal/i18n` used via `i18n.T`).

Same layout as the web dashboard: **one JSON file per language** under `services/telegram_bot/internal/i18n/locales/` (`en.json`, `fa.json`, `ar.json`, `ru.json`, `zh-cn.json`, `zh-tw.json`, `it.json`). Each file is a flat map of key → string. Keys match the `Key` constants in `i18n.go` (e.g. `welcome`, `btn_back`, `usage_text`). Missing keys fall back to English.

- **Optional overlay:** set `TELEGRAM_BOT_I18N_PATH` to a **directory** containing the same `*.json` files (merged over embedded defaults). Restart the bot after changes.

Supported codes are defined in `services/common/models/telegram_languages.go` (kept in sync with `VITE_I18N_LANGUAGES` / `web/src/locales/*.json`).

### Standalone Telegram Bot (BotFather Metadata)
- **Embedded defaults:** `services/telegram_bot/internal/bot/metadata_locales.json`
- **Optional overlay:** set `TELEGRAM_BOT_METADATA_LOCALES_PATH` to a JSON file with the same structure (`en` / `fa` objects with `commands`, `long_description`, `short_description`). Restart the bot after changes.

### Dashboard
The home dashboard shows a read-only snapshot of Telegram settings (`enabled`, whether a bot token is stored, optional `bot_username` from settings). It does not call the Telegram API.
