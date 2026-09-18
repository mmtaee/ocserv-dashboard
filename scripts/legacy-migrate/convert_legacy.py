#!/usr/bin/env python3
"""Convert legacy ocserv-dashboard (SQLite + ocpasswd) to new dashboard backup JSON.

The new panel (https://github.com/mmtaee/ocserv-dashboard) restores data from
JSON/JSON.GZ uploads:

  POST /backup/ocserv_groups  ->  {"default_group": {...}, "groups": [...]}
  POST /backup/ocserv_users   ->  [user, user, ...]

Restore recreates VPN accounts with ``ocpasswd`` using the plaintext password
from JSON, so this converter prefers passwords stored in the legacy SQLite DB.
"""

from __future__ import annotations

import argparse
import gzip
import json
import os
import re
import secrets
import sqlite3
import string
import sys
import unicodedata
from dataclasses import dataclass, field
from datetime import datetime, timezone
from decimal import Decimal, ROUND_HALF_UP
from pathlib import Path
from typing import Any, Iterable, Mapping

GIB = Decimal(1024) ** 3
DEFAULT_GROUP = "defaults"
TRAFFIC_FREE = 1
TRAFFIC_MONTHLY = 2
TRAFFIC_TOTALLY = 3

TRAFFIC_TYPES = {
    "Free",
    "MonthlyTransmit",
    "MonthlyReceive",
    "MonthlyRxTx",
    "TotallyTransmit",
    "TotallyReceive",
    "TotallyRxTx",
}

USER_JSON_KEYS = {
    "uid",
    "owner",
    "group",
    "username",
    "password",
    "is_locked",
    "created_at",
    "updated_at",
    "expire_at",
    "deactivated_at",
    "traffic_type",
    "traffic_size",
    "rx",
    "tx",
    "description",
    "is_online",
    "online_sessions",
    "config",
    "certificate_enabled",
    "certificate_available",
    "certificate",
}

GROUP_JSON_KEYS = {"id", "name", "owner", "config"}

GROUP_CONFIG_KEYS = {
    "dns",
    "nbns",
    "ipv4-network",
    "rx-data-per-sec",
    "tx-data-per-sec",
    "explicit-ipv4",
    "cgroup",
    "iroute",
    "route",
    "no-route",
    "net-priority",
    "deny-roaming",
    "no-udp",
    "keepalive",
    "dpd",
    "mobile-dpd",
    "max-same-clients",
    "tunnel-all-dns",
    "stats-report-time",
    "mtu",
    "idle-timeout",
    "mobile-idle-timeout",
    "restrict-user-to-routes",
    "restrict-user-to-ports",
    "split-dns",
    "session-timeout",
}

USER_CONFIG_KEYS = {
    "explicit-ipv4",
    "ipv4-network",
    "dns",
    "nbns",
    "route",
    "no-route",
    "iroute",
    "split-dns",
    "session-timeout",
    "idle-timeout",
    "mobile-idle-timeout",
    "rekey-time",
    "restrict-to-routes",
    "restrict-to-ports",
}

CERTIFICATE_KEYS = {"status", "key_pem", "cert_pem", "p12_base64"}

INT_CONFIG_KEYS = {
    "rx-data-per-sec",
    "tx-data-per-sec",
    "net-priority",
    "keepalive",
    "dpd",
    "mobile-dpd",
    "max-same-clients",
    "stats-report-time",
    "mtu",
    "idle-timeout",
    "mobile-idle-timeout",
    "session-timeout",
    "rekey-time",
}

BOOL_CONFIG_KEYS = {
    "deny-roaming",
    "no-udp",
    "tunnel-all-dns",
    "restrict-user-to-routes",
    "restrict-to-routes",
}

LIST_CONFIG_KEYS = {"dns", "route", "no-route", "split-dns"}

BIDI_RE = re.compile(r"[\u200e\u200f\u202a-\u202e\u2066-\u2069]")
HASH_PASSWORD_RE = re.compile(r"^!?\$[156]\$")
FAILED_USER_RE = re.compile(r"user\s+(.+?):\s+exit status\s+\d+", re.IGNORECASE)
CROCKFORD = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"


@dataclass
class OcpasswdEntry:
    username: str
    group: str
    locked: bool
    raw: str


@dataclass
class Warning:
    code: str
    message: str
    username: str | None = None


@dataclass
class ConversionResult:
    groups_backup: dict[str, Any]
    users_backup: list[dict[str, Any]]
    locked_usernames: list[str]
    generated_passwords: dict[str, str]
    warnings: list[Warning] = field(default_factory=list)
    stats: dict[str, Any] = field(default_factory=dict)


def generate_ulid(now: datetime | None = None, entropy: bytes | None = None) -> str:
    """Generate a 26-character Crockford-base32 ULID."""
    if now is None:
        now = datetime.now(timezone.utc)
    ts = int(now.timestamp() * 1000)
    if ts < 0 or ts > 0xFFFFFFFFFFFF:
        raise ValueError("timestamp out of ULID range")
    if entropy is None:
        entropy = os.urandom(10)
    if len(entropy) != 10:
        raise ValueError("ULID entropy must be 10 bytes")
    value = (ts << 80) | int.from_bytes(entropy, "big")
    chars = ["0"] * 26
    for i in range(25, -1, -1):
        chars[i] = CROCKFORD[value & 31]
        value >>= 5
    return "".join(chars)


def parse_ocpasswd(text: str) -> dict[str, OcpasswdEntry]:
    users: dict[str, OcpasswdEntry] = {}
    for raw in text.splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        parts = line.split(":")
        if len(parts) < 3:
            continue
        username, group, hashed = parts[0], parts[1], ":".join(parts[2:])
        locked = hashed.startswith("!")
        if group in ("", "*"):
            group = DEFAULT_GROUP
        users[username] = OcpasswdEntry(
            username=username,
            group=group,
            locked=locked,
            raw=line,
        )
    return users


def load_json_object(value: Any) -> dict[str, Any]:
    if value in (None, "", "null"):
        return {}
    if isinstance(value, dict):
        return value
    if isinstance(value, (bytes, bytearray)):
        value = value.decode("utf-8")
    if isinstance(value, str):
        parsed = json.loads(value)
        if parsed is None:
            return {}
        if not isinstance(parsed, dict):
            raise ValueError(f"expected JSON object, got {type(parsed).__name__}")
        return parsed
    raise ValueError(f"unsupported JSON value type: {type(value).__name__}")


def as_list(value: Any) -> list[str]:
    if value is None or value == "":
        return []
    if isinstance(value, (list, tuple)):
        return [str(item).strip() for item in value if str(item).strip()]
    if isinstance(value, str):
        text = value.strip()
        if text.startswith("["):
            parsed = json.loads(text)
            return as_list(parsed)
        return [part.strip() for part in re.split(r"[\s,]+", text) if part.strip()]
    return [str(value)]


def as_bool(value: Any) -> bool | None:
    if value is None or value == "":
        return None
    if isinstance(value, bool):
        return value
    text = str(value).strip().lower()
    if text in {"1", "true", "yes", "on"}:
        return True
    if text in {"0", "false", "no", "off"}:
        return False
    raise ValueError(f"invalid boolean value: {value!r}")


def as_int(value: Any) -> int | None:
    if value is None or value == "":
        return None
    if isinstance(value, bool):
        raise ValueError("boolean is not a valid integer config value")
    return int(str(value).strip())


def convert_group_config(raw: Mapping[str, Any] | None) -> dict[str, Any]:
    """Map legacy group/admin configs to OcservGroupConfig JSON."""
    if not raw:
        return {}

    config: dict[str, Any] = {}
    unknown: list[str] = []

    dns: list[str] = []
    for key in ("dns", "dns1", "dns2", "dns3", "dns4", "dns5", "dns6"):
        if key not in raw or raw[key] in (None, ""):
            continue
        dns.extend(as_list(raw[key]))
    if dns:
        config["dns"] = list(dict.fromkeys(dns))

    if "routes" in raw:
        routes = as_list(raw.get("routes"))
        if routes:
            config["route"] = routes
    if "no_routes" in raw:
        no_routes = as_list(raw.get("no_routes"))
        if no_routes:
            config["no-route"] = no_routes

    mapped_aliases = {
        "dns",
        "dns1",
        "dns2",
        "dns3",
        "dns4",
        "dns5",
        "dns6",
        "routes",
        "no_routes",
    }

    for key, value in raw.items():
        if key in mapped_aliases or value in (None, "", [], {}):
            continue
        if key not in GROUP_CONFIG_KEYS:
            unknown.append(key)
            continue
        if key in LIST_CONFIG_KEYS:
            items = as_list(value)
            if items:
                config[key] = items
            continue
        if key in BOOL_CONFIG_KEYS:
            parsed_bool = as_bool(value)
            if parsed_bool is not None:
                config[key] = parsed_bool
            continue
        if key in INT_CONFIG_KEYS:
            parsed_int = as_int(value)
            if parsed_int is not None:
                config[key] = parsed_int
            continue
        config[key] = str(value).strip()

    if unknown:
        config["_unknown"] = unknown
    return config


def split_unknown(config: dict[str, Any]) -> tuple[dict[str, Any], list[str]]:
    unknown = list(config.pop("_unknown", []))
    return config, unknown


def gib_to_bytes(value: Any) -> int:
    if value in (None, ""):
        return 0
    quantized = (Decimal(str(value)) * GIB).quantize(Decimal("1"), rounding=ROUND_HALF_UP)
    return int(quantized)


def gb_to_bytes(value: Any) -> int:
    if value in (None, ""):
        return 0
    return int(Decimal(str(value))) * int(GIB)


def parse_date(value: Any) -> str | None:
    if value in (None, ""):
        return None
    text = str(value).strip()
    if " " in text:
        text = text.split(" ", 1)[0]
    if "T" in text:
        text = text.split("T", 1)[0]
    datetime.strptime(text, "%Y-%m-%d")
    return f"{text}T00:00:00Z"


def looks_like_hash(password: str | None) -> bool:
    if not password:
        return False
    return bool(HASH_PASSWORD_RE.match(password.strip()))


def is_usable_password(password: str | None) -> bool:
    if not password or not password.strip():
        return False
    if looks_like_hash(password):
        return False
    if password.strip().lower() in {"hashed by ocserv", "ocserv password"}:
        return False
    return True


def map_traffic_type(
    traffic: int,
    default_traffic: Any,
    totally_as: str,
) -> tuple[str, int]:
    if traffic == TRAFFIC_FREE:
        return "Free", 0
    size = gb_to_bytes(default_traffic)
    if traffic == TRAFFIC_MONTHLY:
        return "MonthlyTransmit", size
    if traffic == TRAFFIC_TOTALLY:
        if totally_as not in TRAFFIC_TYPES:
            raise ValueError(f"invalid --totally-as value: {totally_as}")
        return totally_as, size
    raise ValueError(f"unknown legacy traffic mode: {traffic}")


def generate_password(length: int = 16) -> str:
    alphabet = string.ascii_letters + string.digits
    return "".join(secrets.choice(alphabet) for _ in range(length))


def unusual_username_reasons(username: str) -> list[str]:
    reasons: list[str] = []
    if BIDI_RE.search(username):
        reasons.append("contains bidirectional Unicode marks")
    if any(unicodedata.category(ch).startswith("C") for ch in username):
        reasons.append("contains control characters")
    if len(username) > 32:
        reasons.append("longer than 32 characters")
    if not re.fullmatch(r"[A-Za-z0-9._-]+", username):
        reasons.append("contains characters outside [A-Za-z0-9._-]")
    return reasons


def sqlite_rows(conn: sqlite3.Connection, sql: str) -> list[sqlite3.Row]:
    conn.row_factory = sqlite3.Row
    return list(conn.execute(sql))


def convert_legacy(
    sqlite_path: str | os.PathLike[str],
    ocpasswd_path: str | os.PathLike[str] | None = None,
    *,
    owner: str = "",
    include_ocpasswd_only: bool = True,
    generate_missing_passwords: bool = True,
    totally_as: str = "TotallyReceive",
    strip_bidi: bool = False,
) -> ConversionResult:
    warnings: list[Warning] = []
    sqlite_path = Path(sqlite_path)
    if not sqlite_path.is_file():
        raise FileNotFoundError(f"SQLite database not found: {sqlite_path}")

    conn = sqlite3.connect(f"file:{sqlite_path}?mode=ro", uri=True)
    try:
        groups_rows = sqlite_rows(conn, "SELECT id, name, desc, configs FROM app_ocservgroup")
        admin_rows = sqlite_rows(
            conn,
            "SELECT default_traffic, default_configs FROM app_adminpanelconfiguration",
        )
        user_rows = sqlite_rows(
            conn,
            """
            SELECT
                u.username,
                u.password,
                u.active,
                u."create" AS created,
                u.expire_date,
                u.deactivate_date,
                u.desc,
                u.traffic,
                u.default_traffic,
                u.tx,
                u.rx,
                g.name AS group_name
            FROM app_ocservuser AS u
            JOIN app_ocservgroup AS g ON g.id = u.group_id
            ORDER BY u.id
            """,
        )
    finally:
        conn.close()

    ocpasswd_users: dict[str, OcpasswdEntry] = {}
    if ocpasswd_path:
        ocpasswd_file = Path(ocpasswd_path)
        if not ocpasswd_file.is_file():
            raise FileNotFoundError(f"ocpasswd file not found: {ocpasswd_file}")
        ocpasswd_users = parse_ocpasswd(ocpasswd_file.read_text(encoding="utf-8"))

    default_group_config: dict[str, Any] = {}
    if admin_rows:
        admin = admin_rows[0]
        raw_default = load_json_object(admin["default_configs"])
        default_group_config, unknown = split_unknown(convert_group_config(raw_default))
        for key in unknown:
            warnings.append(
                Warning("unknown_default_config_key", f"skipped unknown default config key {key!r}")
            )

    groups_backup_items: list[dict[str, Any]] = []
    group_names: set[str] = set()
    for row in groups_rows:
        name = (row["name"] or "").replace(" ", "_")
        group_names.add(name)
        if name == DEFAULT_GROUP:
            continue
        raw_config = load_json_object(row["configs"])
        config, unknown = split_unknown(convert_group_config(raw_config))
        for key in unknown:
            warnings.append(
                Warning(
                    "unknown_group_config_key",
                    f"group {name}: skipped unknown config key {key!r}",
                )
            )
        if len(name) > 16:
            warnings.append(
                Warning(
                    "long_group_name",
                    f"group {name!r} is longer than 16 characters; the new users.group column started as VARCHAR(16)",
                )
            )
        item: dict[str, Any] = {
            "name": name,
            "owner": owner,
            "config": config or None,
        }
        groups_backup_items.append(item)

    users_backup: list[dict[str, Any]] = []
    locked_usernames: list[str] = []
    generated_passwords: dict[str, str] = {}
    seen: set[str] = set()

    def append_user(
        *,
        username: str,
        password: str | None,
        active: bool,
        created: Any,
        expire_date: Any,
        deactivate_date: Any,
        desc: Any,
        traffic: int,
        default_traffic: Any,
        tx: Any,
        rx: Any,
        group_name: str,
        source: str,
    ) -> None:
        original_username = username
        if strip_bidi:
            username = BIDI_RE.sub("", username)
        if username in seen:
            warnings.append(
                Warning("duplicate_username", f"skipped duplicate username {username!r}", username)
            )
            return

        for reason in unusual_username_reasons(username):
            warnings.append(Warning("unusual_username", f"{username!r}: {reason}", username))
        if original_username != username:
            warnings.append(
                Warning(
                    "username_sanitized",
                    f"stripped bidi marks from {original_username!r} -> {username!r}",
                    username,
                )
            )

        group = (group_name or DEFAULT_GROUP).replace(" ", "_") or DEFAULT_GROUP
        oc_entry = ocpasswd_users.get(original_username) or ocpasswd_users.get(username)
        if oc_entry and oc_entry.group != group:
            warnings.append(
                Warning(
                    "group_mismatch",
                    f"{username}: sqlite group {group!r} != ocpasswd group {oc_entry.group!r}; using sqlite",
                    username,
                )
            )

        locked = not bool(active)
        if oc_entry and oc_entry.locked != locked:
            warnings.append(
                Warning(
                    "lock_mismatch",
                    f"{username}: sqlite active={bool(active)} ocpasswd locked={oc_entry.locked}; using sqlite",
                    username,
                )
            )
            locked = locked or oc_entry.locked

        usable = password if is_usable_password(password) else None
        if usable is None:
            if generate_missing_passwords:
                usable = generate_password()
                generated_passwords[username] = usable
                warnings.append(
                    Warning(
                        "generated_password",
                        f"{username}: no plaintext password in SQLite; generated a new one",
                        username,
                    )
                )
            else:
                warnings.append(
                    Warning(
                        "skipped_no_password",
                        f"{username}: skipped because plaintext password is missing",
                        username,
                    )
                )
                return

        traffic_type, traffic_size = map_traffic_type(int(traffic), default_traffic, totally_as)
        created_at = parse_date(created) or datetime.now(timezone.utc).strftime("%Y-%m-%dT00:00:00Z")
        expire_at = parse_date(expire_date)
        deactivated_at = parse_date(deactivate_date)

        if group != DEFAULT_GROUP and group not in group_names:
            warnings.append(
                Warning(
                    "missing_group",
                    f"{username}: group {group!r} is not present in sqlite groups",
                    username,
                )
            )

        user = {
            "uid": generate_ulid(),
            "owner": owner,
            "group": group,
            "username": username,
            "password": usable,
            "is_locked": locked,
            "created_at": created_at,
            "updated_at": created_at,
            "expire_at": expire_at,
            "deactivated_at": deactivated_at,
            "traffic_type": traffic_type,
            "traffic_size": traffic_size,
            "rx": gib_to_bytes(rx),
            "tx": gib_to_bytes(tx),
            "description": desc or "",
            "is_online": False,
            "online_sessions": [],
            "config": None,
            "certificate_enabled": False,
            "certificate_available": False,
        }
        validate_user_record(user)
        users_backup.append(user)
        seen.add(username)
        if locked:
            locked_usernames.append(username)
        if source == "ocpasswd-only":
            warnings.append(
                Warning(
                    "ocpasswd_only",
                    f"{username}: present in ocpasswd but not in SQLite",
                    username,
                )
            )

    sqlite_usernames: set[str] = set()
    for row in user_rows:
        sqlite_usernames.add(row["username"])
        append_user(
            username=row["username"],
            password=row["password"],
            active=bool(row["active"]),
            created=row["created"],
            expire_date=row["expire_date"],
            deactivate_date=row["deactivate_date"],
            desc=row["desc"],
            traffic=int(row["traffic"]),
            default_traffic=row["default_traffic"],
            tx=row["tx"],
            rx=row["rx"],
            group_name=row["group_name"],
            source="sqlite",
        )

    ocpasswd_only = [name for name in ocpasswd_users if name not in sqlite_usernames]
    sqlite_only = [name for name in sqlite_usernames if name not in ocpasswd_users]
    for name in sqlite_only:
        warnings.append(
            Warning("sqlite_only", f"{name}: present in SQLite but not in ocpasswd", name)
        )

    if include_ocpasswd_only:
        for name in sorted(ocpasswd_only):
            entry = ocpasswd_users[name]
            append_user(
                username=entry.username,
                password=None,
                active=not entry.locked,
                created=None,
                expire_date=None,
                deactivate_date=None,
                desc="Imported from ocpasswd; not present in legacy SQLite",
                traffic=TRAFFIC_FREE,
                default_traffic=0,
                tx=0,
                rx=0,
                group_name=entry.group,
                source="ocpasswd-only",
            )
    else:
        for name in ocpasswd_only:
            warnings.append(
                Warning("skipped_ocpasswd_only", f"{name}: ocpasswd-only user skipped", name)
            )

    groups_backup = {
        "default_group": default_group_config,
        "groups": groups_backup_items,
    }
    validate_groups_backup(groups_backup)
    for user in users_backup:
        validate_user_record(user)

    stats = {
        "sqlite_users": len(user_rows),
        "sqlite_groups": len(groups_rows),
        "ocpasswd_users": len(ocpasswd_users),
        "backup_users": len(users_backup),
        "backup_groups": len(groups_backup_items),
        "locked_users": len(locked_usernames),
        "generated_passwords": len(generated_passwords),
        "sqlite_only": sqlite_only,
        "ocpasswd_only": ocpasswd_only,
        "traffic_types": count_by(users_backup, "traffic_type"),
    }
    return ConversionResult(
        groups_backup=groups_backup,
        users_backup=users_backup,
        locked_usernames=locked_usernames,
        generated_passwords=generated_passwords,
        warnings=warnings,
        stats=stats,
    )


def count_by(rows: Iterable[Mapping[str, Any]], key: str) -> dict[str, int]:
    counts: dict[str, int] = {}
    for row in rows:
        value = str(row.get(key))
        counts[value] = counts.get(value, 0) + 1
    return counts


def assert_known_keys(payload: Mapping[str, Any], allowed: set[str], label: str) -> None:
    extra = set(payload) - allowed
    if extra:
        raise ValueError(f"{label} has unknown JSON keys: {sorted(extra)}")


def validate_group_config(config: Any, label: str) -> None:
    if config is None:
        return
    if not isinstance(config, dict):
        raise ValueError(f"{label} config must be an object")
    assert_known_keys(config, GROUP_CONFIG_KEYS, label)
    for key, value in config.items():
        if key in LIST_CONFIG_KEYS and not isinstance(value, list):
            raise ValueError(f"{label}.{key} must be a JSON array")
        if key in BOOL_CONFIG_KEYS and not isinstance(value, bool):
            raise ValueError(f"{label}.{key} must be a boolean")
        if key in INT_CONFIG_KEYS and not isinstance(value, int):
            raise ValueError(f"{label}.{key} must be an integer")


def validate_user_config(config: Any, label: str) -> None:
    if config is None:
        return
    if not isinstance(config, dict):
        raise ValueError(f"{label} config must be an object")
    assert_known_keys(config, USER_CONFIG_KEYS, label)


def validate_groups_backup(payload: Mapping[str, Any]) -> None:
    extra = set(payload) - {"default_group", "groups"}
    if extra:
        raise ValueError(f"groups backup has unknown JSON keys: {sorted(extra)}")
    if "default_group" not in payload:
        raise ValueError("groups backup is missing default_group")
    validate_group_config(payload.get("default_group") or {}, "default_group")
    groups = payload.get("groups")
    if groups is None:
        return
    if not isinstance(groups, list):
        raise ValueError("groups must be a JSON array")
    for index, group in enumerate(groups):
        if not isinstance(group, dict):
            raise ValueError(f"groups[{index}] must be an object")
        assert_known_keys(group, GROUP_JSON_KEYS, f"groups[{index}]")
        if not group.get("name"):
            raise ValueError(f"groups[{index}] is missing name")
        validate_group_config(group.get("config"), f"groups[{index}]")


def validate_user_record(user: Mapping[str, Any]) -> None:
    assert_known_keys(user, USER_JSON_KEYS, f"user {user.get('username')!r}")
    required = ("username", "password", "group", "traffic_type", "traffic_size")
    for key in required:
        if key not in user:
            raise ValueError(f"user is missing {key}")
    if user["traffic_type"] not in TRAFFIC_TYPES:
        raise ValueError(f"{user['username']}: invalid traffic_type {user['traffic_type']!r}")
    if not isinstance(user["is_locked"], bool):
        raise ValueError(f"{user['username']}: is_locked must be boolean")
    for key in ("created_at", "updated_at", "expire_at", "deactivated_at"):
        value = user.get(key)
        if value is None:
            continue
        datetime.strptime(value, "%Y-%m-%dT%H:%M:%SZ")
    validate_user_config(user.get("config"), f"user {user['username']}")
    certificate = user.get("certificate")
    if certificate is not None:
        if not isinstance(certificate, dict):
            raise ValueError(f"{user['username']}: certificate must be an object")
        assert_known_keys(certificate, CERTIFICATE_KEYS, f"user {user['username']} certificate")


def parse_failed_usernames(text: str) -> list[str]:
    """Parse usernames from dashboard restore errors like 'user Maya: exit status 1'."""
    names: list[str] = []
    seen: set[str] = set()
    for match in FAILED_USER_RE.finditer(text):
        username = match.group(1).strip()
        if username and username not in seen:
            seen.add(username)
            names.append(username)
    return names


def load_username_list(path: str | os.PathLike[str]) -> list[str]:
    text = Path(path).read_text(encoding="utf-8")
    failed = parse_failed_usernames(text)
    if failed:
        return failed
    names: list[str] = []
    seen: set[str] = set()
    for line in text.splitlines():
        username = line.strip()
        if not username or username.startswith("#"):
            continue
        if username not in seen:
            seen.add(username)
            names.append(username)
    return names


def load_users_backup(path: str | os.PathLike[str]) -> list[dict[str, Any]]:
    raw_path = Path(path)
    if raw_path.suffixes[-2:] == [".json", ".gz"] or raw_path.suffix == ".gz":
        payload = json.loads(gzip.decompress(raw_path.read_bytes()).decode("utf-8"))
    else:
        payload = json.loads(raw_path.read_text(encoding="utf-8"))
    if not isinstance(payload, list):
        raise ValueError(f"{raw_path} is not a users backup array")
    users: list[dict[str, Any]] = []
    for index, user in enumerate(payload):
        if not isinstance(user, dict):
            raise ValueError(f"users[{index}] must be an object")
        validate_user_record(user)
        users.append(user)
    return users


def filter_users(
    users: list[dict[str, Any]],
    usernames: Iterable[str],
) -> tuple[list[dict[str, Any]], list[str]]:
    wanted = list(dict.fromkeys(usernames))
    by_name = {str(user["username"]): user for user in users}
    filtered = []
    missing = []
    for name in wanted:
        user = by_name.get(name)
        if user is None:
            missing.append(name)
            continue
        filtered.append(user)
    return filtered, missing


def dumps(payload: Any, compact: bool) -> str:
    if compact:
        return json.dumps(payload, ensure_ascii=False, separators=(",", ":"))
    return json.dumps(payload, ensure_ascii=False, indent=2) + "\n"


def write_json(
    path: Path,
    payload: Any,
    compact: bool,
    gzip_output: bool,
    *,
    private: bool = False,
) -> Path:
    path.parent.mkdir(parents=True, exist_ok=True)
    data = dumps(payload, compact).encode("utf-8")
    if gzip_output:
        gz_path = path.with_suffix(path.suffix + ".gz") if path.suffix == ".json" else path
        with gzip.open(gz_path, "wb") as handle:
            handle.write(data)
        if private:
            os.chmod(gz_path, 0o600)
        return gz_path
    path.write_bytes(data)
    if private:
        os.chmod(path, 0o600)
    return path


def write_user_files(
    users: list[dict[str, Any]],
    out_dir: Path,
    *,
    compact: bool,
    gzip_output: bool,
    also_plain: bool,
    batch_size: int = 0,
    prefix: str = "ocserv_users_backup",
) -> dict[str, str]:
    written: dict[str, str] = {}
    users_json = out_dir / f"{prefix}.json"
    if gzip_output:
        written["users_gz"] = str(write_json(users_json, users, compact, True, private=True))
        if also_plain:
            written["users_json"] = str(write_json(users_json, users, compact, False, private=True))
    else:
        written["users_json"] = str(write_json(users_json, users, compact, False, private=True))

    if batch_size and batch_size > 0:
        batch_dir = out_dir / "batches"
        batch_dir.mkdir(parents=True, exist_ok=True)
        total = (len(users) + batch_size - 1) // batch_size if users else 0
        for index, start in enumerate(range(0, len(users), batch_size), start=1):
            chunk = users[start : start + batch_size]
            name = f"ocserv_users_backup_{index:03d}_of_{total:03d}.json"
            path = write_json(batch_dir / name, chunk, compact, gzip_output, private=True)
            written[f"users_batch_{index:03d}"] = str(path)
    return written


def write_outputs(
    result: ConversionResult,
    out_dir: str | os.PathLike[str],
    *,
    compact: bool = False,
    gzip_output: bool = False,
    also_plain: bool = True,
    skip_groups: bool = False,
    batch_size: int = 0,
) -> dict[str, str]:
    out = Path(out_dir)
    out.mkdir(parents=True, exist_ok=True)
    written: dict[str, str] = {}

    if not skip_groups:
        groups_json = out / "ocserv_groups_backup.json"
        if gzip_output:
            written["groups_gz"] = str(write_json(groups_json, result.groups_backup, compact, True))
            if also_plain:
                written["groups_json"] = str(
                    write_json(groups_json, result.groups_backup, compact, False)
                )
        else:
            written["groups_json"] = str(
                write_json(groups_json, result.groups_backup, compact, False)
            )

    written.update(
        write_user_files(
            result.users_backup,
            out,
            compact=compact,
            gzip_output=gzip_output,
            also_plain=also_plain,
            batch_size=batch_size,
        )
    )

    locked_path = out / "locked_usernames.txt"
    locked_path.write_text(
        "".join(f"{name}\n" for name in result.locked_usernames),
        encoding="utf-8",
    )
    written["locked_usernames"] = str(locked_path)

    if result.generated_passwords:
        generated_path = out / "generated_passwords.tsv"
        lines = ["username\tpassword\n"]
        for username, password in sorted(result.generated_passwords.items()):
            lines.append(f"{username}\t{password}\n")
        generated_path.write_text("".join(lines), encoding="utf-8")
        os.chmod(generated_path, 0o600)
        written["generated_passwords"] = str(generated_path)

    report = {
        "stats": result.stats,
        "warnings": [
            {"code": item.code, "message": item.message, "username": item.username}
            for item in result.warnings
        ],
        "locked_usernames": result.locked_usernames,
        "notes": [
            "Import groups first, then users, in the new dashboard Restore page.",
            "Filenames must contain 'groups' or 'users' so the UI picks the right restore type.",
            "Restore recreates ocpasswd hashes from JSON plaintext passwords.",
            "Restore does not call ocpasswd -l, so lock users from locked_usernames.txt after import.",
            "The new panel restore runs up to 10 ocpasswd processes in parallel; ocpasswd exits 1 if ocpasswd.tmp already exists. Re-upload the users file until no errors remain, or use --batch-size 1.",
        ],
    }
    report_path = out / "conversion_report.json"
    report_path.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    written["report"] = str(report_path)
    return written


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Convert legacy ocserv-dashboard SQLite + ocpasswd to new dashboard backup JSON.",
        epilog=(
            "After conversion, in the new panel: System -> Restore. "
            "Upload ocserv_groups_backup.json first, then ocserv_users_backup.json. "
            "Then lock the users listed in locked_usernames.txt."
        ),
    )
    parser.add_argument("--sqlite", default="db.sqlite3", help="Path to legacy db.sqlite3")
    parser.add_argument("--ocpasswd", default="ocpasswd", help="Path to legacy ocpasswd file")
    parser.add_argument("--out-dir", default="backup", help="Output directory")
    parser.add_argument("--owner", default="", help="Optional owner stored on imported records")
    parser.add_argument(
        "--totally-as",
        default="TotallyReceive",
        choices=sorted(TRAFFIC_TYPES),
        help="How to map legacy traffic=3 (totally). Old panel compared RX, so TotallyReceive is the default.",
    )
    parser.add_argument("--gzip", action="store_true", help="Also write .json.gz files")
    parser.add_argument("--compact", action="store_true", help="Write minified JSON")
    parser.add_argument(
        "--no-ocpasswd-only",
        action="store_true",
        help="Do not import users that exist only in ocpasswd",
    )
    parser.add_argument(
        "--no-generate-passwords",
        action="store_true",
        help="Skip users that have no plaintext password instead of generating one",
    )
    parser.add_argument(
        "--strip-bidi",
        action="store_true",
        help="Strip bidirectional Unicode marks from usernames",
    )
    parser.add_argument(
        "--users-json",
        help="Existing users backup JSON/JSON.GZ to filter instead of converting SQLite again",
    )
    parser.add_argument(
        "--only-file",
        help="Keep only these usernames. Accepts a list or a restore error dump ('user NAME: exit status 1')",
    )
    parser.add_argument(
        "--from-errors",
        help="Alias for --only-file when the file contains dashboard restore errors",
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=0,
        help="Also split users into files of N accounts. Use 1 to avoid ocpasswd.lock races in restore.",
    )
    parser.add_argument(
        "--skip-groups",
        action="store_true",
        help="Do not write the groups backup (useful for retrying failed users)",
    )
    return parser


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    only_file = args.only_file or args.from_errors
    only_usernames = load_username_list(only_file) if only_file else []

    if args.users_json:
        users = load_users_backup(args.users_json)
        if only_usernames:
            users, missing = filter_users(users, only_usernames)
            if missing:
                raise SystemExit(f"usernames not found in users JSON: {', '.join(missing)}")
        result = ConversionResult(
            groups_backup={"default_group": {}, "groups": []},
            users_backup=users,
            locked_usernames=[user["username"] for user in users if user.get("is_locked")],
            generated_passwords={},
            warnings=[],
            stats={
                "backup_users": len(users),
                "backup_groups": 0,
                "locked_users": sum(1 for user in users if user.get("is_locked")),
                "filtered_from": str(args.users_json),
            },
        )
        skip_groups = True
    else:
        result = convert_legacy(
            args.sqlite,
            args.ocpasswd if args.ocpasswd else None,
            owner=args.owner,
            include_ocpasswd_only=not args.no_ocpasswd_only,
            generate_missing_passwords=not args.no_generate_passwords,
            totally_as=args.totally_as,
            strip_bidi=args.strip_bidi,
        )
        if only_usernames:
            result.users_backup, missing = filter_users(result.users_backup, only_usernames)
            if missing:
                raise SystemExit(f"usernames not found in SQLite/ocpasswd: {', '.join(missing)}")
            result.locked_usernames = [
                user["username"] for user in result.users_backup if user.get("is_locked")
            ]
            result.stats["backup_users"] = len(result.users_backup)
            result.stats["locked_users"] = len(result.locked_usernames)
        skip_groups = args.skip_groups

    written = write_outputs(
        result,
        args.out_dir,
        compact=args.compact,
        gzip_output=args.gzip,
        also_plain=True,
        skip_groups=skip_groups,
        batch_size=args.batch_size,
    )

    print(f"Users in backup: {result.stats['backup_users']}")
    print(f"Custom groups:   {result.stats['backup_groups']}")
    print(f"Locked users:    {result.stats['locked_users']}")
    print(f"Warnings:        {len(result.warnings)}")
    if result.generated_passwords:
        print(f"Generated passwords: {len(result.generated_passwords)} (see generated_passwords.tsv)")
    print("Wrote:")
    for key, path in written.items():
        print(f"  {key}: {path}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
