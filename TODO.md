# TODO

## ✅ Completed

### Backend / Core Improvements
- [x] backend log pkg
- [x] sync ocpassw and add OS-user to dashboard users
- [x] rescan groups and add to dashboard groups
- [x] Refactor large interfaces into smaller, focused, single-responsibility interfaces
- [x] search ocserv users by username (#88)
- [x] Add a refresh button on Ocserv Users page
- [x] Restore users with full traffic data and reset monthly status (reset `RX/TX` counters and active status)
- [x] Allow users to disconnect their active sessions from the customer page (#93)
- [x] Add backup and restore support for ocserv users (export/import JSON with full details) (#96)
- [x] Change ocserv installation method from deb repo to binary installer (R&D) (#111)
- [x] Implement real-time Ocserv user stream processing (change log_stream service)
- [x] total inactive user in dashboard summary (#120)
- [x] automatically delete inactive users after x days (#120)
- [x] Add option in `install.sh` for standalone dashboard installation or upgrade

---

## 🔧 System & Services

- [ ] Manage `systemd` services (restart and check status in dashboard)
- [x] Implement Ocserv binary installation using Meson build system (replace current installation flow with reproducible Meson-based build + integrate into install.sh)

---

## 👥 Users & Permissions (RBAC / Ownership)

- [ ] Research and implement permission strategy (super-admin, admin, staff) (#97)
- [ ] Implement activity tracking and logs for super-admin/admin/staff (#97)
- [ ] Allow super-admin to create admin users (#88, #97)
- [ ] Support multiple owners per Ocserv user (R&D) (#88)
- [ ] Separate password update logic from user profile updates (#121)

---

## 📊 Dashboard & UX

- [ ] Implement bulk operations for Ocserv users (checkbox selection) (#113)
- [ ] Display the current VPN version in the dashboard (#140)
- [ ] Add automatic update checks with GitHub release notifications, including a one-click update option (#140)

---

## 🐳 DevOps / Deployment

- [ ] Publish official pre-built Docker images (#100)
- [ ] Add real-time server resource monitoring (CPU, RAM, disk, overall usage) to the dashboard, similar to x-ui (#139)