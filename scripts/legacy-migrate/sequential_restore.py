#!/usr/bin/env python3
"""Restore ocserv-dashboard users one at a time.

The new panel restore runs up to 10 `ocpasswd` processes in parallel.
ocpasswd exits 1 when `/etc/ocserv/ocpasswd.tmp` already exists, so a bulk
JSON upload drops a random subset of accounts.

This script posts a one-user backup to POST /api/backup/ocserv_users,
waits for ocpasswd to finish, then continues. Already imported users are
skipped by the panel.
"""

from __future__ import annotations

import argparse
import json
import ssl
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any


def load_users(path: Path) -> list[dict[str, Any]]:
    payload = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(payload, list):
        raise SystemExit(f"{path} is not a users backup array")
    return payload


def json_request(
    url: str,
    *,
    method: str = "GET",
    token: str | None = None,
    payload: dict[str, Any] | None = None,
    insecure: bool = False,
    timeout: int = 30,
) -> tuple[int, Any]:
    data = None
    headers = {"Accept": "application/json"}
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"
    if token:
        headers["Authorization"] = f"Bearer {token}"
    request = urllib.request.Request(url, data=data, headers=headers, method=method)
    context = ssl._create_unverified_context() if insecure else None
    try:
        with urllib.request.urlopen(request, timeout=timeout, context=context) as response:
            body = response.read().decode("utf-8")
            return response.status, json.loads(body) if body else {}
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace")
        try:
            parsed = json.loads(body) if body else {}
        except json.JSONDecodeError:
            parsed = {"error": body}
        return exc.code, parsed


def restore_one(
    url: str,
    user: dict[str, Any],
    token: str,
    *,
    insecure: bool,
    timeout: int,
) -> tuple[int, Any]:
    body = json.dumps([user], ensure_ascii=False).encode("utf-8")
    boundary = "----OcservMigrateBoundary"
    filename = f"ocserv_users_{user['username']}_backup.json"
    chunk = b"".join(
        [
            f"--{boundary}\r\n".encode(),
            (
                f'Content-Disposition: form-data; name="file"; filename="{filename}"\r\n'
                "Content-Type: application/json\r\n\r\n"
            ).encode(),
            body,
            b"\r\n",
            f"--{boundary}--\r\n".encode(),
        ]
    )
    headers = {
        "Authorization": f"Bearer {token}",
        "Accept": "application/json",
        "Content-Type": f"multipart/form-data; boundary={boundary}",
    }
    request = urllib.request.Request(url, data=chunk, headers=headers, method="POST")
    context = ssl._create_unverified_context() if insecure else None
    try:
        with urllib.request.urlopen(request, timeout=timeout, context=context) as response:
            raw = response.read().decode("utf-8")
            return response.status, json.loads(raw) if raw else {}
    except urllib.error.HTTPError as exc:
        raw = exc.read().decode("utf-8", errors="replace")
        try:
            parsed = json.loads(raw) if raw else {}
        except json.JSONDecodeError:
            parsed = {"error": raw}
        return exc.code, parsed


def error_text(payload: Any) -> str:
    if isinstance(payload, dict):
        for key in ("error", "message", "Error"):
            value = payload.get(key)
            if value:
                return str(value)
        return json.dumps(payload, ensure_ascii=False)
    return str(payload)


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Import ocserv users sequentially through the new dashboard restore API."
    )
    parser.add_argument(
        "--base-url",
        required=True,
        help="API root, for example https://vpn.example.com/api",
    )
    parser.add_argument(
        "--users-json",
        default="backup/retry/ocserv_users_backup.json",
        help="Users backup JSON (array)",
    )
    parser.add_argument("--token", help="Existing Bearer token from the panel")
    parser.add_argument("--username", help="Admin username if --token is not set")
    parser.add_argument("--password", help="Admin password if --token is not set")
    parser.add_argument("--sleep", type=float, default=0.4, help="Delay between users, seconds")
    parser.add_argument("--retries", type=int, default=3, help="Retries per user on exit status 1")
    parser.add_argument("--insecure", action="store_true", help="Ignore TLS certificate errors")
    parser.add_argument("--lock", action="store_true", help="Lock accounts that have is_locked=true")
    parser.add_argument("--dry-run", action="store_true", help="Print usernames without calling the API")
    return parser


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    base = args.base_url.rstrip("/")
    users = load_users(Path(args.users_json))
    print(f"Users to import: {len(users)}")

    if args.dry_run:
        for user in users:
            flag = " locked" if user.get("is_locked") else ""
            print(f"  {user.get('username')}{flag}")
        return 0

    token = args.token
    if not token:
        if not args.username or not args.password:
            raise SystemExit("provide --token or both --username and --password")
        status, payload = json_request(
            f"{base}/system/users/login",
            method="POST",
            payload={
                "username": args.username,
                "password": args.password,
                "remember_me": True,
            },
            insecure=args.insecure,
        )
        if status >= 400 or not isinstance(payload, dict) or not payload.get("token"):
            raise SystemExit(f"login failed ({status}): {error_text(payload)}")
        token = str(payload["token"])
        print("Logged in")

    restore_url = f"{base}/backup/ocserv_users"
    inserted = 0
    existing = 0
    failed: list[str] = []

    for index, user in enumerate(users, start=1):
        username = str(user.get("username"))
        last_error = "unknown error"
        success = False
        for attempt in range(1, args.retries + 1):
            status, payload = restore_one(
                restore_url,
                user,
                token,
                insecure=args.insecure,
                timeout=30,
            )
            if status < 400:
                existing_names = payload.get("existing") or []
                inserted_names = payload.get("inserted") or []
                if username in existing_names or (
                    not inserted_names and existing_names
                ):
                    print(f"[{index}/{len(users)}] {username}: already exists")
                    existing += 1
                else:
                    print(f"[{index}/{len(users)}] {username}: inserted")
                    inserted += 1
                success = True
                if args.lock and user.get("is_locked"):
                    uid = user.get("uid")
                    if uid:
                        lock_status, lock_payload = json_request(
                            f"{base}/ocserv/users/{uid}/lock",
                            method="POST",
                            token=token,
                            insecure=args.insecure,
                        )
                        if lock_status >= 400:
                            print(
                                f"  lock failed for {username}: {error_text(lock_payload)}"
                            )
                break
            last_error = error_text(payload)
            if "exit status 1" in last_error and attempt < args.retries:
                time.sleep(max(args.sleep, 0.5))
                continue
            break
        if not success:
            print(f"[{index}/{len(users)}] {username}: FAILED {last_error}")
            failed.append(username)
        time.sleep(args.sleep)

    print(f"Inserted: {inserted}")
    print(f"Existing: {existing}")
    print(f"Failed:   {len(failed)}")
    if failed:
        print("Failed usernames:")
        for username in failed:
            print(f"  {username}")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
