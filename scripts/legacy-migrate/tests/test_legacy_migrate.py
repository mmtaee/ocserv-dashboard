import json
import sqlite3
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from convert_legacy import (  # noqa: E402
    convert_group_config,
    convert_legacy,
    dumps,
    filter_users,
    generate_ulid,
    gib_to_bytes,
    map_traffic_type,
    parse_failed_usernames,
    parse_ocpasswd,
    split_unknown,
    validate_groups_backup,
    validate_user_record,
    write_outputs,
)
from fill_ocpasswd import merge_files  # noqa: E402


SCHEMA = """
CREATE TABLE app_ocservgroup (
    id INTEGER PRIMARY KEY,
    name TEXT,
    desc TEXT,
    configs TEXT
);
CREATE TABLE app_adminpanelconfiguration (
    id INTEGER PRIMARY KEY,
    default_traffic INTEGER,
    default_configs TEXT
);
CREATE TABLE app_ocservuser (
    id INTEGER PRIMARY KEY,
    username TEXT,
    password TEXT,
    active INTEGER,
    "create" TEXT,
    expire_date TEXT,
    deactivate_date TEXT,
    desc TEXT,
    traffic INTEGER,
    default_traffic INTEGER,
    tx REAL,
    rx REAL,
    group_id INTEGER
);
"""


def write_legacy_db(path: Path) -> None:
    conn = sqlite3.connect(path)
    conn.executescript(SCHEMA)
    conn.execute(
        "INSERT INTO app_ocservgroup VALUES (1, 'defaults', 'defaults group', '{}')"
    )
    conn.execute(
        "INSERT INTO app_ocservgroup VALUES (2, 'staff', 'staff group', ?)",
        ('{"dns1": "1.1.1.1", "max-same-clients": "2"}',),
    )
    conn.execute(
        "INSERT INTO app_adminpanelconfiguration VALUES (1, 10, ?)",
        (
            json.dumps(
                {
                    "routes": ["10.10.1.0/24"],
                    "dns1": "10.61.11.1",
                    "no-udp": "true",
                    "ipv4-network": "10.61.64.0/20",
                }
            ),
        ),
    )
    conn.execute(
        """
        INSERT INTO app_ocservuser VALUES
        (1, 'alice', 'password12', 1, '2024-12-05', NULL, NULL, 'demo', 1, 0, 0.4672816, 0.1, 1),
        (2, 'bob', 'secret99', 0, '2025-01-12', '2026-12-31', NULL, NULL, 2, 10, 0, 0, 2)
        """
    )
    conn.commit()
    conn.close()


class GroupConfigTests(unittest.TestCase):
    def test_legacy_admin_config_mapping(self):
        raw = {
            "routes": ["10.10.1.0/24", "10.40.0.0/16"],
            "dns1": "10.61.11.1",
            "max-same-clients": "1",
            "dpd": "90",
            "tunnel-all-dns": "true",
            "no-udp": "true",
            "ipv4-network": "10.61.64.0/20",
            "restrict-user-to-routes": "true",
            "mystery": "skip-me",
        }
        config, unknown = split_unknown(convert_group_config(raw))
        self.assertEqual(unknown, ["mystery"])
        self.assertEqual(config["dns"], ["10.61.11.1"])
        self.assertEqual(config["route"], ["10.10.1.0/24", "10.40.0.0/16"])
        self.assertEqual(config["max-same-clients"], 1)
        self.assertTrue(config["no-udp"])


class MappingTests(unittest.TestCase):
    def test_traffic_modes(self):
        self.assertEqual(map_traffic_type(1, 100, "TotallyReceive"), ("Free", 0))
        self.assertEqual(
            map_traffic_type(2, 10, "TotallyReceive"),
            ("MonthlyTransmit", 10 * 1024**3),
        )
        self.assertEqual(
            map_traffic_type(3, 5, "TotallyReceive"),
            ("TotallyReceive", 5 * 1024**3),
        )

    def test_gib_to_bytes_rounding(self):
        self.assertEqual(gib_to_bytes("0.4672816"), 501739798)

    def test_ocpasswd_lock_and_default_group(self):
        parsed = parse_ocpasswd("alice:*:$5$hash\nbob:staff:!$5$hash\n# comment\n")
        self.assertEqual(parsed["alice"].group, "defaults")
        self.assertFalse(parsed["alice"].locked)
        self.assertEqual(parsed["bob"].group, "staff")
        self.assertTrue(parsed["bob"].locked)

    def test_ulid_length(self):
        value = generate_ulid()
        self.assertEqual(len(value), 26)
        self.assertTrue(value.isalnum())
        self.assertEqual(value, value.upper())


class ConversionTests(unittest.TestCase):
    def test_sqlite_and_ocpasswd_mapping(self):
        with tempfile.TemporaryDirectory() as tmp:
            db_path = Path(tmp) / "db.sqlite3"
            ocpasswd_path = Path(tmp) / "ocpasswd"
            write_legacy_db(db_path)
            ocpasswd_path.write_text("alice:*:$5$hash\nbob:staff:!$5$hash\n", encoding="utf-8")

            result = convert_legacy(db_path, ocpasswd_path)
            validate_groups_backup(result.groups_backup)
            self.assertEqual(result.stats["backup_users"], 2)
            self.assertEqual(result.stats["backup_groups"], 1)
            self.assertEqual(result.stats["locked_users"], 1)
            self.assertEqual(result.groups_backup["default_group"]["dns"], ["10.61.11.1"])
            self.assertEqual(result.groups_backup["groups"][0]["name"], "staff")

            by_name = {user["username"]: user for user in result.users_backup}
            for user in result.users_backup:
                validate_user_record(user)
            self.assertEqual(by_name["alice"]["traffic_type"], "Free")
            self.assertEqual(by_name["alice"]["group"], "defaults")
            self.assertFalse(by_name["alice"]["is_locked"])
            self.assertEqual(by_name["alice"]["created_at"], "2024-12-05T00:00:00Z")
            self.assertEqual(by_name["alice"]["tx"], 501739798)
            self.assertEqual(by_name["bob"]["traffic_type"], "MonthlyTransmit")
            self.assertEqual(by_name["bob"]["traffic_size"], 10 * 1024**3)
            self.assertTrue(by_name["bob"]["is_locked"])
            self.assertEqual(by_name["bob"]["expire_at"], "2026-12-31T00:00:00Z")
            self.assertEqual(by_name["bob"]["group"], "staff")

            written = write_outputs(result, Path(tmp) / "out", gzip_output=True)
            users = json.loads(Path(written["users_json"]).read_text(encoding="utf-8"))
            groups = json.loads(Path(written["groups_json"]).read_text(encoding="utf-8"))
            self.assertEqual(len(users), 2)
            self.assertTrue(groups["default_group"]["no-udp"])
            json.loads(dumps(result.users_backup, compact=True))

    def test_ocpasswd_only_user_gets_generated_password(self):
        with tempfile.TemporaryDirectory() as tmp:
            db_path = Path(tmp) / "empty.sqlite3"
            conn = sqlite3.connect(db_path)
            conn.executescript(SCHEMA)
            conn.execute(
                "INSERT INTO app_ocservgroup VALUES (1, 'defaults', 'defaults group', '{}')"
            )
            conn.execute(
                "INSERT INTO app_adminpanelconfiguration VALUES (1, 10, '{}')"
            )
            conn.commit()
            conn.close()
            ocpasswd_path = Path(tmp) / "ocpasswd"
            ocpasswd_path.write_text("onlyme:*:$5$hash\n", encoding="utf-8")
            result = convert_legacy(db_path, ocpasswd_path)
            self.assertEqual(result.users_backup[0]["username"], "onlyme")
            self.assertIn("onlyme", result.generated_passwords)


class RetryFilterTests(unittest.TestCase):
    def test_parse_restore_errors(self):
        text = "user Maya: exit status 1; user serg: exit status 1; user 2128_new: exit status 1"
        self.assertEqual(parse_failed_usernames(text), ["Maya", "serg", "2128_new"])

    def test_filter_users_keeps_order(self):
        users = [{"username": "a"}, {"username": "b"}, {"username": "c"}]
        filtered, missing = filter_users(users, ["c", "a", "zzz"])
        self.assertEqual([user["username"] for user in filtered], ["c", "a"])
        self.assertEqual(missing, ["zzz"])


class FillOcpasswdTests(unittest.TestCase):
    def test_merge_adds_missing_legacy_users(self):
        with tempfile.TemporaryDirectory() as tmp:
            legacy = Path(tmp) / "legacy"
            current = Path(tmp) / "current"
            legacy.write_text("alice:*:$5$old\nbob:*:!$5$old\n", encoding="utf-8")
            current.write_text("alice:*:$5$new\ncarol:staff:$5$new\n", encoding="utf-8")
            merged, stats = merge_files(legacy, current)
            usernames = [line.split(":", 1)[0] for line in merged]
            self.assertEqual(usernames, ["alice", "carol", "bob"])
            self.assertTrue(merged[0].endswith("$5$new"))
            self.assertTrue(merged[2].startswith("bob:*:!"))
            self.assertEqual(stats["added_from_legacy"], 1)
            self.assertEqual(stats["only_in_current"], 1)


if __name__ == "__main__":
    unittest.main()
