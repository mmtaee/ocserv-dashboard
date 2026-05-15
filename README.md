# OpenConnect VPN Server (Ocserv) with Dashboard

A simple, efficient, and scalable solution to deploy and manage an **OpenConnect VPN server (ocserv)**
with a powerful **web-based dashboard**.  
Easily manage users, groups, and server configurations while keeping your VPN secure and performant.

<p align="center">
  <img alt="Project Logo" src="docs/logo.png" width="800"/>
</p>

<p align="center">
  <img alt="GitHub stars" src="https://img.shields.io/github/stars/mmtaee/ocserv-dashboard">
  <img alt="GitHub forks" src="https://img.shields.io/github/forks/mmtaee/ocserv-dashboard">
  <img alt="GitHub issues" src="https://img.shields.io/github/issues/mmtaee/ocserv-dashboard">
  <img alt="GitHub contributors" src="https://img.shields.io/github/contributors/mmtaee/ocserv-dashboard">
  <img alt="Repo size" src="https://img.shields.io/github/repo-size/mmtaee/ocserv-dashboard">
</p>

<p align="center">
  <img alt="Dashboard Home Page Preview" src="docs/home.png" width="800"/>
  <img alt="Project Logo" src="docs/home_stats.png" width="800"/>
  <br>
  <i>Dashboard UI Preview</i>
</p>

---

## 🌟 Key Features

### 1. Ocserv User Management
- Create, update, remove, block, and disconnect users with ease.
- Sync the `ocpasswd` file with the database to keep user credentials consistent.
- Set traffic usage limits per user (e.g., GB or monthly quotas).
- Manage account expiration to automatically deactivate users when their subscription ends.

### 2. Ocserv Group Management
- Create, update, and delete user groups.
- Sync the `/etc/ocserv/groups/*` files with the database to ensure consistent group configurations.
- Organize users into logical groups for easier management.

### 3. Ocserv Command-Line Tools
- Use the `occtl` CLI utility to perform various server operations efficiently.

### 4. Ocserv User Statistics & Monitoring
- View real-time statistics for user traffic (RX/TX).
- Track data usage per user and per group.

### 5. Ocserv Live Server Logs
- Monitor Ocserv logs in real-time directly from the web dashboard.

### 6. Staffs and Staff Management
- Manage admin accounts: create, update, delete, and reset passwords.
- Track staff activities and administrative actions for accountability.
- Each staff member can create and manage **their own Ocserv Users and Groups**. 
  Staff members cannot view or modify users/groups created by others;  
  only admin users have full access.

### 7. Customer Account Details & Usage
- View detailed customer account information.
- Monitor user-specific usage summaries and traffic data.

### 8. Internationalization (i18n)
- Multi-language support:
  - English (**en**)
  - Russian (**ru**)
  - Simplified Chinese (**zh-cn**)
  - Traditional Chinese (**zh-tw**)
  - Arabic (**ar**)
  - Persian (**fa**)

### 9. Telegram Bot
- **Customer self-service**:
  - Link VPN accounts, check usage/expiry, request renewals, order new accounts, and upload payment receipts
  - Low-quota warnings and multi-language support
- **Admin dashboard**: Manage settings, packages, requests, and linked accounts
- **Customization**: Override translations and bot metadata via environment variables (see [docs/telegram_bot.md](docs/telegram_bot.md))

---

## ⚠️ Legacy Version Note

- **Branch name:** [legacy](https://github.com/mmtaee/ocserv-dashboard/tree/legacy)
- **Old version:** Developed using **Python backend** with **Vue 2 frontend**.
- **Features:** Minimal, limited functionality compared to the current version — only basic user and group management existed.

---

## ⚙️ System Requirements

- **Docker-based:**
  - [Docker v28.5 or higher](https://docs.docker.com/engine/install/)
  - [Docker Compose v2.40 or higher](https://docs.docker.com/compose/install/)

- **Systemd-based:**
  - **Supported Operating Systems:**
    - [Debian 12 or higher](https://www.debian.org/download)
    - [Ubuntu 20.04 or higher](https://ubuntu.com/download/server)

  - **Programming Language:**
    - [Golang v1.25 or higher](https://go.dev/dl/)

---

## 🚀 Quick Start

1. Clone the repository:
```bash
git clone https://github.com/mmtaee/ocserv-dashboard.git

cd ocserv-dashboard

chmod +x install.sh

./install.sh
```
then select an option to continue:
<p>
  <img alt="Installation Menu" src="docs/menu.png" width="800"/>
</p>

---

## 🌐 Access the Admin Dashboard

1. Open your web browser.
2. Navigate to `https://YOUR-DOMAIN-OR-IP:3443` in the browser.
3. Complete the administrative setup wizard.
4. Start managing users, groups, and VPN settings from the dashboard.

---

## 🌐 Access the Customers page for quick insights

1. Open your web browser.
2. Navigate to `https://YOUR-DOMAIN-OR-IP:3443/summary/` in the browser.
3. Enter your Ocserv username and password to see insights.

---

## 🔒 Security & Scalability

- Designed with **best practices for security** to ensure a safe and reliable VPN environment.
- The web panel is intuitive and easy to use for both administrators and end users.
- Scalable architecture allows efficient management of multiple users and groups.
- Real-time usage tracking and monitoring built-in.
- If you encounter any issues, please refer to the documentation or contact support.

---

## 🧭 Roadmap / TODO

The planned features and upcoming improvements are tracked in the **[TODO.md](TODO.md)** file.

Check it out to see what's coming next!

---

## 🌍 Contributing to Translations (i18n)

We welcome community contributions to improve and expand internationalization (i18n) support! Here's how you can help:

### 📁 Where are the translation files?

This project has 3 main parts that need translations:

1. **Web Dashboard** → [web/src/locales/](https://github.com/mmtaee/ocserv-dashboard/tree/master/web/src/locales)
   - One JSON file per language (e.g., `en.json`, `fa.json`, `es.json`)
   
2. **Telegram Bot** (3 sub-locations):
   - **Bot conversation UI**: [services/telegram_bot/internal/i18n/locales/](https://github.com/mmtaee/ocserv-dashboard/tree/master/services/telegram_bot/internal/i18n/locales)
   - **API notification messages**: [services/api/internal/services/telegram/i18n/default.json](https://github.com/mmtaee/ocserv-dashboard/tree/master/services/api/internal/services/telegram/i18n/default.json)
   - **BotFather metadata**: [services/telegram_bot/internal/bot/metadata_locales.json](https://github.com/mmtaee/ocserv-dashboard/tree/master/services/telegram_bot/internal/bot/metadata_locales.json)

### 🛠️ Step-by-Step Guide to Add/Improve a Translation

#### Case 1: Improve an existing translation
1. Open the JSON file for that language in any of the locations above
2. Edit the values (keep the keys exactly as they are!)
3. Save and submit your changes

#### Case 2: Add a completely new language (e.g., Spanish, `es`)
You need to update **all 4 locations**!

1. **Web Dashboard**:
   - Copy `web/src/locales/en.json` → `web/src/locales/es.json`
   - Translate all values to Spanish

2. **Telegram Bot Conversation UI**:
   - Copy `services/telegram_bot/internal/i18n/locales/en.json` → `services/telegram_bot/internal/i18n/locales/es.json`
   - Translate all values to Spanish

3. **Telegram API Notifications**:
   - Open `services/api/internal/services/telegram/i18n/default.json`
   - Add your language code (`es`) to every object with the translated text

4. **Telegram BotFather Metadata**:
   - Open `services/telegram_bot/internal/bot/metadata_locales.json`
   - Add your language code (`es`) with translated `commands`, `long_description`, and `short_description`

5. **Update supported languages list**:
   - Open `services/common/models/telegram_languages.go`
   - Add your new language to the list
   
6. **Update the Installer**:
   - Open [install.sh](https://github.com/mmtaee/ocserv-dashboard/blob/master/install.sh)
   - Find the `LANGUAGES=` line and add your language in `code:Name` format
   - Example: `LANGUAGES=en:English,it:Italiano,zh-tw:中文(繁體),zh-cn:中文(简体),ru:Русский,fa:فارسی,ar:العربية,es:Español`

### ✅ Tips for Good Translations
- Keep keys unchanged (only translate the values!)
- Use valid JSON syntax (use a JSON validator if needed)
- Maintain the same tone and style as existing translations
- Test your translations if possible!

---

## 📦 License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

---
## 📈 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=mmtaee/ocserv-dashboard&type=Date)](https://www.star-history.com/#mmtaee/ocserv-dashboard&Date)
