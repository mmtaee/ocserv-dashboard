# TODO:

- [X] backend log pkg
- [x] sync ocpassw and add OS-user to dashboard users
- [x] rescan groups and add to dashboard groups
- [x] Add an option in the `install.sh` script to support standalone dashboard installation or upgrade
- [ ] Restore users with full traffic data and reset monthly status (reset `RX/TX` counters and active status)
- [ ] search ocserv users by username (#88)
- [ ] Add a refresh button on Staffs, Ocserv Groups, and Ocserv Users pages
- [ ] Manage `systemd` services: restart and check statuses in dashboard
- [x] Refactor large interfaces into smaller, focused, single-responsibility interfaces
- [ ] Support multiple owners per Ocserv user (R&D) (#88)
- [ ] Allow users to disconnect their active sessions from the customer page(#93)
- [ ] Add backup and restore support for ocserv users (export/import as JSON with full details)(#96)
- [ ] Research and implement a new permission strategy for staff roles, introducing super-admin, admin, and staff levels (#97)
- [ ] Implement super-admin, admin, and staff activities tracking and logs (#97)
- [ ] super-admin can add user for admin (#88) updated with (#97)
