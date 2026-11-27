#!/usr/bin/env python3
"""Regression fixtures use Terraform raw-state v4 shape, never show-JSON input."""
import copy
import importlib.util
import json
from pathlib import Path
import stat
import subprocess
import sys
import tempfile
import unittest

SCRIPT = Path(__file__).with_name("transfer-state.py")
spec = importlib.util.spec_from_file_location("transfer_state", SCRIPT)
transfer = importlib.util.module_from_spec(spec)
spec.loader.exec_module(transfer)


def resource(kind, name, attrs, provider="betterstackhq/logtail", module=None, key=None):
    result = {"mode": "managed", "type": kind, "name": name,
              "provider": f'provider["registry.terraform.io/{provider}"]',
              "instances": [{"schema_version": 0, "attributes": attrs, "sensitive_attributes": [], "private": "b3BhcXVl"}]}
    if module is not None:
        result["module"] = module
    if key is not None:
        result["instances"][0]["index_key"] = key
    return result


def state(serial, resources):
    return {"version": 4, "terraform_version": "1.11.4", "serial": serial,
            "lineage": "5b7afc11-79df-4e4b-a1d4-7a91a29ab0c1", "outputs": {}, "resources": resources}


def webhook_fixture():
    source = {"id": "88", "name": "untouched"}
    target = copy.deepcopy(source)
    source["custom_webhook_template_attributes"] = [{"id": "99", "auth_password": "never-print-this-secret", "auth_username": "fixture"}]
    target["custom_webhook_template_attributes"] = {"auth_password": None, "auth_username": "fixture", "body_template": "unchanged"}
    # The read-back template the provider also stores: it never holds auth_password.
    target["custom_webhook_template"] = {"auth_username": "fixture", "body_template": "unchanged"}
    backup = state(2, [resource("betteruptime_outgoing_webhook", "old", source, "BetterStackHQ/better-uptime")])
    current = state(4, [resource("betterstack_incidents_outgoing_webhook_integration", "new", target, "betterstackhq/betterstack")])
    return backup, current


class TransferStateTest(unittest.TestCase):
    def transfer(self, backup, current, source="betteruptime_outgoing_webhook.old", target="betterstack_incidents_outgoing_webhook_integration.new", child=None):
        return transfer.prepare(backup, current, source, target, current["serial"], child)

    def test_collector_child_password_and_config_transfer_without_changing_parent(self):
        backup, current = self.collector_fixture()
        original = copy.deepcopy(current["resources"][0])
        result = self.transfer(backup, current, "logtail_collector.main", "betterstack_collector_target.database", "9007199254740993")
        attrs = result["resources"][1]["instances"][0]["attributes"]
        self.assertEqual(attrs["password"], "never-print-this-secret")
        self.assertEqual(attrs["host"], "db.internal")
        self.assertEqual(attrs["port"], 5432)
        self.assertEqual(attrs["ssl_mode"], "require")
        self.assertEqual(result["resources"][0], original)

    def test_collector_data_source_is_not_a_managed_parent_or_target_owner(self):
        backup, current = self.collector_fixture()
        parent_lookup = copy.deepcopy(current["resources"][0])
        parent_lookup["mode"] = "data"
        target_lookup = copy.deepcopy(current["resources"][1])
        target_lookup["mode"] = "data"
        current["resources"].extend([parent_lookup, target_lookup])
        result = self.transfer(backup, current, "logtail_collector.main", "betterstack_collector_target.database", "9007199254740993")
        self.assertEqual(result["resources"][1]["instances"][0]["attributes"]["password"], "never-print-this-secret")
        self.assertEqual(result["resources"][2:], current["resources"][2:])

    def test_renamed_legacy_collector_cannot_remain_a_managed_parent_owner(self):
        for mode, parent_id, duplicate in (("managed", 77, True), ("data", "77", False), ("managed", "78", False)):
            with self.subTest(mode=mode, parent_id=parent_id):
                backup, current = self.collector_fixture()
                legacy_parent = copy.deepcopy(backup["resources"][0])
                legacy_parent.update(name="renamed", mode=mode)
                legacy_parent["instances"][0]["attributes"]["id"] = parent_id
                current["resources"].append(legacy_parent)
                if duplicate:
                    with self.assertRaises(transfer.TransferError):
                        self.transfer(backup, current, "logtail_collector.main", "betterstack_collector_target.database", "9007199254740993")
                else:
                    result = self.transfer(backup, current, "logtail_collector.main", "betterstack_collector_target.database", "9007199254740993")
                    self.assertEqual(result["resources"][1]["instances"][0]["attributes"]["password"], "never-print-this-secret")
                    self.assertEqual(result["resources"][2], legacy_parent)

    def test_source_connections_have_no_handoff(self):
        # A logtail_source_aws_account or logtail_source_gcp_project moves to its own resource with
        # its state, write-only values included: there is nothing to hand off.
        for legacy, target in (("logtail_source_aws_account", "betterstack_source_aws_account"), ("logtail_source_gcp_project", "betterstack_source_gcp_project")):
            with self.subTest(legacy=legacy):
                backup = state(1, [resource(legacy, "link", {"id": "42", "source_id": "42"})])
                current = state(2, [resource(target, "link", {"source_id": "42"}, "betterstackhq/betterstack")])
                with self.assertRaises(transfer.TransferError):
                    self.transfer(backup, current, legacy+".link", target+".link")

    def test_legacy_collector_target_ownership_normalizes_parent_and_child(self):
        for legacy_id, parent_id, duplicate in (
            ("9007199254740993", "77", True),
            ("77:9007199254740993", "77", True),
            ("77:9007199254740993", None, True),
            ("78:9007199254740993", "78", False),
        ):
            with self.subTest(legacy_id=legacy_id, parent_id=parent_id):
                backup, current = self.collector_fixture()
                attrs = {"id": legacy_id}
                if parent_id is not None:
                    attrs["collector_id"] = parent_id
                current["resources"].append(resource("logtail_collector_target", "old", attrs))
                if duplicate:
                    with self.assertRaises(transfer.TransferError):
                        self.transfer(backup, current, "logtail_collector.main", "betterstack_collector_target.database", "9007199254740993")
                else:
                    result = self.transfer(backup, current, "logtail_collector.main", "betterstack_collector_target.database", "9007199254740993")
                    self.assertEqual(result["resources"][1]["instances"][0]["attributes"]["password"], "never-print-this-secret")

    def test_unrelated_provider_identity_does_not_create_false_ownership_collision(self):
        backup, current = webhook_fixture()
        current["resources"].append(resource("betteruptime_outgoing_webhook", "foreign", {"id": "88"}, provider="other/better-uptime"))
        result = self.transfer(backup, current)
        self.assertEqual(result["resources"][0]["instances"][0]["attributes"]["custom_webhook_template_attributes"]["auth_password"], "never-print-this-secret")

    # Legacy type -> new type. The two providers do not share every name, so neither is derived
    # from the other.
    UPTIME_HANDOFFS = {
        "betteruptime_outgoing_webhook": "betterstack_incidents_outgoing_webhook_integration",
    }

    def uptime_fixture(self, legacy):
        return webhook_fixture()

    def test_webhook_password_transfer_preserves_template_and_marks_nested_sensitivity(self):
        backup, current = self.uptime_fixture("betteruptime_outgoing_webhook")
        original = copy.deepcopy(current)
        result = self.transfer(backup, current, "betteruptime_outgoing_webhook.old", "betterstack_incidents_outgoing_webhook_integration.new")
        expected = copy.deepcopy(current)
        instance = expected["resources"][0]["instances"][0]
        instance["attributes"]["custom_webhook_template_attributes"]["auth_password"] = "never-print-this-secret"
        instance["sensitive_attributes"] = [[{"type": "get_attr", "value": "custom_webhook_template_attributes"}, {"type": "get_attr", "value": "auth_password"}]]
        expected["serial"] += 1
        self.assertEqual(result, expected)
        self.assertEqual(current, original)

    def test_status_page_item_has_no_handoff(self):
        # A legacy status page item folds into its page's sections list: no state is handed off.
        backup, current = self.uptime_fixture("betteruptime_outgoing_webhook")
        backup["resources"][0]["type"] = "betteruptime_status_page_resource"
        with self.assertRaises(transfer.TransferError):
            self.transfer(backup, current, "betteruptime_status_page_resource.old", "betterstack_incidents_outgoing_webhook_integration.new")

    def test_uptime_handoffs_refuse_wrong_identity_provider_duplicate_owner_and_conflicts(self):
        for legacy, current_type in self.UPTIME_HANDOFFS.items():
            for change in ("identity", "provider", "schema", "duplicate legacy", "duplicate target", "conflict", "empty conflict", "missing destination"):
                with self.subTest(legacy=legacy, change=change):
                    backup, current = self.uptime_fixture(legacy)
                    source = backup["resources"][0]
                    target = current["resources"][0]["instances"][0]["attributes"]
                    if change == "identity":
                        target["id"] = "89"
                    elif change == "provider":
                        source["provider"] = 'provider["registry.terraform.io/other/better-uptime"]'
                    elif change == "schema":
                        source["instances"][0]["schema_version"] = 1
                    elif change == "duplicate legacy":
                        current["resources"].append(dict(copy.deepcopy(source), name="renamed"))
                    elif change == "duplicate target":
                        current["resources"].append(dict(copy.deepcopy(current["resources"][0]), name="duplicate"))
                    elif change in ("conflict", "empty conflict"):
                        target["custom_webhook_template_attributes"]["auth_password"] = "other-secret" if change == "conflict" else ""
                    else:
                        target["custom_webhook_template_attributes"] = None
                    with self.assertRaises(transfer.TransferError):
                        self.transfer(backup, current, legacy+".old", current_type+".new")
        backup, current = self.uptime_fixture("betteruptime_outgoing_webhook")
        backup["resources"][0]["instances"][0]["attributes"]["custom_webhook_template_attributes"][0]["auth_password"] = None
        with self.assertRaises(transfer.TransferError):
            self.transfer(backup, current, "betteruptime_outgoing_webhook.old", "betterstack_incidents_outgoing_webhook_integration.new")

    def collector_fixture(self):
        child = {"id": 9007199254740993, "service_type": "postgres", "host": "db.internal", "port": 5432, "username": "fixture", "password": "never-print-this-secret", "ssl_mode": "require", "tls": None}
        backup = state(2, [resource("logtail_collector", "main", {"id": "77", "secret": "enrollment-secret", "databases": [child]})])
        current = state(4, [resource("betterstack_collector", "main", {"id": "77", "secret": "enrollment-secret", "databases": [{"id": "9007199254740993", "kind": "postgres", "password": "never-print-this-secret"}]}, "betterstackhq/betterstack"), resource("betterstack_collector_target", "database", {"id": "9007199254740993", "collector_id": "77", "kind": "postgres", "host": "db.internal", "port": 5432, "username": "fixture", "password": None, "ssl_mode": "require", "tls": None, "enabled": True, "status": "connected"}, "betterstackhq/betterstack")])
        return backup, current

    def test_rejects_unsafe_source_or_current_state(self):
        cases = {
            "lineage": lambda b, c: c.update(lineage="different"),
            "future backup": lambda b, c: b.update(serial=99),
            "schema": lambda b, c: c["resources"][0]["instances"][0].update(schema_version=1),
            "source schema": lambda b, c: b["resources"][0]["instances"][0].update(schema_version=1),
            "provider lookalike": lambda b, c: b["resources"][0].update(provider='provider["registry.terraform.io/betterstackhq/better-uptime-extra"]'),
            "target provider": lambda b, c: c["resources"][0].update(provider='provider["registry.terraform.io/other/betterstack"]'),
            "wrong identity": lambda b, c: c["resources"][0]["instances"][0]["attributes"].update(id="12"),
            "missing secret": lambda b, c: b["resources"][0]["instances"][0]["attributes"]["custom_webhook_template_attributes"][0].update(auth_password=None),
            "conflict": lambda b, c: c["resources"][0]["instances"][0]["attributes"]["custom_webhook_template_attributes"].update(auth_password="existing-other-secret"),
            "duplicate instance": lambda b, c: c["resources"][0]["instances"].append(copy.deepcopy(c["resources"][0]["instances"][0])),
            "duplicate remote owner": lambda b, c: c["resources"].append(dict(copy.deepcopy(c["resources"][0]), name="duplicate")),
            "tainted target": lambda b, c: c["resources"][0]["instances"][0].update(status="tainted"),
            "deposed source": lambda b, c: b["resources"][0]["instances"][0].update(deposed="deadbeef"),
            "legacy owner retained": lambda b, c: c["resources"].append(copy.deepcopy(b["resources"][0])),
        }
        for name, change in cases.items():
            with self.subTest(name=name):
                backup, current = webhook_fixture()
                change(backup, current)
                with self.assertRaises(transfer.TransferError):
                    self.transfer(backup, current)
        backup, current = webhook_fixture()
        with self.assertRaises(transfer.TransferError):
            transfer.prepare(backup, current, "betteruptime_outgoing_webhook.old", "betterstack_incidents_outgoing_webhook_integration.new", 3, None)

    def test_collector_refuses_wrong_parent_duplicate_child_and_missing_password(self):
        for mutation in (
            lambda b, c: c["resources"][1]["instances"][0]["attributes"].update(collector_id="78"),
            lambda b, c: b["resources"][0]["instances"][0]["attributes"]["databases"].append(copy.deepcopy(b["resources"][0]["instances"][0]["attributes"]["databases"][0])),
            lambda b, c: b["resources"][0]["instances"][0]["attributes"]["databases"][0].update(password=None),
            lambda b, c: c["resources"].pop(0),
        ):
            backup, current = self.collector_fixture()
            mutation(backup, current)
            with self.assertRaises(transfer.TransferError):
                self.transfer(backup, current, "logtail_collector.main", "betterstack_collector_target.database", "9007199254740993")

    def test_cli_exclusive_private_output_and_secret_free_failures(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            backup, current = webhook_fixture()
            old, now, output = root/"backup.json", root/"current.json", root/"candidate.json"
            old.write_text(json.dumps(backup)); now.write_text(json.dumps(current))
            original = (old.read_bytes(), now.read_bytes())
            args = [sys.executable, str(SCRIPT), "--backup", str(old), "--current", str(now), "--source", "betteruptime_outgoing_webhook.old", "--target", "betterstack_incidents_outgoing_webhook_integration.new", "--expected-serial", "4", "--output", str(output)]
            first = subprocess.run(args, capture_output=True, text=True)
            self.assertEqual(first.returncode, 0, first.stderr)
            self.assertEqual(stat.S_IMODE(output.stat().st_mode), 0o600)
            self.assertNotIn("never-print-this-secret", first.stdout+first.stderr)
            written = output.read_bytes()
            second = subprocess.run(args, capture_output=True, text=True)
            self.assertNotEqual(second.returncode, 0)
            self.assertEqual(output.read_bytes(), written)
            self.assertEqual((old.read_bytes(), now.read_bytes()), original)
            self.assertNotIn("never-print-this-secret", second.stdout+second.stderr)
            self.assertEqual(sorted(path.name for path in root.iterdir()), ["backup.json", "candidate.json", "current.json"])

    def test_json_preserves_unrelated_exact_numbers_and_rejects_duplicate_keys(self):
        doc = transfer.decode('{"n":900719925474099312345,"f":0.12345678901234567890123456789}')
        self.assertEqual(transfer.decode(transfer.encode(doc)), doc)
        with self.assertRaises(transfer.TransferError):
            transfer.decode('{"serial":1,"serial":2}')
        with self.assertRaises(transfer.TransferError):
            transfer.decode('{"x":NaN}')


if __name__ == "__main__":
    unittest.main()
