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
  <br>
  <i>Dashboard UI Preview</i>
</p>

---

## 🌟 Key Features

### 1. Ocserv User Management
- Create, update, remove, block, and disconnect users.
- Set traffic usage limits (e.g., GB or monthly usage).

### 2. Ocserv Group Management
- Create, update, and delete user groups.
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
  - Chinese (**zh**)
  - Arabic (**ar**)
  - Persian (**fa**)

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

We welcome community contributions to improve and expand internationalization (i18n) support!

### 📁 Translation Files Directory
All web dashboard translation files are located at:

[web/src/locales/](https://github.com/mmtaee/ocserv-dashboard/tree/master/web/src/locales)

Each language has its own JSON file (e.g., `en.json`, `zh.json`, `ru.json`, etc.).

### 🛠️ How to Contribute

1. Go to the [locales](https://github.com/mmtaee/ocserv-dashboard/tree/master/web/src/locales) directory.
2. Choose an existing language file to improve, or create a new `<lang>.json` file for a new language.
3. Add all required translation keys with proper JSON structure.
4. Make sure the JSON syntax is valid.

### 🔧 Update the Installer (Required for New Languages)

After adding a new `<lang>.json` file, you **must update the `install.sh` file**:

Open 👉 [install.sh](https://github.com/mmtaee/ocserv-dashboard/blob/master/install.sh)

Find the line that defines supported languages, and add your new language in the same format, comma-separated.

Example (adding Spanish):

**LANGUAGES=en:English,zh:中文,ru:Русский,fa:فارسی,ar:العربية,es:Español**

Contributing translations and updating the installer helps ensure the dashboard supports users around the world.

---

## 📦 License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

---
## 📈 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=mmtaee/ocserv-dashboard&type=Date)](https://www.star-history.com/#mmtaee/ocserv-dashboard&Date)
