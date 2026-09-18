# Migrate from the legacy dashboard

The [legacy](https://github.com/mmtaee/ocserv-dashboard/tree/legacy) dashboard stored
VPN accounts in SQLite. The current dashboard uses PostgreSQL and restores users
and groups from JSON backups.

There is no built-in SQLite importer. These scripts convert a legacy
`db.sqlite3` (and optional `ocpasswd`) into backup files that the current
**System → Restore** page can import.

Python 3.10+ is required. There are no third-party dependencies.

## What gets converted

| Legacy (SQLite) | Current backup JSON |
| --- | --- |
| `app_ocservgroup` (except `defaults`) | `groups[]` |
| `app_adminpanelconfiguration.default_configs` | `default_group` |
| `app_ocservuser` | users array |
| `active = false` | `is_locked = true` |
| `traffic = 1/2/3` | `Free` / `MonthlyTransmit` / `TotallyReceive` |
| `default_traffic` in GB | `traffic_size` in bytes |
| `tx` / `rx` in GiB | `tx` / `rx` in bytes |
| `dns1`/`dns2`, `routes`, `no_routes` | `dns`, `route`, `no-route` |

Plaintext passwords are taken from SQLite. If a password is missing, the
converter can generate one and write `generated_passwords.tsv`.

Admin panel accounts (`auth_user`) are not migrated. Create a new admin in the
current dashboard setup wizard.

## Convert SQLite to restore JSON

```bash
python3 scripts/legacy-migrate/convert_legacy.py \
  --sqlite /path/to/db.sqlite3 \
  --ocpasswd /path/to/ocpasswd \
  --out-dir ./legacy-backup \
  --gzip
```

This writes:

- `ocserv_groups_backup.json` — default group settings and custom groups
- `ocserv_users_backup.json` — users, passwords, quotas, lock state
- `.json.gz` copies of the same files
- `locked_usernames.txt` — accounts that were inactive in the legacy panel
- `conversion_report.json` — counts and warnings

The users JSON contains plaintext VPN passwords. Keep it mode `600` and delete
it after restore.

## Import in the dashboard

1. Install and set up the current dashboard.
2. Open **System → Restore**.
3. Upload `ocserv_groups_backup.json` first (the filename must contain `groups`).
4. Upload `ocserv_users_backup.json` (the filename must contain `users`).
5. Lock the usernames listed in `locked_usernames.txt`. Restore creates
   `ocpasswd` entries but does not run `ocpasswd -l`.

### Restore races on ocpasswd

Bulk restore starts several `ocpasswd` processes at once. `ocpasswd` uses a
single `/etc/ocserv/ocpasswd.tmp` lock file and exits `1` when that file
already exists. Two processes can also overwrite each other, so users may
appear in PostgreSQL while missing from `ocpasswd`.

Workarounds:

1. Re-upload the users JSON until the dashboard reports no `exit status 1`
   errors. Users already in the database are skipped.
2. Import one account at a time:

```bash
python3 scripts/legacy-migrate/sequential_restore.py \
  --base-url https://YOUR-HOST/api \
  --token 'PANEL_BEARER_TOKEN' \
  --users-json ./legacy-backup/ocserv_users_backup.json \
  --lock \
  --insecure
```

`--lock` calls `POST /ocserv/users/{uid}/lock` for users with `is_locked`.
`--insecure` skips TLS verification for self-signed certificates.

Filter a retry file from a restore error dump:

```bash
python3 scripts/legacy-migrate/convert_legacy.py \
  --users-json ./legacy-backup/ocserv_users_backup.json \
  --from-errors restore-errors.txt \
  --out-dir ./legacy-backup/retry \
  --batch-size 1
```

## Repair a partial ocpasswd file

If PostgreSQL has every user but `/etc/ocserv/ocpasswd` does not, do **not**
re-import the JSON. Already-imported users are skipped and `ocpasswd` is not
updated.

Replace the file with the original hashed `ocpasswd` from the legacy server
(passwords stay the same, including locked `!` hashes):

```bash
sudo cp /etc/ocserv/ocpasswd /etc/ocserv/ocpasswd.bak
sudo cp /path/to/legacy/ocpasswd /etc/ocserv/ocpasswd
sudo chmod 600 /etc/ocserv/ocpasswd
```

If the new server already has extra accounts, merge them:

```bash
python3 scripts/legacy-migrate/fill_ocpasswd.py \
  --legacy-ocpasswd /path/to/legacy/ocpasswd \
  --current-ocpasswd /etc/ocserv/ocpasswd \
  --out ocpasswd.merged
```

Or create only the missing accounts, one `ocpasswd` call at a time (run on the
VPN host):

```bash
sudo python3 scripts/legacy-migrate/fill_ocpasswd.py \
  --apply \
  --users-json ./legacy-backup/ocserv_users_backup.json \
  --current-ocpasswd /etc/ocserv/ocpasswd
```

## Tests

```bash
python3 -m unittest discover -s scripts/legacy-migrate/tests -v
```
