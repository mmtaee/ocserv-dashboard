#!/usr/bin/env python3
"""Add VPN users missing from /etc/ocserv/ocpasswd after a raced dashboard restore.

The new panel restore writes ocpasswd from 10 processes at once. The last writer
wins, so accounts can exist in PostgreSQL and still be absent from ocpasswd.
Re-importing the JSON does not fix this: already-imported DB users are skipped
and ocpasswd is not called again.

Typical fix: merge the original hashed ocpasswd into the current file, or run
ocpasswd sequentially for the missing usernames only.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
import time
from pathlib import Path


def parse_ocpasswd(text: str) -> dict[str, str]:
    users: dict[str, str] = {}
    for raw in text.splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        username = line.split(":", 1)[0]
        users[username] = line
    return users


def load_users_json(path: Path) -> list[dict]:
    payload = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(payload, list):
        raise SystemExit(f"{path} is not a users backup array")
    return payload


def write_ocpasswd(path: Path, lines: list[str]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    data = "".join(f"{line}\n" for line in lines)
    path.write_text(data, encoding="utf-8")
    os.chmod(path, 0o600)


def merge_files(legacy_path: Path, current_path: Path | None) -> tuple[list[str], dict[str, int]]:
    legacy = parse_ocpasswd(legacy_path.read_text(encoding="utf-8"))
    current = parse_ocpasswd(current_path.read_text(encoding="utf-8")) if current_path else {}

    names: list[str] = []
    seen: set[str] = set()
    for source in (current, legacy):
        for name in source:
            if name not in seen:
                seen.add(name)
                names.append(name)

    merged = []
    added_from_legacy = 0
    kept_current = 0
    for name in names:
        if name in current:
            merged.append(current[name])
            kept_current += 1
        else:
            merged.append(legacy[name])
            added_from_legacy += 1

    stats = {
        "legacy": len(legacy),
        "current": len(current),
        "merged": len(merged),
        "kept_from_current": kept_current,
        "added_from_legacy": added_from_legacy,
        "only_in_current": len(set(current) - set(legacy)),
    }
    return merged, stats


def missing_from_current(wanted: list[str], current: dict[str, str]) -> list[str]:
    return [name for name in wanted if name not in current]


def run_ocpasswd(
    ocpasswd_bin: str,
    ocpasswd_path: str,
    username: str,
    password: str,
    group: str,
    locked: bool,
) -> None:
    args = [ocpasswd_bin, "-c", ocpasswd_path, username]
    if group and group not in {"", "*", "defaults"}:
        args = [ocpasswd_bin, "-g", group, "-c", ocpasswd_path, username]
    proc = subprocess.run(
        args,
        input=f"{password}\n{password}\n",
        text=True,
        capture_output=True,
        check=False,
    )
    if proc.returncode != 0:
        detail = (proc.stderr or proc.stdout or f"exit {proc.returncode}").strip()
        raise RuntimeError(detail)
    if locked:
        lock = subprocess.run(
            [ocpasswd_bin, "-l", "-c", ocpasswd_path, username],
            capture_output=True,
            text=True,
            check=False,
        )
        if lock.returncode != 0:
            detail = (lock.stderr or lock.stdout or f"exit {lock.returncode}").strip()
            raise RuntimeError(f"created but lock failed: {detail}")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Fill missing ocpasswd entries after a partial dashboard restore."
    )
    parser.add_argument(
        "--legacy-ocpasswd",
        default="ocpasswd",
        help="Original ocpasswd from the old server (complete hashed file)",
    )
    parser.add_argument(
        "--current-ocpasswd",
        help="Current ocpasswd from the new server. For --apply defaults to /etc/ocserv/ocpasswd.",
    )
    parser.add_argument(
        "--users-json",
        default="backup/ocserv_users_backup.json",
        help="Users backup JSON, used by --apply to create missing accounts",
    )
    parser.add_argument("--out", default="ocpasswd.merged", help="Merged ocpasswd output path")
    parser.add_argument(
        "--apply",
        action="store_true",
        help="On the VPN host: sequentially run ocpasswd for users missing from --current-ocpasswd",
    )
    parser.add_argument("--ocpasswd-bin", default="ocpasswd", help="ocpasswd executable")
    parser.add_argument("--sleep", type=float, default=0.2, help="Delay between ocpasswd calls")
    return parser


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    legacy_path = Path(args.legacy_ocpasswd)
    if not legacy_path.is_file():
        raise SystemExit(f"legacy ocpasswd not found: {legacy_path}")

    current_path = Path(args.current_ocpasswd) if args.current_ocpasswd else None
    if args.apply and current_path is None:
        current_path = Path("/etc/ocserv/ocpasswd")
    if current_path and not current_path.is_file():
        raise SystemExit(f"current ocpasswd not found: {current_path}")

    if args.apply:
        users = load_users_json(Path(args.users_json))
        current = parse_ocpasswd(current_path.read_text(encoding="utf-8")) if current_path else {}
        missing = missing_from_current([str(user["username"]) for user in users], current)
        print(f"Current ocpasswd users: {len(current)}")
        print(f"Missing: {len(missing)}")
        failed: list[str] = []
        by_name = {str(user["username"]): user for user in users}
        for index, username in enumerate(missing, start=1):
            user = by_name[username]
            try:
                run_ocpasswd(
                    args.ocpasswd_bin,
                    str(current_path),
                    username,
                    str(user["password"]),
                    str(user.get("group") or "defaults"),
                    bool(user.get("is_locked")),
                )
                print(f"[{index}/{len(missing)}] {username}: added")
            except Exception as exc:  # noqa: BLE001
                print(f"[{index}/{len(missing)}] {username}: FAILED {exc}")
                failed.append(username)
            time.sleep(args.sleep)
        print(f"Added: {len(missing) - len(failed)}")
        print(f"Failed: {len(failed)}")
        return 1 if failed else 0

    merged, stats = merge_files(legacy_path, current_path)
    write_ocpasswd(Path(args.out), merged)
    print(f"Legacy users:          {stats['legacy']}")
    print(f"Current users:         {stats['current']}")
    print(f"Kept from current:     {stats['kept_from_current']}")
    print(f"Added from legacy:     {stats['added_from_legacy']}")
    print(f"Only on new server:    {stats['only_in_current']}")
    print(f"Merged users:          {stats['merged']}")
    print(f"Wrote: {args.out}")
    print("Install on the VPN host:")
    print("  sudo cp /etc/ocserv/ocpasswd /etc/ocserv/ocpasswd.bak")
    print(f"  sudo cp {args.out} /etc/ocserv/ocpasswd")
    print("  sudo chmod 600 /etc/ocserv/ocpasswd")
    print("  sudo chown root:root /etc/ocserv/ocpasswd")
    return 0


if __name__ == "__main__":
    sys.exit(main())
